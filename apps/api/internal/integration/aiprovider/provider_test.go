package aiprovider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/integration/anthropic"
	"github.com/kristopher1027/devflow-ai/internal/integration/gemini"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestAutoDetectsFromWhichKeyIsSet(t *testing.T) {
	generator, name := NewTextGenerator(env(map[string]string{"GEMINI_API_KEY": "g"}))
	require.IsType(t, &gemini.Client{}, generator)
	require.Equal(t, "gemini", name)

	generator, name = NewTextGenerator(env(map[string]string{"ANTHROPIC_API_KEY": "a"}))
	require.IsType(t, &anthropic.Client{}, generator)
	require.Equal(t, "anthropic", name)
}

func TestAnthropicWinsWhenBothKeysAreSetUnlessOverridden(t *testing.T) {
	both := map[string]string{"ANTHROPIC_API_KEY": "a", "GEMINI_API_KEY": "g"}

	generator, name := NewTextGenerator(env(both))
	require.IsType(t, &anthropic.Client{}, generator)
	require.Equal(t, "anthropic", name)

	both["AI_PROVIDER"] = " Gemini "
	generator, name = NewTextGenerator(env(both))
	require.IsType(t, &gemini.Client{}, generator)
	require.Equal(t, "gemini", name)
}

func TestWithoutAnyKeyAIRequestsReportNotConfigured(t *testing.T) {
	generator, name := NewTextGenerator(env(nil))

	_, err := generator.Generate(context.Background(), "s", "p")

	require.ErrorIs(t, err, anthropic.ErrAPIKeyRequired)
	require.Contains(t, name, "none")
}

func TestChosenProviderWithoutItsKeyIsReportedNotConfigured(t *testing.T) {
	generator, name := NewTextGenerator(env(map[string]string{
		"AI_PROVIDER":       "gemini",
		"ANTHROPIC_API_KEY": "a",
	}))

	_, err := generator.Generate(context.Background(), "s", "p")

	require.ErrorIs(t, err, anthropic.ErrAPIKeyRequired)
	require.Contains(t, name, "gemini (disabled")
}

func TestUnknownProviderIsReportedNotConfigured(t *testing.T) {
	generator, name := NewTextGenerator(env(map[string]string{"AI_PROVIDER": "openai"}))

	_, err := generator.Generate(context.Background(), "s", "p")

	require.ErrorIs(t, err, anthropic.ErrAPIKeyRequired)
	require.Contains(t, err.Error(), `unknown AI_PROVIDER "openai"`)
	require.Contains(t, name, "none")
}
