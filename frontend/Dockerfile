FROM nginx:alpine

# 直接拷贝本地已经由 Mac 打包构建好的静态 dist 目录
COPY dist /usr/share/nginx/html

# 拷贝 nginx 配置文件
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
