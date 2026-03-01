package handler

import (
	"context"

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
