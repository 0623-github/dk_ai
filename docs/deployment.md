# dk_ai 部署指南

**服务器**: 腾讯云 106.54.52.111  
**部署路径**: `/root/dk_ai`  
**部署日期**: 2026-03-20  

---

## 部署状态

| 步骤 | 状态 | 说明 |
|------|------|------|
| 代码上传 | ✅ | 已上传到 /root/dk_ai |
| 依赖下载 | 🔄 | go mod download 进行中 |
| 编译构建 | ⏳ | 等待依赖完成 |
| 服务启动 | ⏳ | 等待编译完成 |
| 前端部署 | ⏳ | 待定 |

---

## 部署步骤

### 1. 后端部署

```bash
# 登录服务器
ssh root@106.54.52.111

# 进入项目目录
cd /root/dk_ai

# 下载依赖
go mod download

# 编译
go build -o dk_ai .

# 创建数据目录
mkdir -p data

# 启动服务
./dk_ai > /var/log/dk_ai.log 2>&1 &
```

### 2. 验证后端

```bash
curl http://localhost:9090/ping
# 期望返回: {"message":"pong"}
```

### 3. 前端部署（可选）

方案A: 本地开发（当前）
- 前端运行在 localhost:5173
- 通过 SSH 隧道访问

方案B: 服务器部署（生产）
```bash
cd /root/dk_ai/fe-react
npm install
npm run build
# 将 dist 目录复制到后端静态文件目录
```

---

## 服务管理

```bash
# 查看服务状态
ps aux | grep dk_ai

# 查看日志
tail -f /var/log/dk_ai.log

# 停止服务
pkill dk_ai

# 重启服务
pkill dk_ai
cd /root/dk_ai && ./dk_ai > /var/log/dk_ai.log 2>&1 &
```

---

## 配置说明

### 后端配置
- 监听端口: `0.0.0.0:9090`
- 数据库: `data/dk_ai.db` (SQLite)
- AI 配置: `conf/openai.yaml`

### AI 配置示例
```yaml
model: "gemma:2b"
baseURL: "http://localhost:11434/v1"
```

---

## 访问方式

| 服务 | 地址 | 说明 |
|------|------|------|
| 后端 API | http://106.54.52.111:9090 | 需开放防火墙端口 |
| 前端开发 | http://localhost:5173 | 本地开发 |

---

## 防火墙配置

如需外网访问后端，需开放 9090 端口：

```bash
# 腾讯云控制台安全组配置
# 入站规则: 允许 TCP 9090 端口
```

---

**李一一** 🐾  
2026-03-20
