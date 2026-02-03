package graph

import (
	"context"
	"crypto/sha1"
	"fmt"

	"ai-agent-system/cache"
	"ai-agent-system/service"
	"ai-agent-system/graph/model"
)

type Resolver struct {
	Redis *cache.RedisClient
	AI    *service.AIService
}

func hashPrompt(prompt string) string {
	h := sha1.Sum([]byte(prompt))
	return fmt.Sprintf("%x", h)
}

func (r *Resolver) AskAI(ctx context.Context, prompt string) (*model.AIResponse, error) {
	key := "ai:" + hashPrompt(prompt)

	if val, err := r.Redis.Get(key); err == nil {
		return &model.AIResponse{
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

	return &model.AIResponse{
		Prompt: prompt,
		Result: result,
		Cached: false,
	}, nil
}

type QueryResolver interface {
	AskAI(ctx context.Context, prompt string) (*model.AIResponse, error)
}

type ResolverRoot interface {
	Query() QueryResolver
}
