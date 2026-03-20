package ai

import "context"

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

// StreamChunk 流式响应数据块
type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}
