# ---- ビルドステージ ----
FROM node:18-alpine AS builder

WORKDIR /app

# 依存関係のインストール（キャッシュ最適化のため package.json を先にコピー）
COPY package.json package-lock.json ./
RUN npm ci

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
    # セキュリティ HTTP ヘッダー\n\
    add_header X-Content-Type-Options "nosniff" always;\n\
    add_header X-Frame-Options "DENY" always;\n\
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;\n\
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;\n\
\n\
    # SPA のクライアントサイドルーティング対応\n\
    location / {\n\
        try_files $uri $uri/ /index.html;\n\
    }\n\
\n\
    # API へのプロキシ（本番環境ではバックエンドサービスにルーティング）\n\
    # resolver を指定し変数経由にすることで起動時の名前解決エラーを回避\n\
    resolver 127.0.0.11 valid=30s;\n\
    location /api/ {\n\
        set $backend_host backend;\n\
        proxy_pass http://$backend_host:8080;\n\
        proxy_set_header Host $host;\n\
        proxy_set_header X-Real-IP $remote_addr;\n\
    }\n\
}\n' > /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
