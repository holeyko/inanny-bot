package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

func (client *GoogleAIClient) Complete(ctx context.Context, request CompletionRequest) (string, error) {
	if client.apiKey == "" {
		return "", ErrNotConfigured
	}

	payload := googleGenerateContentRequest{
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

	endpoint, err := url.Parse(googleAIStudioEndpoint)
	if err != nil {
		return "", fmt.Errorf("create Google AI Studio request: %w", err)
	}
	query := endpoint.Query()
	query.Set("key", client.apiKey)
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create Google AI Studio request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return "", fmt.Errorf("call Google AI Studio API: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read Google AI Studio response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Google AI Studio API returned status %d: %s", response.StatusCode, string(responseBody))
	}

	var completion googleGenerateContentResponse
	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return "", fmt.Errorf("decode Google AI Studio response: %w", err)
	}
	if len(completion.Candidates) == 0 {
		return "", errors.New("Google AI Studio returned an empty completion")
	}

	parts := completion.Candidates[0].Content.Parts
	texts := make([]string, 0, len(parts))
	for _, part := range parts {
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
	SystemInstruction googleContent   `json:"system_instruction"`
	Contents          []googleContent `json:"contents"`
}

type googleContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text string `json:"text"`
}

type googleGenerateContentResponse struct {
	Candidates []struct {
		Content googleContent `json:"content"`
	} `json:"candidates"`
}
