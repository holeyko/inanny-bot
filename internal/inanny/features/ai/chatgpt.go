package ai

const (
	chatGPTEndpoint = "https://api.openai.com/v1/chat/completions"
	chatGPTModel    = "gpt-4o-mini"
)

func NewChatGPTClient(token string) AIClient {
	return newOpenAICompatibleClient("ChatGPT", token, chatGPTEndpoint, chatGPTModel)
}
