# ---- 构建阶段 ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 用国内代理加速依赖拉取（微信云托管在国内节点构建）
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /munch-server ./cmd/api

# ---- 运行阶段 ----
FROM alpine:3.20

WORKDIR /app

# 时区 + 证书（HTTPS 调用微信接口需要 ca-certificates）
RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder /munch-server /app/munch-server

# 微信云托管默认要求监听 80
ENV PORT=80
EXPOSE 80

CMD ["/app/munch-server"]
