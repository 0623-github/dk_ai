package wrapper

import (
	"context"

	"github.com/0623-github/dk_ai/lib/ai"
	"github.com/0623-github/dk_ai/lib/ai/ollama"
)

type ChatRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

type ChatResp struct {
	Message string `json:"message"`
}

type Wrapper interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResp, error)
}

type Impl struct {
	AIClient ai.AI
}

func NewImpl(ctx context.Context) *Impl {
	aiClient := ollama.NewImpl(ctx)
	return &Impl{AIClient: aiClient}
}

func (w *Impl) Chat(ctx context.Context, req ChatRequest) (resp ChatResp, err error) {
	resp.Message, err = w.AIClient.Chat(ctx, ai.ChatRequest{
		User:    req.User,
		Message: req.Message,
	})
	return resp, err
}
