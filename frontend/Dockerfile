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

# nginx 設定テンプレートをコピー（起動時に envsubst で展開される）
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

EXPOSE 8080

# nginx:alpine の docker-entrypoint.sh は /etc/nginx/templates/*.template を
# envsubst で展開して /etc/nginx/conf.d/ に配置する。
# デフォルトでは全環境変数を展開してしまい $uri/$host 等の nginx 変数も破壊される。
# NGINX_ENVSUBST_TEMPLATE_SUFFIX を空にして自動展開を無効化し、
# CMD で明示的に展開対象変数（PORT と BACKEND_URL のみ）を指定する。
CMD ["/bin/sh", "-c", "envsubst '${PORT} ${BACKEND_URL}' < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"]
