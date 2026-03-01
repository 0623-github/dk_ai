# AI 聊天助手

一个基于 Web 的 AI 聊天应用，支持多会话管理，后端使用 Go + Hertz 框架，前端使用 React + Tailwind CSS。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | React 18 + Vite + Tailwind CSS |
| 后端 | Go + Hertz |
| AI | Ollama (本地大模型) |
| 缓存 | 本地内存缓存 |

## 项目结构

```
dk_ai/
├── fe-react/           # 前端 React 应用
│   ├── src/
│   │   ├── App.jsx     # 主应用组件
│   │   ├── main.jsx    # 入口文件
│   │   └── index.css   # 全局样式
│   ├── index.html
│   ├── vite.config.js
│   └── tailwind.config.js
├── biz/                # 业务逻辑层
│   ├── handler/        # HTTP 处理器
│   ├── wrapper/        # 业务包装层
│   └── router/         # 路由注册
├── lib/                # 核心库
│   ├── ai/             # AI 客户端
│   ├── cache/          # 缓存实现
│   └── helper/         # 工具函数
├── conf/               # 配置文件
│   └── openai.yaml     # AI 配置
├── docs/               # 设计文档
└── main.go             # 入口文件
```

## 快速开始

### 前置条件

- Go 1.18+
- Node.js 18+
- Ollama (可选，用于真实 AI 对话)

### 1. 启动后端服务

```bash
# 方式一：直接运行
cd /path/to/dk_ai
go run .

# 方式二：使用脚本
./script/start.sh
```

后端服务默认监听 `http://localhost:9090`

### 2. 启动前端开发服务器

```bash
cd fe-react
npm install
npm run dev
```

前端默认访问 `http://localhost:5173`

### 3. (可选) 启动 Ollama

如需使用真实 AI 对话，启动 Ollama：

```bash
# 启动 Ollama 服务
ollama serve

# 下载模型
ollama pull qwen:7b
```

未启动 Ollama 时，系统会自动使用模拟模式。

## API 接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/ping` | 健康检查 |
| POST | `/chat` | 发送聊天消息 |

### 聊天接口

**请求**

```json
POST /chat
Content-Type: application/json

{
  "session_id": "会话ID（可选）",
  "message": "用户消息",
  "mode": "text"
}
```

**响应**

```json
{
  "session_id": "会话ID",
  "reply": "AI 回复",
  "timestamp": 1234567890
}
```

## 功能特性

- [x] 多会话管理
- [x] 消息持久化（LocalStorage）
- [x] 模拟/真实 AI 模式切换
- [x] 响应式设计
- [x] 加载状态展示

## 开发相关

### 构建前端

```bash
cd fe-react
npm run build
```

### 代码格式

```bash
# Go 格式化
go fmt ./...

# 前端代码检查
cd fe-react
npm run lint
```

## 文档

- [系统设计文档](./docs/system_design_xiaoai_chat.md)
- [前端设计文档](./docs/frontend_design.md)

## 许可证

MIT
