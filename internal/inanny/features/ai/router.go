package ai

import (
	"context"
	"log"
	"time"
)

// Quota describes the current primary-provider allowance for one chat.
type Quota struct {
	Used     int
	Limit    int
	ResetsAt time.Time
}

// QuotaStore must reserve attempts atomically. Failed attempts are not refunded.
type QuotaStore interface {
	Reserve(context.Context, int64) (bool, error)
	Usage(context.Context, int64) (Quota, error)
}

type Router struct {
	primary  AIClient
	fallback AIClient
	quota    QuotaStore
}

func NewRouter(primary, fallback AIClient, quota QuotaStore) *Router {
	return &Router{primary: primary, fallback: fallback, quota: quota}
}

func (router *Router) PrimaryConfigured() bool {
	return router.primary != nil
}

func (router *Router) Usage(ctx context.Context, chatID int64) (Quota, error) {
	if err := ctx.Err(); err != nil {
		return Quota{}, err
	}
	if router.quota == nil {
		return Quota{}, ErrNotConfigured
	}
	return router.quota.Usage(ctx, chatID)
}

func (router *Router) Complete(ctx context.Context, request CompletionRequest) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	// A missing or unavailable quota store must never permit primary usage.
	if router.primary != nil && router.quota != nil {
		text, err := router.completePrimary(ctx, request)
		if err == nil {
			return text, nil
		}
	} else {
		log.Printf("AI route chat=%d provider=fallback reason=primary_or_quota_unconfigured", request.ChatID)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if router.fallback == nil {
		return "", ErrNotConfigured
	}
	fallbackCtx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	if err := fallbackCtx.Err(); err != nil {
		return "", err
	}
	return router.fallback.Complete(fallbackCtx, request)
}

func (router *Router) completePrimary(ctx context.Context, request CompletionRequest) (string, error) {
	primaryCtx, cancel := context.WithTimeout(ctx, 55*time.Second)
	defer cancel()
	if err := primaryCtx.Err(); err != nil {
		return "", err
	}
	allowed, err := router.quota.Reserve(primaryCtx, request.ChatID)
	if err != nil {
		log.Printf("AI route chat=%d provider=fallback reason=quota_unavailable", request.ChatID)
		return "", err
	}
	if !allowed {
		log.Printf("AI route chat=%d provider=fallback reason=quota_exhausted", request.ChatID)
		return "", ErrNotConfigured
	}
	if err := primaryCtx.Err(); err != nil {
		return "", err
	}
	text, err := router.primary.Complete(primaryCtx, request)
	if err != nil {
		log.Printf("AI route chat=%d provider=fallback reason=primary_failed", request.ChatID)
	}
	return text, err
}
