package mirait

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/catatsuy/bento/config"
)

const (
	DefaultAPITimeout = 10
)

var (
	targetURL = "https://miraitranslate.com"
)

type Session struct {
	URL        *url.URL
	HTTPClient *http.Client
	Token      string

	config config.Config
}

func NewSession(cfg config.Config) (*Session, error) {
	parsedURL, err := url.ParseRequestURI(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %s: %w", targetURL, err)
	}

	jar, _ := cookiejar.New(&cookiejar.Options{})

	session := &Session{
		URL: &url.URL{
			Scheme: parsedURL.Scheme,
			Host:   parsedURL.Host,
		},
		HTTPClient: &http.Client{
			Jar:     jar,
			Timeout: time.Duration(DefaultAPITimeout) * time.Second,
		},
		config: cfg,
	}

	return session, nil
}

func (s *Session) SetCacheCookie(ccs []config.Cookie) {
	cookies := make([]*http.Cookie, 0, len(ccs))
	for _, cc := range ccs {
		cookies = append(cookies, &http.Cookie{
			Name:  cc.Name,
			Value: cc.Value,
		})
	}
	s.HTTPClient.Jar.SetCookies(s.URL, cookies)
}

func (s *Session) GetToken() (string, error) {
	u := s.URL
	u.Path = "/trial"

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", s.config.UserAgent)

	res, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	bb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`var tran = "([a-zA-Z0-9]*)"`)
	tokenByte := re.FindAllSubmatch(bb, 1)

	if len(tokenByte) == 0 {
		return "", errors.New("empty token")
	}

	return string(tokenByte[0][1]), nil
}

func (s *Session) SetToken(token string) {
	s.Token = token
}

func (s *Session) Refresh() error {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	s.HTTPClient.Jar = jar

	token, err := s.GetToken()
	if err != nil {
		return err
	}
	s.SetToken(token)
	return nil
}

type outputRes struct {
	Output string `json:"output"`
}

// {"status":"success","outputs":[{"output":"こんにちは。"}]}
type PostTranslateRes struct {
	Status  string      `json:"status"`
	Outputs []outputRes `json:"outputs"`
}

func (s *Session) PostTranslate(input string, isJP bool) (output string, err error) {
	u := s.URL
	u.Path = "/trial/translate.php"

	q := url.Values{}
	q.Set("input", input)
	q.Set("tran", s.Token)

	if isJP {
		q.Set("source", "ja")
		q.Set("target", "en")
	} else {
		q.Set("source", "en")
		q.Set("target", "ja")
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(q.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", s.config.UserAgent)

	res, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	ptr := &PostTranslateRes{}
	err = json.NewDecoder(res.Body).Decode(ptr)
	if err != nil {
		return "", fmt.Errorf("failed to encode json: %w", err)
	}

	if len(ptr.Outputs) == 0 {
		return "", fmt.Errorf("empty response")
	}

	if ptr.Status != "success" {
		return "", fmt.Errorf("no success response: %s", ptr.Status)
	}

	return ptr.Outputs[0].Output, nil
}

func (s *Session) DumpCookies() []config.Cookie {
	cookies := s.HTTPClient.Jar.Cookies(s.URL)
	ccs := make([]config.Cookie, 0, len(cookies))

	for _, c := range cookies {
		cc := config.Cookie{
			Name:  c.Name,
			Value: c.Value,
		}
		ccs = append(ccs, cc)
	}

	return ccs
}
