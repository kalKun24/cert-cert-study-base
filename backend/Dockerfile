# マルチステージビルド: Goバイナリをビルドして最小イメージに含める

# ---- ビルドステージ ----
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 依存関係のキャッシュを活用するため、先にgo.mod/go.sumをコピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピーしてビルド
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/server

# ---- 実行ステージ ----
FROM alpine:3.21

# タイムゾーンとCA証明書を追加
RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    echo "Asia/Tokyo" > /etc/timezone

WORKDIR /app

# 非rootユーザーで実行
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

COPY --from=builder /app/server /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
