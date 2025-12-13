package local

import (
	"errors"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

type Impl struct {
	cache sync.Map
}

func NewImpl() *Impl {
	return &Impl{cache: sync.Map{}}
}

func (i *Impl) Get(key string) ([]openai.ChatCompletionMessage, error) {
	if key == "" {
		return []openai.ChatCompletionMessage{}, nil
	}
	value, ok := i.cache.Load(key)
	if !ok {
		return []openai.ChatCompletionMessage{}, errors.New("key not found")
	}
	return value.([]openai.ChatCompletionMessage), nil
}

func (i *Impl) Set(key string, value []openai.ChatCompletionMessage, expire int64) error {
	if key == "" {
		return nil
	}
	i.cache.Store(key, value)
	return nil
}
