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

# ビルド成果物をコピー
COPY --from=builder /app/dist /usr/share/nginx/html

# nginx 設定テンプレートを /tmp に配置する。
# /etc/nginx/templates/ に置くと docker-entrypoint.sh が全環境変数を展開してしまい
# $uri / $host 等の nginx 変数が破壊されるため、意図的にそのディレクトリを避ける。
COPY nginx.conf.template /tmp/nginx.conf.template

EXPOSE 8080

# 起動時に PORT と BACKEND_URL のみを展開し、nginx 変数（$uri/$host 等）は保護する。
CMD ["/bin/sh", "-c", "RESOLVER=${RESOLVER:-8.8.8.8}; envsubst '${PORT} ${BACKEND_URL} ${BACKEND_HOST} ${RESOLVER}' < /tmp/nginx.conf.template > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"]
