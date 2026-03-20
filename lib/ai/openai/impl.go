package openai

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/0623-github/dk_ai/lib/helper"
	"github.com/cloudwego/hertz/cmd/hz/util/logs"
	openai_sdk "github.com/sashabaranov/go-openai"
)

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	User    string        `json:"user"`
	Message string        `json:"message"`
	History []ChatMessage `json:"history,omitempty"`
}

// Impl OpenAI 兼容实现（也用于 Kimi）
type Impl struct {
	client   *openai_sdk.Client
	model    string
	apiKey   string
	mockMode bool
	mu       sync.RWMutex
}

// NewImplFromConfig 从配置创建 OpenAI 客户端
func NewImplFromConfig(ctx context.Context, conf helper.OpenAICompatibleConfig) *Impl {
	if conf.APIKey == "" {
		logs.Warnf("api key is empty, using mock mode")
		return &Impl{mockMode: true}
	}

	config := openai_sdk.DefaultConfig(conf.APIKey)
	if conf.BaseURL != "" {
		config.BaseURL = conf.BaseURL
	}
	client := openai_sdk.NewClientWithConfig(config)

	impl := &Impl{
		client: client,
		model:  conf.Model,
		apiKey: conf.APIKey,
	}

	// 异步检查连接
	go impl.checkConnection(ctx)

	logs.Infof("openai impl created, model: %s", conf.Model)
	return impl
}

// NewKimiImpl 创建 Kimi 客户端
func NewKimiImpl(ctx context.Context) *Impl {
	conf, err := helper.GetConfig[helper.AIConfig](ctx, helper.AIConfigPath)
	if err != nil {
		logs.Warnf("failed to load config: %v, using mock mode", err)
		return &Impl{mockMode: true}
	}
	return NewImplFromConfig(ctx, conf.Kimi)
}

// NewOpenAIImpl 创建 OpenAI 客户端
func NewOpenAIImpl(ctx context.Context) *Impl {
	conf, err := helper.GetConfig[helper.AIConfig](ctx, helper.AIConfigPath)
	if err != nil {
		logs.Warnf("failed to load config: %v, using mock mode", err)
		return &Impl{mockMode: true}
	}
	return NewImplFromConfig(ctx, conf.OpenAI)
}

func (i *Impl) checkConnection(ctx context.Context) {
	_, err := i.client.CreateChatCompletion(ctx, openai_sdk.ChatCompletionRequest{
		Model: i.model,
		Messages: []openai_sdk.ChatCompletionMessage{
			{Role: openai_sdk.ChatMessageRoleUser, Content: "hi"},
		},
	})
	if err != nil {
		logs.Warnf("connection check failed: %v, will use mock mode", err)
		i.mu.Lock()
		i.mockMode = true
		i.mu.Unlock()
	}
}

func (i *Impl) isMockMode() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.mockMode
}

func (i *Impl) Chat(ctx context.Context, req ChatRequest) (string, error) {
	if i.isMockMode() {
		return i.mockChat(req.Message), nil
	}

	messages := i.buildMessages(req)
	resp, err := i.client.CreateChatCompletion(ctx, openai_sdk.ChatCompletionRequest{
		Model:    i.model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no choices")
	}
	return resp.Choices[0].Message.Content, nil
}

func (i *Impl) ChatStream(ctx context.Context, req ChatRequest, callback func(string, bool)) error {
	if i.isMockMode() {
		reply := i.mockChat(req.Message)
		for j, char := range reply {
			callback(string(char), j == len(reply)-1)
		}
		return nil
	}

	messages := i.buildMessages(req)
	stream, err := i.client.CreateChatCompletionStream(ctx, openai_sdk.ChatCompletionRequest{
		Model:    i.model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return err
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		if len(response.Choices) > 0 {
			callback(response.Choices[0].Delta.Content, false)
		}
	}
	callback("", true)
	return nil
}

func (i *Impl) buildMessages(req ChatRequest) []openai_sdk.ChatCompletionMessage {
	messages := make([]openai_sdk.ChatCompletionMessage, 0)
	for _, h := range req.History {
		role := openai_sdk.ChatMessageRoleUser
		if h.Role == "assistant" {
			role = openai_sdk.ChatMessageRoleAssistant
		} else if h.Role == "system" {
			role = openai_sdk.ChatMessageRoleSystem
		}
		messages = append(messages, openai_sdk.ChatCompletionMessage{
			Role:    role,
			Content: h.Content,
		})
	}
	messages = append(messages, openai_sdk.ChatCompletionMessage{
		Role:    openai_sdk.ChatMessageRoleUser,
		Content: req.Message,
	})
	return messages
}

func (i *Impl) mockChat(message string) string {
	msg := strings.ToLower(message)
	if strings.Contains(msg, "你好") || strings.Contains(msg, "hello") {
		return "你好！我是 AI 助手。当前使用的是模拟模式，请检查 API Key 配置。"
	}
	return "我收到了你的消息。当前服务处于模拟模式，请配置正确的 API Key。"
}
