package gemini

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

// GeminiAPIURL is the correct API endpoint for Gemini.
var GeminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"

// Client handles requests to the Gemini API.
type Client struct {
	URL        *url.URL
	HTTPClient *http.Client
	APIKey     string
}

// Payload is the request body for the Gemini API.
// Note the use of "messages" to match the API specification.
type Payload struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response is the API response from Gemini.
type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// Choice represents an answer choice in the response.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// NewClient creates a new Gemini client.
func NewClient(urlStr, apiKey string) (*Client, error) {
	if urlStr == "" {
		return nil, fmt.Errorf("gemini client: missing url")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("gemini client: missing api key")
	}
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %s: %w", urlStr, err)
	}
	return &Client{
		URL:        parsedURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		APIKey:     apiKey,
	}, nil
}

// newRequest creates a new HTTP request.
func (c *Client) newRequest(ctx context.Context, method string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.URL.String(), body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return req, nil
}

// Chat sends a request to the Gemini API and returns the response.
func (c *Client) Chat(ctx context.Context, param *Payload) (*Response, error) {
	if len(param.Messages) == 0 || param.Messages[0].Content == "" {
		return nil, fmt.Errorf("missing message content")
	}

	b, err := json.Marshal(param)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

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
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("read body error: %w", err)
		}
		return nil, fmt.Errorf("status code: %d; body: %s", res.StatusCode, bodyBytes)
	}

	response := &Response{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return response, nil
}
