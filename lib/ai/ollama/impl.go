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

	// 构建消息列表（历史 + 当前消息）
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
	firstChoice := resp.Choices[0]
	if firstChoice.Message.Content == "" {
		return "", errors.New("empty content")
	}

	return firstChoice.Message.Content, nil
}

// ChatStream 流式聊天实现
func (i *Impl) ChatStream(ctx context.Context, req ai.ChatRequest, callback func(string, bool)) error {
	if i.mockMode {
		// 模拟模式下逐字输出
		reply := i.mockChat(req.Message)
		for j, char := range reply {
			callback(string(char), j == len(reply)-1)
		}
		return nil
	}

	// 构建消息列表
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
			content := response.Choices[0].Delta.Content
			callback(content, false)
		}
	}

	callback("", true) // 发送完成标记
	return nil
}

// buildMessages 构建 OpenAI 格式的消息列表
func (i *Impl) buildMessages(req ai.ChatRequest) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0)

	// 添加历史消息
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

	// 添加当前消息
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	})

	return messages
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
