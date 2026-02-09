#!/bin/bash

# 启动脚本 - 用于启动编译后的Go应用
set -e

echo "=== 开始启动服务 ==="

# 项目根目录
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUTPUT_DIR="$PROJECT_DIR/output"
APP_NAME="dk_ai"

# 根据操作系统设置二进制文件名
if [ "$OSTYPE" == "msys" ] || [ "$OSTYPE" == "win32" ]; then
    BINARY_PATH="$OUTPUT_DIR/$APP_NAME.exe"
else
    BINARY_PATH="$OUTPUT_DIR/$APP_NAME"
fi

# 检查二进制文件是否存在
if [ ! -f "$BINARY_PATH" ]; then
    echo "错误: 二进制文件不存在，请先运行编译脚本"
    echo "执行: bash script/build.sh"
    exit 1
fi

# 检查配置文件目录
CONFIG_DIR="$OUTPUT_DIR/conf"
if [ ! -d "$CONFIG_DIR" ]; then
    echo "错误: 配置文件目录不存在，请先运行编译脚本"
    exit 1
fi

# 设置环境变量
export GO_ENV=production
export APP_HOME="$OUTPUT_DIR"
export CONFIG_PATH="$CONFIG_DIR"

echo "=== 服务信息 ==="
echo "应用路径: $BINARY_PATH"
echo "配置目录: $CONFIG_DIR"
echo "环境变量: GO_ENV=$GO_ENV"

# 启动应用
echo "=== 启动应用 ==="
cd "$OUTPUT_DIR"

# 如果是开发环境，可以使用以下方式运行(开发模式)
# go run "$PROJECT_DIR/main.go"

# 生产环境运行编译后的二进制文件
"$BINARY_PATH"

# 注意：如果需要后台运行，可以使用以下命令
# nohup "$BINARY_PATH" > app.log 2>&1 &
echo "=== 服务启动完成 ==="
