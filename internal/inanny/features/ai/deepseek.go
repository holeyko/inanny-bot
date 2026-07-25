package ai

const (
	deepSeekEndpoint = "https://api.deepseek.com/chat/completions"
	deepSeekModel    = "deepseek-chat"
)

func NewDeepSeekClient(token string) AIClient {
	return newOpenAICompatibleClient("DeepSeek", token, deepSeekEndpoint, deepSeekModel)
}
