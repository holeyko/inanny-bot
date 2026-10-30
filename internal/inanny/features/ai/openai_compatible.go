package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type openAICompatibleClient struct {
	provider   string
	token      string
	endpoint   string
	model      string
	httpClient *http.Client
}

func newOpenAICompatibleClient(provider, token, endpoint, model string) AIClient {
	return &openAICompatibleClient{
		provider:   provider,
		token:      token,
		endpoint:   endpoint,
		model:      model,
		httpClient: http.DefaultClient,
	}
}

func (client *openAICompatibleClient) Complete(ctx context.Context, request CompletionRequest) (text string, resultErr error) {
	started := time.Now()
	var completion chatCompletionResponse
	defer func() {
		logCompletion(client.provider, client.model, started, completion.Usage.PromptTokens,
			completion.Usage.CompletionTokens, completion.Usage.TotalTokens, resultErr)
	}()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if client.token == "" {
		return "", ErrNotConfigured
	}

	payload := chatCompletionRequest{
		Model:     client.model,
		MaxTokens: outputTokenBudget(request.MaxOutputTokens),
		Messages: []chatCompletionMessage{
			{Role: "system", Content: request.SystemPrompt},
			{Role: "user", Content: request.UserPrompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal %s request: %w", client.provider, err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, client.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create %s request: %w", client.provider, err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+client.token)
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return "", fmt.Errorf("call %s API: %w", client.provider, err)
	}
	defer response.Body.Close()

	responseBody, err := readResponseBody(response.Body)
	if err != nil {
		return "", fmt.Errorf("read %s response: %w", client.provider, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%s API returned status %d", client.provider, response.StatusCode)
	}

	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return "", fmt.Errorf("decode %s response: invalid JSON", client.provider)
	}
	if len(completion.Error) > 0 && string(completion.Error) != "null" {
		return "", errors.New(client.provider + " returned an API error")
	}
	if len(completion.Choices) == 0 || strings.TrimSpace(completion.Choices[0].Message.Content) == "" {
		return "", errors.New(client.provider + " returned an empty completion")
	}
	choice := completion.Choices[0]
	if choice.Message.Refusal != "" || (choice.FinishReason != "" && choice.FinishReason != "stop") {
		return "", errors.New(client.provider + " returned a blocked or incomplete completion")
	}

	return completion.Choices[0].Message.Content, nil
}

type chatCompletionRequest struct {
	Model     string                  `json:"model"`
	MaxTokens int                     `json:"max_tokens"`
	Messages  []chatCompletionMessage `json:"messages"`
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Refusal string `json:"refusal,omitempty"`
}

type chatCompletionResponse struct {
	Error   json.RawMessage `json:"error"`
	Choices []struct {
		Message      chatCompletionMessage `json:"message"`
		FinishReason string                `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
