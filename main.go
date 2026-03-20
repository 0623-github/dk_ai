package main

import (
	"context"
	"os"

	"github.com/0623-github/dk_ai/biz/handler"
	"github.com/0623-github/dk_ai/biz/wrapper"
	"github.com/0623-github/dk_ai/lib/db"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

func main() {
	database, err := db.New("data/dk_ai.db")
	if err != nil {
		panic(err)
	}
	defer database.Close()

	h := server.New(server.WithHostPorts("0.0.0.0:9090"))

	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// API 路由
	w := wrapper.NewImpl(context.Background(), database)
	h2 := &handler.Handler{Wrapper: w}
	register(h, h2)

	// 静态文件 - 放在 API 后面作为 fallback
	feRoot := "fe-react/dist"
	if _, err := os.Stat(feRoot + "/index.html"); err == nil {
		h.StaticFS("/", &app.FS{Root: feRoot, IndexNames: []string{"index.html"}})
	}

	h.Spin()
}
