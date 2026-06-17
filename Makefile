# プロジェクト共通Makefile
# バックエンド・フロントエンドをまとめて操作するコマンドを定義

.PHONY: up down test lint swagger build

GOLANGCI_LINT_VERSION := v1.62.2
BACKEND_DIR := ./backend

## up: バックエンド・フロントエンドをまとめて起動
up:
	@echo "サーバーを起動します..."
	cd $(BACKEND_DIR) && go run ./cmd/server

## down: 全サービスを停止（将来Docker Compose導入時に拡張）
down:
	@echo "サービスを停止します..."
	@pkill -f "go run ./cmd/server" 2>/dev/null || true

## test: 全テストを実行
test:
	@echo "テストを実行します..."
	cd $(BACKEND_DIR) && go test ./... -v

## lint: golangci-lintを実行
lint:
	@echo "Lintを実行します..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint をインストールします..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi
	cd $(BACKEND_DIR) && golangci-lint run ./...

## swagger: Swagger UIを起動・openapi.yamlを反映（将来実装）
swagger:
	@echo "Swagger UIを起動します..."
	@echo "TODO: swaggo/swagのセットアップ後に実装"

## build: 本番用Dockerイメージをビルド
build:
	@echo "Dockerイメージをビルドします..."
	docker build -t cert-study-base-backend:latest $(BACKEND_DIR)
