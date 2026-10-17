// Package aiprovider picks which AI service answers questions, based on
// environment variables, so the rest of the app does not care which one.
package aiprovider

import (
	"fmt"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/integration/anthropic"
	"github.com/kristopher1027/devflow-ai/internal/integration/gemini"
)

// NewTextGenerator chooses a provider and returns it with a short
// description for the start-up log.
//
//	AI_PROVIDER      "anthropic" or "gemini" (optional)
//	ANTHROPIC_API_KEY, ANTHROPIC_MODEL
//	GEMINI_API_KEY,    GEMINI_MODEL
//
// Without AI_PROVIDER, Anthropic is used if its key is set, otherwise Gemini
// if its key is set. With no key at all the server still starts, and AI
// requests answer "not configured".
func NewTextGenerator(getenv func(string) string) (anthropic.TextGenerator, string) {
	provider := strings.ToLower(strings.TrimSpace(getenv("AI_PROVIDER")))

	if provider == "" {
		switch {
		case strings.TrimSpace(getenv("ANTHROPIC_API_KEY")) != "":
			provider = "anthropic"
		case strings.TrimSpace(getenv("GEMINI_API_KEY")) != "":
			provider = "gemini"
		default:
			return unavailable(anthropic.ErrAPIKeyRequired),
				"none (set GEMINI_API_KEY or ANTHROPIC_API_KEY to enable AI answers)"
		}
	}

	switch provider {
	case "anthropic":
		client, err := anthropic.NewClient(
			getenv("ANTHROPIC_API_KEY"),
			getenv("ANTHROPIC_MODEL"),
		)
		if err != nil {
			return unavailable(err), "anthropic (disabled: " + err.Error() + ")"
		}
		return client, "anthropic"

	case "gemini":
		client, err := gemini.NewClient(
			getenv("GEMINI_API_KEY"),
			getenv("GEMINI_MODEL"),
		)
		if err != nil {
			return unavailable(err), "gemini (disabled: " + err.Error() + ")"
		}
		return client, "gemini"

	default:
		err := fmt.Errorf(
			"%w: unknown AI_PROVIDER %q (use anthropic or gemini)",
			anthropic.ErrAPIKeyRequired,
			provider,
		)
		return unavailable(err), "none (" + err.Error() + ")"
	}
}

func unavailable(err error) anthropic.TextGenerator {
	return anthropic.NewUnavailableGenerator(err)
}
