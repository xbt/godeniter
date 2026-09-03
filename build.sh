#!/bin/bash
# ==============================================================================
# Godeniter 2.0 跨平台打包与一键编译脚本
# 编译所有官方示例与 CLI 脚手架工具，支持生成 Windows .exe / Linux / macOS 二进制
# ==============================================================================

set -e

OUTPUT_DIR="./dist"
mkdir -p ${OUTPUT_DIR}

echo "=========================================================="
echo " Starting Godeniter 2.0 Full Build Process"
echo "=========================================================="

# 1. 运行全局单元测试
echo ">> Running all unit tests..."
go test ./... -v
echo ">> All tests passed!"

# 2. 编译 CLI 脚手架工具 (cmd/godeniter)
echo ">> Building CLI Tool (cmd/godeniter)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ${OUTPUT_DIR}/godeniter-cli.exe ./cmd/godeniter/main.go
go build -ldflags="-s -w" -o ${OUTPUT_DIR}/godeniter ./cmd/godeniter/main.go
echo "   -> Created: ${OUTPUT_DIR}/godeniter-cli.exe & ${OUTPUT_DIR}/godeniter"

# 3. 编译 Demo 1: 前后端分离 API + SPA 模式
echo ">> Building Demo 1 (API + SPA)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ${OUTPUT_DIR}/demo1_api_spa.exe ./examples/01_api_spa/main.go
go build -ldflags="-s -w" -o ${OUTPUT_DIR}/demo1_api_spa ./examples/01_api_spa/main.go
echo "   -> Created: ${OUTPUT_DIR}/demo1_api_spa.exe & ${OUTPUT_DIR}/demo1_api_spa"

# 4. 编译 Demo 2: 经典 PHP 风格模板渲染 MVC 模式
echo ">> Building Demo 2 (MVC + SSR Template)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ${OUTPUT_DIR}/demo2_mvc_template.exe ./examples/02_mvc_template/main.go
go build -ldflags="-s -w" -o ${OUTPUT_DIR}/demo2_mvc_template ./examples/02_mvc_template/main.go
echo "   -> Created: ${OUTPUT_DIR}/demo2_mvc_template.exe & ${OUTPUT_DIR}/demo2_mvc_template"

echo "=========================================================="
echo " Build Completed Successfully!"
echo " All binaries generated in: ${OUTPUT_DIR}/"
echo "   - dist/godeniter (CLI 脚手架工具)"
echo "   - dist/demo1_api_spa.exe (API + SPA 模式单文件)"
echo "   - dist/demo2_mvc_template.exe (经典 MVC 模式单文件)"
echo "=========================================================="
