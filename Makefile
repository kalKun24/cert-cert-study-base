# プロジェクト共通Makefile
# バックエンド・フロントエンドをまとめて操作するコマンドを定義

.PHONY: up down test lint swagger build

GOLANGCI_LINT_VERSION := v1.62.2
BACKEND_DIR := ./backend

## up: バックエンド・フロントエンドをまとめて起動（Docker Compose）
up:
	@echo "サーバーを起動します..."
	docker-compose up -d

## down: 全サービスを停止（Docker Compose）
down:
	@echo "サービスを停止します..."
	docker-compose down

## test: 全テストを実行
test:
	@echo "テストを実行します..."
	cd $(BACKEND_DIR) && go test ./... -v

## lint: golangci-lintを実行
lint:
	@echo "Lintを実行します..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint をインストールします..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi
	cd $(BACKEND_DIR) && golangci-lint run ./...

## swagger: Swagger UIを起動・openapi.yamlを反映（将来実装）
swagger:
	@echo "Swagger UIを起動します..."
	@echo "TODO: swaggo/swagのセットアップ後に実装"

## build: バックエンド・フロントエンドのDockerイメージをビルド
build:
	@echo "Dockerイメージをビルドします..."
	docker-compose build
