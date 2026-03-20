// Package kimi 提供 Kimi (Moonshot) AI 支持
// Kimi 使用 OpenAI 兼容接口
package kimi

import (
	"github.com/0623-github/dk_ai/lib/ai/openai"
)

// Config 是 openai.Config 的别名
type Config = openai.Config

// Impl 是 openai.Impl 的别名
type Impl = openai.Impl

// NewImpl 是 openai.NewImpl 的别名
var NewImpl = openai.NewImpl
