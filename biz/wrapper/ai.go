package wrapper

import (
	"context"

	"github.com/0623-github/dk_ai/lib/ai"
	"github.com/0623-github/dk_ai/lib/ai/ollama"
)

type ChatRequest struct {
	SessionID string `json:"session_id"`
	User     string `json:"user"`
	Message  string `json:"message"`
	Mode     string `json:"mode"`
}

type ChatResp struct {
	SessionID  string `json:"session_id"`
	Reply      string `json:"reply"`
	Timestamp  int64  `json:"timestamp"`
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
	reply, err := w.AIClient.Chat(ctx, ai.ChatRequest{
		User:    req.User,
		Message: req.Message,
	})
	
	resp = ChatResp{
		SessionID: req.SessionID,
		Reply:     reply,
		Timestamp: 0,
	}
	return resp, err
}
