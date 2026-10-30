package ai

import (
	"context"
	"errors"
)

var ErrNotConfigured = errors.New("ai client is not configured")

type CompletionRequest struct {
	SystemPrompt    string
	UserPrompt      string
	ChatID          int64
	MaxOutputTokens int
}

type AIClient interface {
	Complete(ctx context.Context, request CompletionRequest) (string, error)
}
