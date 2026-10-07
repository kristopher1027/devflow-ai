package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient("test-key", "")
	require.NoError(t, err)
	client.baseURL = server.URL

	return client
}

func TestNewClientRequiresAPIKey(t *testing.T) {
	_, err := NewClient("   ", "")

	require.ErrorIs(t, err, ErrAPIKeyRequired)
}

func TestNewClientDefaultsModel(t *testing.T) {
	client, err := NewClient("key", "")

	require.NoError(t, err)
	require.Equal(t, DefaultModel, client.model)
}

func TestGenerateSendsRequestAndJoinsTextBlocks(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/messages", r.URL.Path)
		require.Equal(t, "test-key", r.Header.Get("x-api-key"))
		require.Equal(t, apiVersion, r.Header.Get("anthropic-version"))
		require.Equal(t, "application/json", r.Header.Get("content-type"))

		var body struct {
			Model     string `json:"model"`
			MaxTokens int    `json:"max_tokens"`
			System    string `json:"system"`
			Messages  []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, DefaultModel, body.Model)
		require.Equal(t, defaultMaxTokens, body.MaxTokens)
		require.Equal(t, "be brief", body.System)
		require.Len(t, body.Messages, 1)
		require.Equal(t, "user", body.Messages[0].Role)
		require.Equal(t, "explain this", body.Messages[0].Content)

		_, _ = w.Write([]byte(`{"content":[
			{"type":"text","text":"Hello "},
			{"type":"tool_use","id":"x"},
			{"type":"text","text":"world\n"}
		]}`))
	})

	got, err := client.Generate(context.Background(), "be brief", "explain this")

	require.NoError(t, err)
	require.Equal(t, "Hello world", got)
}

func TestGenerateReturnsStatusErrorWithoutLeakingKey(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key test-key"}`))
	})

	_, err := client.Generate(context.Background(), "s", "p")

	require.ErrorIs(t, err, ErrUnexpectedStatus)
	require.Contains(t, err.Error(), "401")
	require.NotContains(t, err.Error(), "test-key")
}

func TestGenerateRejectsEmptyContent(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"content":[]}`))
	})

	_, err := client.Generate(context.Background(), "s", "p")

	require.True(t, errors.Is(err, ErrEmptyResponse))
}

func TestGenerateRejectsInvalidJSON(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	})

	_, err := client.Generate(context.Background(), "s", "p")

	require.Error(t, err)
}
