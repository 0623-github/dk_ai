#!/bin/bash

# 开发启动脚本 - 启动所有服务
# 用法: ./script/dev.sh

set -e

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
FE_DIR="$PROJECT_DIR/fe-react"

echo "========================================"
echo "  AI 聊天助手 - 开发环境启动"
echo "========================================"

# 检查 Ollama
echo ""
echo "[1/3] 检查 Ollama 服务..."
if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "  ✓ Ollama 已在运行"
else
    echo "  → 启动 Ollama 服务..."
    ollama serve > /dev/null 2>&1 &
    sleep 2
    if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo "  ✓ Ollama 已启动"
    else
        echo "  ✗ Ollama 启动失败"
    fi
fi

# 检查模型
echo ""
echo "[2/3] 检查 AI 模型..."
MODEL=$(grep "^model:" "$PROJECT_DIR/conf/openai.yaml" | cut -d'"' -f2)
if ollama list | grep -q "$MODEL"; then
    echo "  ✓ 模型 '$MODEL' 已安装"
else
    echo "  ! 模型 '$MODEL' 未安装，尝试下载..."
    ollama pull "$MODEL"
fi

# 启动后端
echo ""
echo "[3/3] 启动后端服务..."
cd "$PROJECT_DIR"
export GOPROXY=https://goproxy.cn,direct
go run . &
BACKEND_PID=$!
sleep 3
echo "  ✓ 后端服务已启动 (PID: $BACKEND_PID)"

# 启动前端
echo ""
echo "启动前端服务..."
cd "$FE_DIR"
npm run dev -- --host 0.0.0.0 &
FRONTEND_PID=$!
sleep 3
echo "  ✓ 前端服务已启动 (PID: $FRONTEND_PID)"

echo ""
echo "========================================"
echo "  所有服务已启动!"
echo "========================================"
echo ""
echo "  前端:  http://localhost:5173"
echo "  后端:  http://localhost:9090"
echo "  Ollama: localhost:11434"
echo ""
echo "  按 Ctrl+C 停止所有服务"
echo ""

# 等待用户中断
trap "echo ''; echo '正在停止服务...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0" INT TERM

wait
