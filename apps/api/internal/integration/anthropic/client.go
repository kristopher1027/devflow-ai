package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://api.anthropic.com"
	apiVersion       = "2023-06-01"
	defaultMaxTokens = 2000
	maxResponseBytes = 1 << 20

	// DefaultModel is used when no model is configured.
	DefaultModel = "claude-sonnet-5-5"
)

var (
	ErrAPIKeyRequired   = errors.New("anthropic API key is required")
	ErrUnexpectedStatus = errors.New("anthropic API returned unexpected status")
	ErrEmptyResponse    = errors.New("anthropic API returned no text")
)

// TextGenerator is what the rest of the app depends on, so tests can use
// a fake instead of calling the real API.
type TextGenerator interface {
	Generate(
		ctx context.Context,
		system string,
		prompt string,
	) (string, error)
}

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string, model string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrAPIKeyRequired
	}
	if strings.TrimSpace(model) == "" {
		model = DefaultModel
	}

	return &Client{
		apiKey:     apiKey,
		model:      model,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}, nil
}

func (c *Client) Generate(
	ctx context.Context,
	system string,
	prompt string,
) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":      c.model,
		"max_tokens": defaultMaxTokens,
		"system":     system,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("encode anthropic request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/v1/messages",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("create anthropic request: %w", err)
	}
	request.Header.Set("content-type", "application/json")
	request.Header.Set("x-api-key", c.apiKey)
	request.Header.Set("anthropic-version", apiVersion)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("call anthropic API: %w", err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("read anthropic response: %w", err)
	}

	// The status is reported without the body or any request headers so the
	// API key can never end up in logs or error messages.
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d", ErrUnexpectedStatus, response.StatusCode)
	}

	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("decode anthropic response: %w", err)
	}

	var text strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}

	result := strings.TrimSpace(text.String())
	if result == "" {
		return "", ErrEmptyResponse
	}

	return result, nil
}

var _ TextGenerator = (*Client)(nil)
