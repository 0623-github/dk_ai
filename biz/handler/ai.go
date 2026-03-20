package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0623-github/dk_ai/biz/wrapper"
	"github.com/0623-github/dk_ai/lib/helper"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func (h *Handler) Chat(ctx context.Context, c *app.RequestContext) {
	req := &wrapper.ChatRequest{}
	if err := c.Bind(req); err != nil {
		c.JSON(consts.StatusBadRequest, wrapper.ChatResp{
			Reply: "Invalid request",
		})
		return
	}

	resp, err := h.Wrapper.Chat(ctx, *req)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, wrapper.ChatResp{
			Reply: err.Error(),
		})
		return
	}
	c.JSON(consts.StatusOK, resp)
}

func (h *Handler) ChatStream(ctx context.Context, c *app.RequestContext) {
	req := &wrapper.ChatRequest{}
	if err := c.Bind(req); err != nil {
		c.JSON(consts.StatusBadRequest, wrapper.ChatResp{
			Reply: "Invalid request",
		})
		return
	}

	// 设置 SSE 响应头
	c.SetContentType("text/event-stream")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")
	c.SetStatusCode(consts.StatusOK)

	// 发送流式响应
	streamCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := h.Wrapper.ChatStream(streamCtx, *req, func(chunk string, done bool) {
		data, _ := json.Marshal(map[string]interface{}{
			"content": chunk,
			"done":    done,
		})
		fmt.Fprintf(c, "data: %s\n\n", data)
		c.Flush()
	})

	if err != nil {
		data, _ := json.Marshal(map[string]interface{}{
			"error": err.Error(),
			"done":  true,
		})
		fmt.Fprintf(c, "data: %s\n\n", data)
	}
}

// CreateSession 创建会话
func (h *Handler) CreateSession(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Title string `json:"title"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	session, err := h.Wrapper.CreateSession(req.Title)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, session)
}

// GetSession 获取会话
func (h *Handler) GetSession(ctx context.Context, c *app.RequestContext) {
	id := c.Query("id")
	if id == "" {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "session id required"})
		return
	}

	session, err := h.Wrapper.GetSession(id)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if session == nil {
		c.JSON(consts.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}
	c.JSON(consts.StatusOK, session)
}

// ListSessions 列出所有会话
func (h *Handler) ListSessions(ctx context.Context, c *app.RequestContext) {
	sessions, err := h.Wrapper.ListSessions(100)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{
		"sessions": sessions,
	})
}

// UpdateSession 更新会话
func (h *Handler) UpdateSession(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "session id required"})
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if err := h.Wrapper.UpdateSessionTitle(id, req.Title); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]string{"message": "updated"})
}

// DeleteSession 删除会话
func (h *Handler) DeleteSession(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "session id required"})
		return
	}

	if err := h.Wrapper.DeleteSession(id); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]string{"message": "deleted"})
}

// GetMessages 获取会话消息
func (h *Handler) GetMessages(ctx context.Context, c *app.RequestContext) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "session_id required"})
		return
	}

	messages, err := h.Wrapper.GetMessages(sessionID, 1000)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, map[string]interface{}{
		"messages": messages,
	})
}

// GetAIConfig 获取当前 AI 配置
func (h *Handler) GetAIConfig(ctx context.Context, c *app.RequestContext) {
	conf, err := helper.GetConfig[helper.AIConfig](ctx, "conf/ai.yaml")
	if err != nil {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"provider": "ollama",
			"model":    "gemma:2b",
		})
		return
	}

	// 隐藏 API Key
	safeConf := map[string]interface{}{
		"provider": conf.Provider,
		"ollama": map[string]string{
			"model":   conf.Ollama.Model,
			"baseURL": conf.Ollama.BaseURL,
		},
		"kimi": map[string]string{
			"model":   conf.Kimi.Model,
			"baseURL": conf.Kimi.BaseURL,
		},
		"openai": map[string]string{
			"model":   conf.OpenAI.Model,
			"baseURL": conf.OpenAI.BaseURL,
		},
	}
	c.JSON(consts.StatusOK, safeConf)
}

// GetAvailableProviders 获取可用的模型提供商列表
func (h *Handler) GetAvailableProviders(ctx context.Context, c *app.RequestContext) {
	conf, err := helper.GetConfig[helper.AIConfig](ctx, "conf/ai.yaml")
	if err != nil {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"providers": []map[string]interface{}{
				{"id": "ollama", "name": "Ollama (本地)", "available": true},
			},
		})
		return
	}

	providers := []map[string]interface{}{
		{"id": "ollama", "name": "Ollama (本地)", "model": conf.Ollama.Model, "available": true},
	}

	// Kimi 需要 API Key
	kimiAvailable := conf.Kimi.APIKey != ""
	providers = append(providers, map[string]interface{}{
		"id":        "kimi",
		"name":      "Kimi (Moonshot)",
		"model":     conf.Kimi.Model,
		"available": kimiAvailable,
	})

	// OpenAI 需要 API Key
	openaiAvailable := conf.OpenAI.APIKey != ""
	providers = append(providers, map[string]interface{}{
		"id":        "openai",
		"name":      "OpenAI",
		"model":     conf.OpenAI.Model,
		"available": openaiAvailable,
	})

	c.JSON(consts.StatusOK, map[string]interface{}{
		"providers":        providers,
		"current_provider": conf.Provider,
	})
}
