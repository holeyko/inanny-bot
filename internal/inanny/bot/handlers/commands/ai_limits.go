package command

import (
	"log"
	"os"
	"strconv"
	"sync"
)

// Character limits are conservative input-size budgets, not tokenizer counts.
var summaryInputChars = boundedEnvInt("AI_MAX_INPUT_CHARS", 32000, 1000, 60000)
var summaryOutputTokens = boundedEnvInt("AI_MAX_OUTPUT_TOKENS", 2048, 256, 4096)
var summarySlots = make(chan struct{}, boundedEnvInt("AI_MAX_CONCURRENT_SUMMARIES", 3, 1, 8))
var summaryChats sync.Map

func boundedEnvInt(name string, fallback, minimum, maximum int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < minimum || n > maximum {
		log.Printf("invalid %s; using default %d", name, fallback)
		return fallback
	}
	return n
}

// Admission is non-blocking, before history reads or quota reservations.
func acquireSummary(chatID int64) (func(), bool) {
	if _, loaded := summaryChats.LoadOrStore(chatID, struct{}{}); loaded {
		return nil, false
	}
	select {
	case summarySlots <- struct{}{}:
		return func() { <-summarySlots; summaryChats.Delete(chatID) }, true
	default:
		summaryChats.Delete(chatID)
		return nil, false
	}
}
