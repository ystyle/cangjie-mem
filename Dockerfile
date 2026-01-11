# 构建阶段
FROM golang:1.23-alpine AS builder

# 安装构建依赖（SQLite 需要 CGO 和 gcc）
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# 设置工作目录
WORKDIR /build

# 设置 Go 环境变量（使用国内代理加速）
ENV GOPROXY=https://goproxy.cn,direct \
    GO111MODULE=on

# 复制源代码
COPY . .

# 下载依赖并整理 go.mod
RUN go mod download && go mod tidy

# 构建应用（启用 FTS5）
RUN CGO_ENABLED=1 GOOS=linux go build \
    -tags sqlite_fts5 \
    -ldflags="-s -w" \
    -o cangjie-mem ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache sqlite-libs ca-certificates

# 创建非 root 用户
RUN addgroup -g 1000 cangjie && \
    adduser -D -u 1000 -G cangjie cangjie

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/cangjie-mem .

# 创建数据目录
RUN mkdir -p /home/cangjie/.cangjie-mem && \
    chown -R cangjie:cangjie /home/cangjie

# 切换到非 root 用户
USER cangjie

# 环境变量
ENV DB_PATH=/home/cangjie/.cangjie-mem/memory.db

# 暴露端口
EXPOSE 8080

# 默认启动 HTTP 服务器（可覆盖）
CMD ["./cangjie-mem", "-http", "-addr", ":8080"]

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/mcp || exit 1
