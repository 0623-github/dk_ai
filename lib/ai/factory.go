package ai

import (
	"context"
	"fmt"

	"github.com/0623-github/dk_ai/lib/ai/openai"
	"github.com/0623-github/dk_ai/lib/ai/ollama"
	"github.com/0623-github/dk_ai/lib/helper"
	"github.com/cloudwego/hertz/cmd/hz/util/logs"
)

// AI 接口
type AI interface {
	Chat(ctx context.Context, req ChatRequest) (string, error)
	ChatStream(ctx context.Context, req ChatRequest, callback func(string, bool)) error
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	User    string        `json:"user"`
	Message string        `json:"message"`
	History []ChatMessage `json:"history,omitempty"`
}

// NewAI 根据配置创建对应的 AI 客户端
func NewAI(ctx context.Context) (AI, error) {
	conf, err := helper.GetConfig[helper.AIConfig](ctx, helper.AIConfigPath)
	if err != nil {
		logs.Warnf("failed to load ai config: %v, using default ollama", err)
		return NewOllamaAdapter(ollama.NewImpl(ctx)), nil
	}

	logs.Infof("using AI provider: %s", conf.Provider)

	switch conf.Provider {
	case "kimi":
		return NewOpenAIAdapter(openai.NewKimiImpl(ctx)), nil
	case "openai":
		return NewOpenAIAdapter(openai.NewOpenAIImpl(ctx)), nil
	case "ollama":
		return NewOllamaAdapter(ollama.NewImpl(ctx)), nil
	default:
		logs.Warnf("unknown provider: %s, fallback to ollama", conf.Provider)
		return NewOllamaAdapter(ollama.NewImpl(ctx)), nil
	}
}

// NewAIWithProvider 指定提供商创建 AI 客户端
func NewAIWithProvider(ctx context.Context, provider string) (AI, error) {
	switch provider {
	case "kimi":
		return NewOpenAIAdapter(openai.NewKimiImpl(ctx)), nil
	case "openai":
		return NewOpenAIAdapter(openai.NewOpenAIImpl(ctx)), nil
	case "ollama":
		return NewOllamaAdapter(ollama.NewImpl(ctx)), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

// ollamaAdapter 适配 ollama.Impl 到 AI 接口
type ollamaAdapter struct {
	impl *ollama.Impl
}

func NewOllamaAdapter(impl *ollama.Impl) AI {
	return &ollamaAdapter{impl: impl}
}

func (a *ollamaAdapter) Chat(ctx context.Context, req ChatRequest) (string, error) {
	ollamaReq := ollama.ChatRequest{
		User:    req.User,
		Message: req.Message,
	}
	for _, h := range req.History {
		ollamaReq.History = append(ollamaReq.History, ollama.ChatMessage{
			Role:    h.Role,
			Content: h.Content,
		})
	}
	return a.impl.Chat(ctx, ollamaReq)
}

func (a *ollamaAdapter) ChatStream(ctx context.Context, req ChatRequest, callback func(string, bool)) error {
	ollamaReq := ollama.ChatRequest{
		User:    req.User,
		Message: req.Message,
	}
	for _, h := range req.History {
		ollamaReq.History = append(ollamaReq.History, ollama.ChatMessage{
			Role:    h.Role,
			Content: h.Content,
		})
	}
	return a.impl.ChatStream(ctx, ollamaReq, callback)
}

// openaiAdapter 适配 openai.Impl 到 AI 接口
type openaiAdapter struct {
	impl *openai.Impl
}

func NewOpenAIAdapter(impl *openai.Impl) AI {
	return &openaiAdapter{impl: impl}
}

func (a *openaiAdapter) Chat(ctx context.Context, req ChatRequest) (string, error) {
	openaiReq := openai.ChatRequest{
		User:    req.User,
		Message: req.Message,
	}
	for _, h := range req.History {
		openaiReq.History = append(openaiReq.History, openai.ChatMessage{
			Role:    h.Role,
			Content: h.Content,
		})
	}
	return a.impl.Chat(ctx, openaiReq)
}

func (a *openaiAdapter) ChatStream(ctx context.Context, req ChatRequest, callback func(string, bool)) error {
	openaiReq := openai.ChatRequest{
		User:    req.User,
		Message: req.Message,
	}
	for _, h := range req.History {
		openaiReq.History = append(openaiReq.History, openai.ChatMessage{
			Role:    h.Role,
			Content: h.Content,
		})
	}
	return a.impl.ChatStream(ctx, openaiReq, callback)
}
