package handler

import (
	"context"

	"github.com/0623-github/dk_ai/biz/wrapper"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func (h *Handler) Chat(ctx context.Context, c *app.RequestContext) {
	req := &wrapper.ChatRequest{}
	handlerWrapper(ctx, c, func(ctx context.Context) (int, interface{}, error) {
		resp, err := h.Wrapper.Chat(ctx, *req)
		if err != nil {
			return consts.StatusInternalServerError, nil, err
		}
		return consts.StatusOK, resp, nil
	})
}
