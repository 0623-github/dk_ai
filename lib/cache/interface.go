package cache

import "github.com/sashabaranov/go-openai"

type Cache interface {
	Get(key string) ([]openai.ChatCompletionMessage, error)
	Set(key string, value []openai.ChatCompletionMessage, expire int64) error
}
