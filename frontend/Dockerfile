# ---- ビルドステージ ----
FROM node:18-alpine AS builder

WORKDIR /app

# 依存関係のインストール（キャッシュ最適化のため package.json を先にコピー）
COPY package.json ./
RUN npm install

# ソースコードをコピーしてビルド
COPY . .
RUN npm run build

# ---- 本番ステージ ----
FROM nginx:alpine AS production

# nginx のデフォルト設定を削除し、カスタム設定を配置
COPY --from=builder /app/dist /usr/share/nginx/html

# SPA 向け nginx 設定（すべてのパスを index.html にフォールバック）
RUN printf 'server {\n\
    listen 80;\n\
    server_name localhost;\n\
    root /usr/share/nginx/html;\n\
    index index.html;\n\
\n\
    # SPA のクライアントサイドルーティング対応\n\
    location / {\n\
        try_files $uri $uri/ /index.html;\n\
    }\n\
\n\
    # API へのプロキシ（本番環境ではバックエンドサービスにルーティング）\n\
    location /api/ {\n\
        proxy_pass http://backend:8080;\n\
        proxy_set_header Host $host;\n\
        proxy_set_header X-Real-IP $remote_addr;\n\
    }\n\
}\n' > /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
