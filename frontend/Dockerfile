# 第一阶段：编译
FROM node:20-alpine AS builder

WORKDIR /app

# 复制 package.json 和 package-lock.json 安装依赖
COPY package*.json ./
RUN npm install

# 复制源码进行打包
COPY . .
RUN npm run build

# 第二阶段：运行
FROM nginx:alpine

# 拷贝打包好的静态文件
COPY --from=builder /app/dist /usr/share/nginx/html

# 拷贝 nginx 配置文件
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
