#!/bin/bash
# ==============================================================================
# Godeniter 2.0 跨平台打包与编译脚本
# 能够将静态资源与模板内嵌为单一二进制文件，客户在目标机器上双击即可直接运行
# ==============================================================================

set -e

APP_NAME="godeniter-app"
SRC_PATH="./cmd/server"
OUTPUT_DIR="./dist"

mkdir -p ${OUTPUT_DIR}

echo "=========================================================="
echo " Starting Godeniter Build Process (Zero-Dependency)"
echo "=========================================================="

# 1. 运行所有自动化测试
echo ">> Running unit tests..."
go test ./... -v
echo ">> All tests passed!"

# 2. 编译本机可执行文件
echo ">> Building native binary for local machine..."
go build -ldflags="-s -w" -o ${OUTPUT_DIR}/${APP_NAME} ${SRC_PATH}/main.go
echo "   -> Created: ${OUTPUT_DIR}/${APP_NAME}"

# 3. 交叉编译 Windows 64位可执行文件 (用于交付给客户，双击即用)
echo ">> Cross-compiling for Windows (x86_64)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ${OUTPUT_DIR}/${APP_NAME}.exe ${SRC_PATH}/main.go
echo "   -> Created: ${OUTPUT_DIR}/${APP_NAME}.exe"

# 4. 交叉编译 Linux 64位可执行文件
echo ">> Cross-compiling for Linux (x86_64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ${OUTPUT_DIR}/${APP_NAME}-linux-amd64 ${SRC_PATH}/main.go
echo "   -> Created: ${OUTPUT_DIR}/${APP_NAME}-linux-amd64"

echo "=========================================================="
echo " Build Completed Successfully!"
echo " Deployment files located in: ${OUTPUT_DIR}/"
echo " On Windows: Simply double-click '${APP_NAME}.exe' to run."
echo "=========================================================="
