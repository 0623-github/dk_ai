# dk_ai 部署指南

**服务器**: 腾讯云 106.54.52.111  
**部署路径**: `/root/dk_ai`  
**部署日期**: 2026-03-20  
**状态**: ✅ 后端服务已部署  

---

## 访问地址

| 服务 | 地址 | 状态 |
|------|------|------|
| 后端 API | http://106.54.52.111:9090 | ✅ 运行中 |
| 前端开发 | http://localhost:5173 | 本地运行 |

---

## 部署状态

| 步骤 | 状态 | 说明 |
|------|------|------|
| 代码上传 | ✅ | 已上传到 /root/dk_ai |
| 依赖下载 | ✅ | go mod download 完成 |
| 编译构建 | ✅ | 可执行文件 23MB |
| 服务启动 | ✅ | 监听 0.0.0.0:9090 |
| 防火墙配置 | ✅ | 端口 9090 已开放 |
| Ollama 连接 | ⚠️ | 响应较慢，已降级到模拟模式 |
| 前端部署 | ⏳ | 待完成 |

---

## API 测试

```bash
# 健康检查
curl http://106.54.52.111:9090/ping
# {"message":"pong"}

# 创建会话
curl -X POST http://106.54.52.111:9090/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"title":"测试会话"}'

# 列出会话
curl http://106.54.52.111:9090/api/sessions

# 发送消息（当前为模拟模式）
curl -X POST http://106.54.52.111:9090/chat \
  -H "Content-Type: application/json" \
  -d '{"session_id":"xxx","message":"你好"}'
```

---

## 服务器上的模型

Ollama 已安装，可用模型：
- `gemma:2b` (3B参数)
- `llama3:latest` (8B参数)
- `llama3.2:latest` (3.2B参数)

当前 Ollama 响应较慢，服务已自动降级到模拟模式。

---

## 服务管理

```bash
# SSH 登录
ssh root@106.54.52.111

# 查看服务状态
cd /root/dk_ai
ps aux | grep dk_ai

# 查看日志
tail -f /tmp/dk_ai.log

# 重启服务
pkill dk_ai
./dk_ai > /tmp/dk_ai.log 2>&1 &

# 检查 Ollama
curl http://localhost:11434/api/tags
```

---

## 已知问题

1. **Ollama 响应慢**: 服务器上的 Ollama 生成回复较慢，已自动降级到模拟模式
2. **前端未部署**: 目前只有后端 API，前端需在本地运行

---

## 下一步

- [ ] 优化 Ollama 配置或使用更轻量模型
- [ ] 前端打包部署到服务器
- [ ] 配置域名和 HTTPS

---

**李一一** 🐾  
2026-03-20
