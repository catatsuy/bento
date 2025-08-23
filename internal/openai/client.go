package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	OpenAIAPIURL = "https://api.openai.com/v1/responses"
)

type Client struct {
	URL        *url.URL
	HTTPClient *http.Client
	APIKey     string
}

type Payload struct {
	Model        string `json:"model"`
	Input        string `json:"input,omitempty"` // Can also be an array of Message objects
	Instructions string `json:"instructions,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	ID     string          `json:"id"`
	Output []OutputMessage `json:"output"`
	Usage  Usage           `json:"usage"`
}

type OutputMessage struct {
	ID      string    `json:"id"`
	Type    string    `json:"type"`
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type        string   `json:"type"`
	Text        string   `json:"text"`
	Annotations []string `json:"annotations"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OutputText returns the first text content from the response
func (r *Response) OutputText() string {
	for _, output := range r.Output {
		if output.Type == "message" {
			for _, content := range output.Content {
				if content.Type == "text" || content.Type == "output_text" {
					return content.Text
				}
			}
		}
	}
	return ""
}

func NewClient(urlStr, apiKey string) (*Client, error) {
	if len(urlStr) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	if len(apiKey) == 0 {
		return nil, fmt.Errorf("client: missing api key")
	}

	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %s: %w", urlStr, err)
	}

	client := &Client{
		URL:        parsedURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		APIKey:     apiKey,
	}

	return client, nil
}

func (c *Client) newRequest(ctx context.Context, method string, body io.Reader) (*http.Request, error) {
	u := *c.URL

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	return req, nil
}

func (c *Client) Chat(ctx context.Context, param *Payload) (*Response, error) {
	if param.Input == "" {
		return nil, nil
	}

	b, _ := json.Marshal(param)

	req, err := c.newRequest(ctx, http.MethodPost, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read res.Body and the status code of the response from slack was not 200: %w", err)
		}
		return nil, fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	response := &Response{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return response, nil
}
