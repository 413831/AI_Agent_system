package graph

import (
	"context"
	"crypto/sha1"
	"fmt"

	"ai-agent-system/cache"
	"ai-agent-system/service"
)

type Resolver struct {
	Redis *cache.RedisClient
	AI    *service.AIService
}

func hashPrompt(prompt string) string {
	h := sha1.Sum([]byte(prompt))
	return fmt.Sprintf("%x", h)
}

func (r *Resolver) AskAI(ctx context.Context, prompt string) (*AIResponse, error) {
	key := "ai:" + hashPrompt(prompt)

	if val, err := r.Redis.Get(key); err == nil {
		return &AIResponse{
			Prompt: prompt,
			Result: val,
			Cached: true,
		}, nil
	}

	result, err := r.AI.Ask(prompt)
	if err != nil {
		return nil, err
	}

	_ = r.Redis.Set(key, result)

	return &AIResponse{
		Prompt: prompt,
		Result: result,
		Cached: false,
	}, nil
}
