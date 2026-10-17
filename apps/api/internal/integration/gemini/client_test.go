package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/integration/anthropic"
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

func TestNewClientRequiresAPIKeyAndDefaultsModel(t *testing.T) {
	_, err := NewClient("  ", "")
	require.ErrorIs(t, err, anthropic.ErrAPIKeyRequired)

	client, err := NewClient("key", "")
	require.NoError(t, err)
	require.Equal(t, DefaultModel, client.model)

	client, err = NewClient("key", " custom-model ")
	require.NoError(t, err)
	require.Equal(t, "custom-model", client.model)
}

func TestGenerateSendsRequestAndJoinsAnswerParts(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1beta/models/"+DefaultModel+":generateContent", r.URL.Path)
		require.Equal(t, "test-key", r.Header.Get("x-goog-api-key"))
		require.Empty(t, r.URL.RawQuery, "the key must travel in a header, not the URL")

		var body struct {
			Contents []struct {
				Role  string `json:"role"`
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"contents"`
			SystemInstruction struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"systemInstruction"`
			GenerationConfig struct {
				MaxOutputTokens int `json:"maxOutputTokens"`
			} `json:"generationConfig"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Len(t, body.Contents, 1)
		require.Equal(t, "user", body.Contents[0].Role)
		require.Equal(t, "explain this", body.Contents[0].Parts[0].Text)
		require.Equal(t, "be brief", body.SystemInstruction.Parts[0].Text)
		require.Equal(t, defaultMaxOutputTokens, body.GenerationConfig.MaxOutputTokens)

		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[
			{"text":"hidden reasoning","thought":true},
			{"text":"Hello "},
			{"text":"world\n"}
		]}}]}`))
	})

	got, err := client.Generate(context.Background(), "be brief", "explain this")

	require.NoError(t, err)
	require.Equal(t, "Hello world", got)
}

func TestGenerateOmitsEmptySystemInstruction(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.NotContains(t, body, "systemInstruction")

		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"ok"}]}}]}`))
	})

	got, err := client.Generate(context.Background(), "   ", "hi")

	require.NoError(t, err)
	require.Equal(t, "ok", got)
}

func TestGenerateReportsStatusWithoutLeakingKey(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusTooManyRequests, http.StatusInternalServerError} {
		client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":{"message":"API key not valid: test-key"}}`))
		})

		_, err := client.Generate(context.Background(), "s", "p")

		require.ErrorIs(t, err, anthropic.ErrUnexpectedStatus)
		require.NotContains(t, err.Error(), "test-key")
	}
}

func TestGenerateRejectsEmptyOrBlockedAnswers(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"candidates":[]}`))
	})
	_, err := client.Generate(context.Background(), "s", "p")
	require.ErrorIs(t, err, anthropic.ErrEmptyResponse)

	client = newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"promptFeedback":{"blockReason":"SAFETY"}}`))
	})
	_, err = client.Generate(context.Background(), "s", "p")
	require.ErrorIs(t, err, anthropic.ErrEmptyResponse)
	require.Contains(t, err.Error(), "SAFETY")
}

func TestGenerateRejectsInvalidJSON(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	})

	_, err := client.Generate(context.Background(), "s", "p")

	require.Error(t, err)
}
