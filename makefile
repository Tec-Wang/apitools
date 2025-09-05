# Makefile for apitools project
# 支持多平台构建的Go项目

# 变量定义
GO_VERSION := 1.23.11
OUTPUT_DIR := output
API_DIR := api

# 构建目标
.PHONY: all build clean help env info

# 默认目标
all: build

# 显示帮助信息
help:
	@echo "可用的构建命令:"
	@echo "  make build     - 构建所有平台的可执行文件"
	@echo "  make build-api - 只构建API服务"
	@echo "  make env       - 显示Go环境信息"
	@echo "  make run       - 运行服务"

# 显示Go环境信息
env:
	@echo "=== Go环境信息 ==="
	@echo "GOROOT: $$(go env GOROOT)"
	@echo "GOPATH: $$(go env GOPATH)"
	@echo "GOMODCACHE: $$(go env GOMODCACHE)"
	@go version
	@go env GOPROXY
	@go env GOSUMDB


# 创建输出目录
$(OUTPUT_DIR):
	@mkdir -p $(OUTPUT_DIR)

# 构建所有服务
build: $(OUTPUT_DIR) build-api
	@echo "=== 构建完成，查看输出文件 ==="
	@chmod +x $(OUTPUT_DIR)/*-linux-*
	@ls -la $(OUTPUT_DIR)/

# 构建API服务
build-api: $(OUTPUT_DIR)
	@echo "=== 开始构建API服务... ==="
	@cd $(API_DIR) && \
		echo "API服务依赖包:" && \
		go list -m all | head -20 || echo "无法获取依赖包列表" && \
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o ../$(OUTPUT_DIR)/api-server-linux-amd64 tools.go
	@echo "API服务构建完成"

# 运行生产环境服务
run:
	@echo "=== 启动生产服务 ==="
	@echo "创建必要目录..."
	@mkdir -p logs etc
	@echo "启动API服务..."
	@nohup ./api-server-linux-amd64 -f etc/api.yaml > logs/api.log 2>&1 & echo "API服务已启动，PID: $$!"
	@echo "服务启动完成，查看日志:"
	@echo "API日志: tail -f logs/api.log"

# 部署服务（停止旧服务 + 启动新服务）
deploy:
	@echo "=== 开始部署 ==="
	@echo "1. 停止现有服务..."
	@echo "停止api-server进程..."
	@killall api-server-linux-amd64 2>/dev/null || echo "api-server进程已停止或不存在"
	@echo "停止rpc-server进程..."
	@killall rpc-server-linux-amd64 2>/dev/null || echo "rpc-server进程已停止或不存在"
	@sleep 2
	@echo "2. 启动新服务..."
	@chmod +x api-server-linux-amd64
	@$(MAKE) run

gen-api:
	@goctl api go -api api/tools.api -dir api

