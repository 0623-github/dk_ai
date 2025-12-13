package ollama

import (
	"context"
	"errors"

	"github.com/0623-github/dk_ai/lib/ai"
	"github.com/0623-github/dk_ai/lib/cache"
	"github.com/0623-github/dk_ai/lib/cache/local"
	"github.com/0623-github/dk_ai/lib/helper"
	"github.com/cloudwego/hertz/cmd/hz/util/logs"
	openai "github.com/sashabaranov/go-openai"
)

type Impl struct {
	cache   cache.Cache
	client  *openai.Client
	model   string
	baseURL string
}

type Config struct {
	Model   string `yaml:"model"`
	BaseURL string `yaml:"baseURL"`
}

func NewImpl(ctx context.Context) *Impl {
	conf, err := helper.GetConfig[Config](ctx, helper.AIConfigPath)
	if err != nil {
		panic(err)
	}
	config := openai.DefaultConfig("")
	config.BaseURL = conf.BaseURL
	client := openai.NewClientWithConfig(config)
	cache := local.NewImpl()
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: conf.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "你好",
			},
		},
	})
	if err != nil {
		panic("create chat completion failed: model " + conf.Model + ", baseURL: " + conf.BaseURL + ", err: " + err.Error())
	}
	if len(resp.Choices) == 0 {
		panic("no choices")
	}
	firstChoice := resp.Choices[0]
	if firstChoice.Message.Content == "" {
		panic("empty content")
	}
	logs.Infof("ollama impl created, model: %s, baseURL: %s, answer: %s", conf.Model, conf.BaseURL, firstChoice.Message.Content)
	return &Impl{cache: cache, client: client, model: conf.Model, baseURL: conf.BaseURL}
}

func (i *Impl) Chat(ctx context.Context, req ai.ChatRequest) (string, error) {
	messages, err := i.cache.Get(req.User)
	if err != nil {
		return "", err
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	})
	resp, err := i.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    i.model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no choices")
	}
	firstChoice := resp.Choices[0]
	if firstChoice.Message.Content == "" {
		return "", errors.New("empty content")
	}
	messages = append(messages, firstChoice.Message)
	err = i.cache.Set(req.User, messages, 0)
	if err != nil {
		return "", err
	}
	return firstChoice.Message.Content, nil
}
