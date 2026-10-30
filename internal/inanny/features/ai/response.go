package ai

import (
	"errors"
	"io"
	"log"
	"time"
)

const maxResponseBytes = 1 << 20

func outputTokenBudget(requested int) int {
	if requested <= 0 {
		return 2048
	}
	if requested > 4096 {
		return 4096
	}
	return requested
}

func readResponseBody(reader io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, maxResponseBytes+1))
	if err != nil {
		return nil, errors.New("could not read AI response")
	}
	if len(body) > maxResponseBytes {
		return nil, errors.New("AI response exceeds size limit")
	}
	return body, nil
}

// Never log prompts, completions, error bodies, credentials, or request URLs.
func logCompletion(provider, model string, started time.Time, inputTokens, outputTokens, totalTokens int, err error) {
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	log.Printf("ai provider=%q model=%q input_tokens=%d output_tokens=%d total_tokens=%d latency_ms=%d outcome=%s",
		provider, model, inputTokens, outputTokens, totalTokens, time.Since(started).Milliseconds(), outcome)
}
