package ai

import "context"

type AI interface {
	Chat(ctx context.Context, req ChatRequest) (string, error)
}

type ChatRequest struct {
	User    string
	Message string
}
