package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0623-github/dk_ai/biz/wrapper"
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
