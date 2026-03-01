package ollama

import (
	"context"
	"errors"
	"strings"

	"github.com/0623-github/dk_ai/lib/ai"
	"github.com/0623-github/dk_ai/lib/cache"
	"github.com/0623-github/dk_ai/lib/cache/local"
	"github.com/0623-github/dk_ai/lib/helper"
	"github.com/cloudwego/hertz/cmd/hz/util/logs"
	openai "github.com/sashabaranov/go-openai"
)

type Impl struct {
	cache    cache.Cache
	client   *openai.Client
	model    string
	baseURL  string
	mockMode bool
}

type Config struct {
	Model   string `yaml:"model"`
	BaseURL string `yaml:"baseURL"`
}

func NewImpl(ctx context.Context) *Impl {
	conf, err := helper.GetConfig[Config](ctx, helper.AIConfigPath)
	if err != nil {
		logs.Warnf("failed to load config: %v, using mock mode", err)
		return &Impl{
			cache:    local.NewImpl(),
			mockMode: true,
		}
	}

	config := openai.DefaultConfig("")
	config.BaseURL = conf.BaseURL
	client := openai.NewClientWithConfig(config)

	_, err = client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: conf.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "你好",
			},
		},
	})
	if err != nil {
		logs.Warnf("ollama not available: %v, using mock mode", err)
		return &Impl{
			cache:    local.NewImpl(),
			mockMode: true,
		}
	}

	logs.Infof("ollama impl created, model: %s, baseURL: %s", conf.Model, conf.BaseURL)
	return &Impl{cache: local.NewImpl(), client: client, model: conf.Model, baseURL: conf.BaseURL}
}

func (i *Impl) Chat(ctx context.Context, req ai.ChatRequest) (string, error) {
	if i.mockMode {
		return i.mockChat(req.Message), nil
	}

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

var mockResponses = []string{
	"你好！有什么我可以帮助你的吗？",
	"我收到了你的消息。这是一个模拟回复，因为 Ollama 服务暂时不可用。",
	"感谢你的提问！如果你启动了 Ollama 服务，我就能提供更智能的回答了。",
	"我正在以模拟模式运行。请确保 Ollama 服务正在运行，然后我可以帮你与 AI 模型对话。",
	"你好！我是 AI 助手。当前使用的是模拟模式响应。",
}

func (i *Impl) mockChat(message string) string {
	msg := strings.ToLower(message)
	
	if strings.Contains(msg, "你好") || strings.Contains(msg, "hello") || strings.Contains(msg, "hi") {
		return "你好！很高兴见到你！有什么我可以帮助你的吗？"
	}
	if strings.Contains(msg, "天气") {
		return "抱歉，我无法获取实时天气信息。但我希望今天你有一个好心情！"
	}
	if strings.Contains(msg, "名字") || strings.Contains(msg, "你是谁") {
		return "我是 AI 聊天助手，基于 Ollama 本地模型运行。"
	}
	if strings.Contains(msg, "帮助") || strings.Contains(msg, "能做什么") {
		return "我可以帮你：\n1. 回答问题\n2. 聊天对话\n3. 提供信息\n4. 编写代码\n\n注意：当前是模拟模式，启动 Ollama 后可获得更智能的回答。"
	}
	
	return mockResponses[len(message)%len(mockResponses)]
}
