# 第一阶段：编译
FROM golang:1.26-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server cmd/server/main.go

# 第二阶段：运行
FROM alpine:latest

# 安装时区和证书（防 https 请求报错，且保持时区正确）
RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/configs ./configs

# 创建必要目录
RUN mkdir -p /app/storage/uploads /app/storage/logs

EXPOSE 8080

CMD ["./server"]
