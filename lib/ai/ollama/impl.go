package ollama

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/0623-github/dk_ai/lib/helper"
	"github.com/cloudwego/hertz/cmd/hz/util/logs"
	openai "github.com/sashabaranov/go-openai"
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

type Impl struct {
	client   *openai.Client
	model    string
	baseURL  string
	mockMode bool
	mu       sync.RWMutex
}

func NewImpl(ctx context.Context) *Impl {
	// 读取新的配置格式
	conf, err := helper.GetConfig[helper.AIConfig](ctx, helper.AIConfigPath)
	if err != nil {
		logs.Warnf("failed to load config: %v, using default ollama", err)
		return &Impl{
			mockMode: true,
		}
	}

	// 如果配置指定了其他 provider，但我们要创建 Ollama 实例
	// 使用 Ollama 配置
	ollamaConf := conf.Ollama
	if ollamaConf.Model == "" {
		ollamaConf.Model = "gemma:2b"
	}
	if ollamaConf.BaseURL == "" {
		ollamaConf.BaseURL = "http://localhost:11434/v1"
	}

	config := openai.DefaultConfig("")
	config.BaseURL = ollamaConf.BaseURL
	client := openai.NewClientWithConfig(config)

	impl := &Impl{
		client:  client,
		model:   ollamaConf.Model,
		baseURL: ollamaConf.BaseURL,
	}

	// 异步检查 Ollama 可用性
	go impl.checkOllama(ctx)

	logs.Infof("ollama impl created, model: %s, baseURL: %s", ollamaConf.Model, ollamaConf.BaseURL)
	return impl
}

func (i *Impl) checkOllama(ctx context.Context) {
	_, err := i.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: i.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: "hi"},
		},
	})
	if err != nil {
		logs.Warnf("ollama check failed: %v, will use mock mode", err)
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
	stream, err := i.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
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

func (i *Impl) buildMessages(req ChatRequest) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0)
	for _, h := range req.History {
		role := openai.ChatMessageRoleUser
		if h.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		} else if h.Role == "system" {
			role = openai.ChatMessageRoleSystem
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: h.Content,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	})
	return messages
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
	return "你好！我是 AI 助手。当前 Ollama 服务暂时不可用，这是模拟回复。"
}
