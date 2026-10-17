package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	// The shared TextGenerator interface and error values live here so the
	// HTTP layer treats every provider the same way.
	"github.com/kristopher1027/devflow-ai/internal/integration/anthropic"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com"

	// DefaultModel can be overridden with the GEMINI_MODEL environment
	// variable if Google retires or renames it.
	DefaultModel = "gemini-2.5-flash"

	// Generous, because some models spend part of this budget on thinking.
	defaultMaxOutputTokens = 4096
	maxResponseBytes       = 2 << 20
)

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string, model string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, anthropic.ErrAPIKeyRequired
	}
	if strings.TrimSpace(model) == "" {
		model = DefaultModel
	}

	return &Client{
		apiKey:     apiKey,
		model:      strings.TrimSpace(model),
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}, nil
}

func (c *Client) Generate(
	ctx context.Context,
	system string,
	prompt string,
) (string, error) {
	body := map[string]any{
		"contents": []map[string]any{
			{
				"role":  "user",
				"parts": []map[string]string{{"text": prompt}},
			},
		},
		"generationConfig": map[string]any{
			"maxOutputTokens": defaultMaxOutputTokens,
		},
	}
	if strings.TrimSpace(system) != "" {
		body["systemInstruction"] = map[string]any{
			"parts": []map[string]string{{"text": system}},
		}
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("encode gemini request: %w", err)
	}

	endpoint := fmt.Sprintf(
		"%s/v1beta/models/%s:generateContent",
		c.baseURL,
		url.PathEscape(c.model),
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return "", fmt.Errorf("create gemini request: %w", err)
	}
	request.Header.Set("content-type", "application/json")
	request.Header.Set("x-goog-api-key", c.apiKey)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("call gemini API: %w", err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("read gemini response: %w", err)
	}

	// Report only the status: never the body or headers, so the API key
	// cannot end up in logs or error messages.
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"%w: %d",
			anthropic.ErrUnexpectedStatus,
			response.StatusCode,
		)
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text    string `json:"text"`
					Thought bool   `json:"thought"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason string `json:"blockReason"`
		} `json:"promptFeedback"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("decode gemini response: %w", err)
	}

	var text strings.Builder
	if len(parsed.Candidates) > 0 {
		for _, part := range parsed.Candidates[0].Content.Parts {
			if !part.Thought {
				text.WriteString(part.Text)
			}
		}
	}

	result := strings.TrimSpace(text.String())
	if result == "" {
		if reason := parsed.PromptFeedback.BlockReason; reason != "" {
			return "", fmt.Errorf("%w (blocked: %s)", anthropic.ErrEmptyResponse, reason)
		}
		return "", anthropic.ErrEmptyResponse
	}

	return result, nil
}

var _ anthropic.TextGenerator = (*Client)(nil)
