FROM alpine:latest

# 安装时区和证书（防 https 请求报错，且保持时区正确）
RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 直接复制本地已经由 Mac 交叉编译好的 Linux 二进制文件
COPY server .
COPY configs ./configs

# 创建必要目录
RUN mkdir -p /app/storage/uploads /app/storage/logs

EXPOSE 8080

CMD ["./server"]
