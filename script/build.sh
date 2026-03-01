#!/bin/bash

# 编译脚本 - 用于构建 Go 项目
set -e

echo "=== 开始编译项目 ==="

# 设置 Go 环境变量
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct

# 项目根目录
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_DIR"

# 创建输出目录
OUTPUT_DIR="$PROJECT_DIR/output"
mkdir -p "$OUTPUT_DIR"

echo "=== 安装 Go 依赖 ==="
go mod tidy
go mod download

echo "=== 编译项目 ==="
APP_NAME="dk_ai"

if [ "$GOOS" == "windows" ]; then
    OUTPUT_BIN="$OUTPUT_DIR/$APP_NAME.exe"
else
    OUTPUT_BIN="$OUTPUT_DIR/$APP_NAME"
fi

go build -o "$OUTPUT_BIN"

echo "=== 复制配置文件 ==="
mkdir -p "$OUTPUT_DIR/conf"
cp -r conf/* "$OUTPUT_DIR/conf/"

echo "=== 构建前端 ==="
cd "$PROJECT_DIR/fe-react"
npm install
npm run build

echo "=== 复制前端文件 ==="
mkdir -p "$OUTPUT_DIR/fe"
cp -r dist/* "$OUTPUT_DIR/fe/"
cp -r "$PROJECT_DIR/fe-react/index.html" "$OUTPUT_DIR/fe/"

echo "=== 编译完成 ==="
echo "输出文件: $OUTPUT_BIN"
echo "配置文件目录: $OUTPUT_DIR/conf/"
echo "前端文件目录: $OUTPUT_DIR/fe/"
echo "编译脚本执行完毕！"
