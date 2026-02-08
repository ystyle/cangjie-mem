.PHONY: help build build-frontend build-go test run docker-build docker-run clean

# 默认目标
.DEFAULT_GOAL := help

# 变量
BINARY_NAME=cangjie-mem
GO_FILES=$(shell find . -name '*.go' -type f)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

help: ## 显示帮助信息
	@echo "cangjie-mem 构建脚本"
	@echo ""
	@echo "使用方法:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: build-frontend build-go ## 构建所有（前端 + 后端）

build-frontend: ## 构建前端
	@echo "构建前端..."
	cd web && npm install && npm run build

build-go: ## 构建后端
	@echo "构建后端..."
	CGO_ENABLED=1 go build $(LDFLAGS) -tags sqlite_fts5 -o $(BINARY_NAME) ./cmd/server

test: ## 运行测试
	go test -v -tags sqlite_fts5 ./...

run: ## 运行服务器（开发模式）
	@echo "启动开发服务器..."
	@make -j2 run-frontend run-go

run-frontend: ## 运行前端开发服务器
	cd web && npm run dev

run-go: ## 运行后端服务器
	./$(BINARY_NAME) -http -api -ui

docker-build: ## 构建 Docker 镜像
	docker build -t cangjie-mem:$(VERSION) .

docker-run: ## 运行 Docker 容器
	docker-compose up -d

docker-logs: ## 查看 Docker 日志
	docker-compose logs -f

docker-stop: ## 停止 Docker 容器
	docker-compose down

clean: ## 清理构建产物
	rm -f $(BINARY_NAME)
	rm -rf web/dist
	rm -rf web/node_modules

deps-go: ## 更新 Go 依赖
	go mod tidy
	go mod download

deps-frontend: ## 更新前端依赖
	cd web && npm install

lint: ## 代码检查
	golangci-lint run
	cd web && npm run lint

fmt: ## 格式化代码
	go fmt ./...
	goimports -w .
	cd web && npm run format
