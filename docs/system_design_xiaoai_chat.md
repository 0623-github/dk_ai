# AI 聊天系统设计文档（支持小爱音箱）

---

## 1. 背景与目标

### 问题陈述

随着智能家居的普及，用户希望通过语音交互与 AI 系统进行对话。目前市场上的 AI 对话系统多为纯文本交互，缺乏语音场景的整合能力。用户希望能够通过小米音箱（小爱同学）直接与 AI 系统进行自然语言对话，实现更便捷的交互体验。

### 目标

- 实现基于 HTTP 的聊天接口，支持文本消息的发送与接收
- 集成小米音箱（小爱同学）语音交互能力，实现语音对话功能
- 支持多轮对话上下文管理，保持会话连贯性
- 提供实时语音响应，支持流式输出
- 支持会话历史记录存储与查询

### 非目标

- 不支持视频通话或视频输出
- 不实现复杂的多设备联动场景
- 不支持第三方智能音箱（仅限小米音箱）
- 不实现复杂的情感识别或个性化推荐

---

## 2. 技术约束与设计原则

### 技术约束

- 后端服务使用 Go 语言开发（与现有 dk_ai 项目保持一致）
- Web 框架使用 Hertz（轻量级、高性能）
- AI 能力通过 Ollama 本地部署（支持离线运行）
- 小米音箱通过小爱开放平台 API 或 Mesh 协议对接
- 需要支持实时语音流式传输

### 设计原则

- **轻量化优先**：最小化外部依赖，保持系统简洁
- **模块化设计**：聊天模块与音箱模块解耦，便于单独扩展
- **低延迟响应**：语音交互延迟控制在 2 秒以内
- **容错优先**：网络异常时保证核心聊天功能可用
- **隐私保护**：用户对话数据本地存储，不上传云端

---

## 3. 系统架构

### 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         用户层                                   │
│  ┌─────────────┐              ┌─────────────────────────────┐ │
│  │  Web 界面    │              │    小米音箱（小爱同学）       │ │
│  └──────┬──────┘              └──────────────┬──────────────┘ │
└─────────┼───────────────────────────────────┼──────────────────┘
          │                                   │
          ▼                                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                        API 网关层                                 │
│                    （Hertz HTTP Server）                         │
│  ┌──────────────────┐  ┌─────────────────────────────────────┐ │
│  │  /api/chat       │  │      /api/xiaoai/* (小爱对接)        │ │
│  └────────┬─────────┘  └──────────────────┬─────────────────┘ │
└───────────┼───────────────────────────────┼───────────────────┘
            │                               │
            ▼                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      业务逻辑层                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │  聊天服务     │  │  会话管理     │  │   语音合成服务        │ │
│  │ ChatService  │  │ SessionMgr   │  │   TTS Service        │ │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘ │
└─────────┼─────────────────┼──────────────────────┼─────────────┘
          │                 │                      │
          ▼                 ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                       能力层                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │   Ollama     │  │   本地存储    │  │   小爱开放平台        │ │
│  │  (AI 引擎)   │  │ (会话历史)    │  │   (语音交互)          │ │
│  └──────────────┘  └──────────────┘  └──────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### 核心模块

| 模块 | 职责 | 关键接口 |
|------|------|----------|
| ChatService | 处理聊天请求，调用 AI 模型，返回响应 | `Chat(ctx, msg) string` |
| SessionManager | 管理多轮对话上下文，维护会话状态 | `CreateSession()`, `GetSession()`, `SaveContext()` |
| TTSService | 将文本转换为语音，支持流式输出 | `Synthesize(text) stream` |
| XiaoAIAdapter | 适配小爱音箱协议，处理语音输入输出 | `HandleVoiceInput()`, `PlayAudio()` |
| HistoryStore | 存储和查询会话历史 | `Save()`, `Query()`, `Delete()` |

### 数据流

**文本聊天流程**：
1. 用户通过 Web 界面发送消息
2. API 接收请求，创建或获取会话
3. ChatService 调用 Ollama 获取 AI 响应
4. 响应存储到会话历史
5. 返回结果给用户

**语音对话流程（小米音箱）**：
1. 用户唤醒小爱同学，说出唤醒词
2. 小爱音箱通过 Mesh 协议或开放平台回调发送语音请求
3. XiaoAIAdapter 接收语音数据
4. ASR（语音识别）转换为文本
5. ChatService 处理对话，获取 AI 响应
6. TTSService 将文本转换为语音
7. 语音流式返回给小米音箱播放

---

## 4. API 设计

### 接口列表

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/chat | 发送聊天消息 |
| GET | /api/chat/history | 获取会话历史 |
| POST | /api/chat/stream | 流式聊天（支持语音） |
| GET | /api/session/:id | 获取会话详情 |
| DELETE | /api/session/:id | 删除会话 |
| POST | /api/xiaoai/webhook | 小爱音箱回调接口 |
| POST | /api/tts | 文本转语音 |

### 详细接口定义

#### POST /api/chat

**请求体**：
```json
{
  "session_id": "可选，会话ID",
  "message": "用户消息内容",
  "mode": "text|voice"
}
```

**响应**：
```json
{
  "session_id": "会话ID",
  "reply": "AI 回复内容",
  "timestamp": 1699999999,
  "voice_url": "可选，语音回复地址"
}
```

#### POST /api/chat/stream

**请求体**：
```json
{
  "session_id": "可选",
  "message": "用户消息",
  "stream": true
}
```

**响应**：SSE 流式返回

#### POST /api/xiaoai/webhook

**请求体**：
```json
{
  "intent": "intent_name",
  "slots": {},
  "session_id": "小爱会话ID",
  "audio": "base64编码的音频"
}
```

**响应**：
```json
{
  "response_text": "回复文本",
  "audio_url": "语音URL",
  "continue": true
}
```

---

## 5. 数据库设计

### 表结构

#### sessions（会话表）

| 字段 | 类型 | 描述 |
|------|------|------|
| id | VARCHAR(36) | 主键，UUID |
| user_id | VARCHAR(64) | 用户标识 |
| title | VARCHAR(255) | 会话标题 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 最后活跃时间 |
| status | VARCHAR(20) | active/closed |

#### messages（消息表）

| 字段 | 类型 | 描述 |
|------|------|------|
| id | VARCHAR(36) | 主键 |
| session_id | VARCHAR(36) | 外键，关联会话 |
| role | VARCHAR(20) | user/assistant/system |
| content | TEXT | 消息内容 |
| voice_url | VARCHAR(512) | 语音链接 |
| created_at | TIMESTAMP | 创建时间 |

#### xiaoai_devices（小米设备表）

| 字段 | 类型 | 描述 |
|------|------|------|
| device_id | VARCHAR(64) | 设备唯一标识 |
| user_id | VARCHAR(64) | 用户ID |
| name | VARCHAR(128) | 设备名称 |
| status | VARCHAR(20) | online/offline |
| registered_at | TIMESTAMP | 注册时间 |

---

## 6. 错误处理

### 错误码定义

| 错误码 | 描述 | 处理方式 |
|--------|------|----------|
| 1000 | 参数错误 | 返回错误描述，引导用户重新输入 |
| 1001 | 会话不存在 | 自动创建新会话 |
| 1002 | AI 服务不可用 | 返回友好提示，记录日志 |
| 1003 | 语音识别失败 | 提示用户重新说话 |
| 1004 | 语音合成失败 | 返回文本响应 |
| 1005 | 小米设备不在线 | 通知用户设备状态 |
| 2001 | 请求超时 | 重试机制，返回部分结果 |
| 2002 | 服务内部错误 | 返回通用错误，触发告警 |

### 异常处理策略

- **AI 服务超时**：设置 10 秒超时，超时后返回 "抱歉，我正在思考中，请稍等"
- **小米设备离线**：推送通知给用户，保留离线消息
- **网络中断**：使用本地缓存队列，重试发送

---

## 7. 安全性

### 认证与授权

- API 请求使用 JWT Token 认证
- 小米设备使用设备密钥 + 签名验证
- 会话隔离，用户只能访问自己的会话

### 数据安全

- 用户对话数据加密存储（AES-256）
- 语音数据本地处理，不上传第三方
- API 通信使用 HTTPS

### 接口鉴权

```go
func AuthMiddleware() app.HandlerFunc {
    return func(c context.Context, ctx *app.RequestContext) {
        token := ctx.GetHeader("Authorization")
        if !validateToken(token) {
            ctx.AbortWithStatusJSON(401, ErrorResponse{Error: "Unauthorized"})
            return
        }
        ctx.Next(c)
    }
}
```

---

## 8. 性能考虑

### 性能指标

| 指标 | 目标值 |
|------|--------|
| 文本响应延迟 | < 500ms（P95 < 1s） |
| 语音首包延迟 | < 1s |
| 最大并发会话 | > 1000 |
| 语音识别准确率 | > 95% |

### 优化策略

- **缓存策略**：热门对话使用 Redis 缓存
- **连接池**：数据库连接池复用
- **流式输出**：AI 响应使用 SSE 流式返回，减少等待时间
- **异步处理**：语音识别和合成异步执行
- **预热模型**：服务启动时预加载 AI 模型

### 资源限制

- 单次对话最大 token 数：2048
- 会话最大历史消息数：50 条
- 语音单次最大时长：60 秒

---

## 9. 部署架构

### 部署架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              云服务器层                                       │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                         负载均衡 (Nginx/云 LB)                        │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                     │                                       │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                    API 服务 (Hertz Go 应用)                          │  │
│   │   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌──────────┐  │  │
│   │   │ ChatService │  │  SessionMgr │  │ XiaoAIAdapt│  │ TTSProxy │  │  │
│   │   └─────────────┘  └─────────────┘  └─────────────┘  └──────────┘  │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                     │                                       │
│   ┌──────────────────┐    ┌──────────────────┐    ┌────────────────────┐   │
│   │   MySQL          │    │   Redis          │    │   本地模型服务      │   │
│   │  (会话数据)       │    │  (缓存/会话)      │    │  (Ollama)          │   │
│   └──────────────────┘    └──────────────────┘    └────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                             第三方服务层                                      │
│                                                                             │
│   ┌──────────────────┐    ┌──────────────────┐    ┌────────────────────┐   │
│   │  小爱开放平台     │    │  小爱音箱设备    │    │   语音识别服务     │   │
│   │  (语音回调)      │    │  (终端用户)      │    │  (ASR)             │   │
│   └──────────────────┘    └──────────────────┘    └────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 部署组件清单

| 组件 | 描述 | 部署方式 | 数量 |
|------|------|----------|------|
| **API 服务** | Go 后端服务 (Hertz) | Docker/K8s | 2-4 实例 |
| **MySQL** | 会话和消息存储 | 云数据库 | 1 主 1 从 |
| **Redis** | 缓存、会话状态 | 云缓存 | 1 主 1 从 |
| **Ollama** | 本地 AI 模型服务 | Docker | 1-2 实例 |
| **Nginx** | 负载均衡、反向代理 | Docker | 2 实例 |
| **TTS 服务** | 语音合成服务 | Docker/K8s | 2 实例 |

### 云服务器配置建议

#### 方案一：轻量级部署（个人/小团队使用）

| 服务 | 配置 | 月费用参考 |
|------|------|-----------|
| 云服务器 (API + Ollama) | 4核16G + GPU | ¥300-500/月 |
| MySQL (云数据库) | 2核4G | ¥60/月 |
| Redis (云缓存) | 1G | ¥30/月 |
| **总计** | | **¥400-600/月** |

#### 方案二：生产级部署

| 服务 | 配置 | 月费用参考 |
|------|------|-----------|
| 云服务器 (API) | 4核8G × 3 | ¥300/月 |
| GPU 服务器 (Ollama) | 8核32G + GPU | ¥800-1500/月 |
| MySQL (高可用) | 4核8G | ¥200/月 |
| Redis (集群) | 2G | ¥80/月 |
| 负载均衡 | 按流量 | ¥100/月 |
| **总计** | | **¥1500-2200/月** |

### 部署流程

#### 1. 环境准备

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
```

#### 2. 配置文件

```yaml
# config/deploy.yaml
app:
  name: dk_ai_chat
  port: 8080
  env: production

database:
  host: mysql
  port: 3306
  name: chat_db
  user: chat_user
  password: ${DB_PASSWORD}

redis:
  host: redis
  port: 6379
  password: ${REDIS_PASSWORD}

ollama:
  endpoint: http://ollama:11434
  model: qwen:7b

xiaoai:
  app_key: ${XIAOAI_APP_KEY}
  app_secret: ${XIAOAI_APP_SECRET}
  webhook_url: ${WEBHOOK_URL}

tts:
  endpoint: http://tts-service:8000
```

#### 3. Docker Compose 部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  api:
    build: ./cmd/api
    ports:
      - "8080:8080"
    environment:
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    depends_on:
      - mysql
      - redis
      - ollama
    restart: unless-stopped

  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    restart: unless-stopped
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=${DB_PASSWORD}
      - MYSQL_DATABASE=chat_db
    volumes:
      - mysql_data:/var/lib/mysql
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - api
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
  ollama_data:
```

#### 4. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f api

# 重新加载配置
docker-compose restart api
```

### Nginx 配置

```nginx
upstream api_backend {
    server api:8080 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

server {
    listen 80;
    server_name your-domain.com;

    # API 代理
    location /api/ {
        proxy_pass http://api_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # SSE 流式响应
    location /api/chat/stream {
        proxy_pass http://api_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "keep-alive";
        proxy buffering off;
        proxy_cache off;
    }

    # 静态文件
    location /static/ {
        alias /var/www/static/;
        expires 30d;
    }
}
```

### 域名与 HTTPS

1. **域名解析**：在 DNS 控制台添加 A 记录指向服务器 IP
2. **HTTPS 证书**：使用 Let's Encrypt 免费证书

```bash
# 使用 certbot 获取证书
certbot --nginx -d your-domain.com

# 或使用 acme.sh
acme.sh --issue -d your-domain.com --nginx
```

### 健康检查与监控

```yaml
# docker-compose healthcheck
services:
  api:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

**监控指标**：
- 服务健康状态
- API 响应时间
- 并发连接数
- AI 模型响应延迟
- 数据库查询时间
- 内存/CPU 使用率

### CI/CD 部署流程

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker image
        run: docker-compose build api
        
      - name: Push to registry
        run: |
          docker tag dk_ai_api:latest registry.example.com/dk_ai_api:latest
          docker push registry.example.com/dk_ai_api:latest
      
      - name: Deploy to server
        run: |
          ssh user@server "docker-compose -f /opt/dk_ai/docker-compose.yml pull"
          ssh user@server "docker-compose -f /opt/dk_ai/docker-compose.yml up -d"
```

---

## 11. 测试计划

### 单元测试

| 模块 | 测试用例 |
|------|----------|
| ChatService | 单轮对话、多轮对话、异常处理 |
| SessionManager | 会话创建、上下文管理、过期清理 |
| TTSService | 文本转换、音频格式、边界情况 |
| XiaoAIAdapter | 协议解析、设备绑定、状态同步 |

### 集成测试

- API 接口测试（HTTP 客户端）
- AI 模型集成测试
- 小米设备模拟器测试
- 数据库读写测试

### 手动测试

- 真实小米设备语音对话测试
- 网络异常场景测试
- 多设备并发测试

---

## 12. 实施计划

### 里程碑

| 阶段 | 任务 | 预期时间 |
|------|------|----------|
| Phase 1 | 基础聊天功能 | 第 1-2 周 |
| | - 完善 ChatService | |
| | - 实现会话管理 | |
| | - API 接口开发 | |
| Phase 2 | 语音能力 | 第 3-4 周 |
| | - TTS 服务集成 | |
| | - ASR 服务集成 | |
| | - 流式响应支持 | |
| Phase 3 | 小米音箱集成 | 第 5-6 周 |
| | - 小爱开放平台对接 | |
| | - 设备管理 | |
| | - 语音对话测试 | |
| Phase 4 | 优化与上线 | 第 7-8 周 |
| | - 性能优化 | |
| | - 压力测试 | |
| | - 灰度上线 | |

---

## 13. 验收标准

- [ ] 用户通过 Web 界面发送消息，AI 在 1 秒内返回响应
- [ ] 支持至少 10 轮以上的多轮对话
- [ ] 会话历史可以正确存储和查询
- [ ] 文本转语音延迟小于 1 秒
- [ ] 小米音箱可以触发对话并播放 AI 响应
- [ ] 系统可承受 100 并发用户
- [ ] 单元测试覆盖率达到 70% 以上
- [ ] API 接口文档完整

---

## 14. 未来扩展

### 可扩展性设计

- **插件化 AI 引擎**：支持接入 Claude、GPT 等其他 AI 服务
- **多设备支持**：扩展支持天猫精灵、小度音箱
- **技能市场**：支持自定义问答技能

### 未来功能规划

- **情感识别**：根据用户情绪调整回复风格
- **多语言支持**：支持中英文双语对话
- **声纹识别**：识别不同用户，提供个性化响应
- **智能家居控制**：通过对话控制家中的智能设备
