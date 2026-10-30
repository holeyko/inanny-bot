package ai

import "strings"

const (
	openRouterEndpoint = "https://openrouter.ai/api/v1/chat/completions"
	openRouterModel    = "~deepseek/deepseek-flash-latest"
)

// NewOpenRouterClient uses the supplied credentials and optional model override.
// Environment configuration is intentionally left to the caller.
func NewOpenRouterClient(apiKey, model string) AIClient {
	if strings.TrimSpace(model) == "" {
		model = openRouterModel
	}
	return newOpenAICompatibleClient("OpenRouter", apiKey, openRouterEndpoint, model)
}
