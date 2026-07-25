package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func (client *openAICompatibleClient) Complete(ctx context.Context, request CompletionRequest) (string, error) {
	if client.token == "" {
		return "", ErrNotConfigured
	}

	payload := chatCompletionRequest{
		Model: client.model,
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

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read %s response: %w", client.provider, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%s API returned status %d: %s", client.provider, response.StatusCode, string(responseBody))
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return "", fmt.Errorf("decode %s response: %w", client.provider, err)
	}
	if len(completion.Choices) == 0 || completion.Choices[0].Message.Content == "" {
		return "", errors.New(client.provider + " returned an empty completion")
	}

	return completion.Choices[0].Message.Content, nil
}

type chatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []chatCompletionMessage `json:"messages"`
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatCompletionMessage `json:"message"`
	} `json:"choices"`
}
