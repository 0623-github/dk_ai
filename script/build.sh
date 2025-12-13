#!/bin/bash

# 编译脚本 - 用于构建Go项目
set -e

echo "=== 开始编译项目 ==="

# 设置Go环境变量
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct

# 项目根目录
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_DIR"

# 创建输出目录
OUTPUT_DIR="$PROJECT_DIR/output"
mkdir -p "$OUTPUT_DIR"

echo "=== 安装依赖 ==="
go mod tidy
go mod download

echo "=== 编译项目 ==="
# 获取项目名称
APP_NAME="dk_ai"

# 根据操作系统设置输出文件名
if [ "$GOOS" == "windows" ]; then
    OUTPUT_BIN="$OUTPUT_DIR/$APP_NAME.exe"
else
    OUTPUT_BIN="$OUTPUT_DIR/$APP_NAME"
fi

# 编译
go build -o "$OUTPUT_BIN"

echo "=== 复制配置文件 ==="
# 复制配置文件到输出目录
mkdir -p "$OUTPUT_DIR/conf"
cp -r conf/* "$OUTPUT_DIR/conf/"

echo "=== 编译完成 ==="
echo "输出文件: $OUTPUT_BIN"
echo "配置文件目录: $OUTPUT_DIR/conf/"
echo "编译脚本执行完毕！"
