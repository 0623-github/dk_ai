package handler

import (
	"context"

	"github.com/0623-github/dk_ai/biz/wrapper"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

type Handler struct {
	Wrapper wrapper.Wrapper
}

func handlerWrapper(ctx context.Context, c *app.RequestContext, f func(ctx context.Context) (int, interface{}, error)) {
	code, resp, err := f(ctx)
	if err != nil {
		c.JSON(code, utils.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(code, resp)
}
