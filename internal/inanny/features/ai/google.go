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

const googleAIStudioEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.1-flash-lite:generateContent"

type GoogleAIClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewGoogleAIClient(apiKey string) AIClient {
	return &GoogleAIClient{
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

func (client *GoogleAIClient) Complete(ctx context.Context, request CompletionRequest) (text string, resultErr error) {
	started := time.Now()
	var completion googleGenerateContentResponse
	defer func() {
		logCompletion("Google AI Studio", "gemini-3.1-flash-lite", started,
			completion.UsageMetadata.PromptTokenCount, completion.UsageMetadata.CandidatesTokenCount,
			completion.UsageMetadata.TotalTokenCount, resultErr)
	}()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if client.apiKey == "" {
		return "", ErrNotConfigured
	}

	payload := googleGenerateContentRequest{
		GenerationConfig: googleGenerationConfig{MaxOutputTokens: outputTokenBudget(request.MaxOutputTokens)},
		SystemInstruction: googleContent{
			Parts: []googlePart{{Text: request.SystemPrompt}},
		},
		Contents: []googleContent{{
			Role:  "user",
			Parts: []googlePart{{Text: request.UserPrompt}},
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal Google AI Studio request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, googleAIStudioEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create Google AI Studio request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("x-goog-api-key", client.apiKey)

	response, err := client.httpClient.Do(httpRequest)
	if err != nil {
		// Do not propagate transport errors that could include credentials or URLs.
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", errors.New("call Google AI Studio API failed")
	}
	defer response.Body.Close()

	responseBody, err := readResponseBody(response.Body)
	if err != nil {
		return "", fmt.Errorf("read Google AI Studio response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Google AI Studio API returned status %d", response.StatusCode)
	}

	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return "", errors.New("decode Google AI Studio response: invalid JSON")
	}
	if len(completion.Error) > 0 && string(completion.Error) != "null" {
		return "", errors.New("Google AI Studio returned an API error")
	}
	if reason := completion.PromptFeedback.BlockReason; reason != "" && reason != "BLOCK_REASON_UNSPECIFIED" {
		return "", errors.New("Google AI Studio blocked the prompt")
	}
	if len(completion.Candidates) == 0 {
		return "", errors.New("Google AI Studio returned an empty completion")
	}
	if reason := completion.Candidates[0].FinishReason; reason != "" && reason != "STOP" {
		return "", errors.New("Google AI Studio returned a blocked or incomplete completion")
	}

	parts := completion.Candidates[0].Content.Parts
	texts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part.Thought {
			continue
		}
		if text := strings.TrimSpace(part.Text); text != "" {
			texts = append(texts, text)
		}
	}
	if len(texts) == 0 {
		return "", errors.New("Google AI Studio returned an empty completion")
	}

	return strings.Join(texts, "\n"), nil
}

type googleGenerateContentRequest struct {
	GenerationConfig  googleGenerationConfig `json:"generationConfig"`
	SystemInstruction googleContent          `json:"system_instruction"`
	Contents          []googleContent        `json:"contents"`
}

type googleGenerationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
}

type googleContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"`
}

type googleGenerateContentResponse struct {
	Error          json.RawMessage `json:"error"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	Candidates []struct {
		Content      googleContent `json:"content"`
		FinishReason string        `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}
