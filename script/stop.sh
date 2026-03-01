#!/bin/bash

# 停止所有服务脚本
# 用法: ./script/stop.sh

echo "========================================"
echo "  停止所有服务"
echo "========================================"

# 停止前端 (Vite)
echo ""
echo "[1/3] 停止前端服务..."
if lsof -ti:5173 > /dev/null 2>&1; then
    lsof -ti:5173 | xargs kill -9 2>/dev/null || true
    echo "  ✓ 前端服务已停止"
else
    echo "  ✓ 前端服务未运行"
fi

# 停止后端 (Go)
echo ""
echo "[2/3] 停止后端服务..."
if lsof -ti:9090 > /dev/null 2>&1; then
    lsof -ti:9090 | xargs kill -9 2>/dev/null || true
    echo "  ✓ 后端服务已停止"
else
    echo "  ✓ 后端服务未运行"
fi

# 停止 Ollama
echo ""
echo "[3/3] 停止 Ollama 服务..."
if pgrep -f "ollama serve" > /dev/null 2>&1; then
    pkill -f "ollama serve" 2>/dev/null || true
    echo "  ✓ Ollama 服务已停止"
else
    echo "  ✓ Ollama 服务未运行"
fi

echo ""
echo "========================================"
echo "  所有服务已停止!"
echo "========================================"
