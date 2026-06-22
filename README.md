# cert-study-base

> セキュリティ資格勉強会（CISSP・情報処理安全確保支援士など）を効率的に運営・管理するための Web アプリケーション

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://react.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?logo=typescript)](https://www.typescriptlang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/kalKun24/cert-study-base/actions/workflows/ci.yml/badge.svg)](https://github.com/kalKun24/cert-study-base/actions/workflows/ci.yml)

---

## プロジェクト概要

勉強会メンバーが資格試験の問題・解答・解説・議論点メモを Markdown 形式で投稿・共有できるプラットフォームです。チーム単位でコンテンツを管理し、タグによる分類・検索・フィルタリングで知識の蓄積と再利用を支援します。

---

## 機能一覧

### 問題管理
- Markdown エディタによる問題・解答・解説・議論点メモの作成・編集
- 公開ステータス管理（draft / private / published）
- タグ付け（フラット・複数付与可）によるカテゴリ分類
- タグ絞り込み・キーワード検索・ページネーション対応の一覧表示
- 問題へのコメント投稿・編集・削除

### ノート管理
- Markdown 形式の自由記述ノート（本文・議論点・メモ）の作成・編集
- 問題管理と同等の公開ステータス・タグ・検索機能
- ノートへのコメント機能

### チーム管理
- チームの作成・更新・削除
- メンバー招待（UUID / ユーザー名 / メールアドレスで指定）・承諾・拒否
- チーム内ロール管理（owner / member）
- メンバーごとの活動統計（問題数・コメント数・最終ログイン日時）

### 認証・ユーザー管理
- ID / パスワード認証（JWT 発行）
- ロールベースアクセス制御（admin / teamowner / user）
- admin によるユーザー作成・停止・チームオーナー権限の付与
- 本人によるプロフィール編集・パスワード変更

---

## 技術スタック

| レイヤー | 技術 |
|---|---|
| バックエンド | Go 1.25 |
| フロントエンド | React 18 / TypeScript 5.6 / Vite |
| API 設計 | REST API / OpenAPI 3.0 |
| 永続化 | Google Cloud Firestore |
| ストレージ | Google Cloud Storage（GCS） |
| 認証 | JWT（`golang-jwt/jwt`）/ bcrypt |
| インフラ | GCP Cloud Run |
| CI/CD | GitHub Actions（PR・main マージトリガー） |

---

## 前提条件

ローカル開発を始める前に以下のツールをインストールしてください。

| ツール | 推奨バージョン | 用途 |
|---|---|---|
| [Docker](https://docs.docker.com/get-docker/) | 24 以上 | `make up` によるサービス一括起動 |
| [Docker Compose](https://docs.docker.com/compose/install/) | v2.x | コンテナオーケストレーション |
| [Go](https://go.dev/dl/) | 1.25 | バックエンドのビルド・テスト（ローカル実行時のみ） |
| [Node.js](https://nodejs.org/) | 18 LTS | フロントエンドの依存インストール（ローカル実行時のみ） |

> `make up` を使う場合、Go・Node.js のローカルインストールは必須ではありません。テスト（`make test`）やリント（`make lint`）を実行する際は必要です。

---

## ローカル環境セットアップ

### ステップ 1: リポジトリをクローンする

```bash
git clone https://github.com/kalKun24/cert-study-base.git
cd cert-study-base
```

### ステップ 2: git フックを有効化する

初回クローン後に一度だけ実行してください。コミット時に `gofmt` が自動で走り、未フォーマットのコードをステージング前に修正します。

```bash
make hooks
```

### ステップ 3: 環境変数ファイルを作成する

```bash
cp .env.example .env
```

`.env` を編集して各値を設定してください（`.env` は `.gitignore` により git 管理外です）。

| 変数 | 必須 | 説明 |
|---|---|---|
| `PORT` | - | バックエンドのポート番号（デフォルト: `8080`） |
| `JWT_SECRET` | 必須 | JWT 署名用シークレットキー（下記コマンドで生成） |
| `GCP_PROJECT_ID` | 必須 | GCP プロジェクト ID（Firestore 接続に必要） |
| `GCS_BUCKET` | 必須 | GCS バケット名 |
| `SEED_ADMIN_USERNAME` | 推奨 | 初回起動時に自動作成する admin のユーザー名 |
| `SEED_ADMIN_PASSWORD` | 推奨 | 初回起動時に自動作成する admin のパスワード |
| `SEED_ADMIN_EMAIL` | 推奨 | 初回起動時に自動作成する admin のメールアドレス |
| `SEED_ADMIN_DISPLAY_NAME` | - | 初回起動時に自動作成する admin の表示名（デフォルト: `Administrator`） |

`JWT_SECRET` の生成例:

```bash
openssl rand -base64 32
```

> **セキュリティ上の注意**: `.env` ファイルには実際のシークレット（JWT_SECRET、パスワードなど）が含まれます。
> `.gitignore` によって git 管理外になっていることを必ず確認し、絶対にリポジトリにコミットしないでください。
> 本番環境では環境変数を GCP Secret Manager で管理し、Cloud Run に注入してください。

### ステップ 4: アプリケーションを起動する

```bash
make up
```

バックエンド・フロントエンド・Firestore エミュレータの 3 サービスが Docker コンテナとして起動します。初回はイメージのビルドが発生するため数分かかります。

起動後、以下の URL でアクセスできます。

| サービス | URL |
|---|---|
| フロントエンド | http://localhost:3000 |
| バックエンド API | http://localhost:8080 |
| ヘルスチェック | http://localhost:8080/health |

### ステップ 5: 初回 admin ユーザーを確認する

`.env` に `SEED_ADMIN_*` が設定されている場合、**ユーザーストアが空のときのみ** サーバー起動時に admin ユーザーが自動作成されます。

```bash
# ログで確認する
docker-compose logs backend | grep "admin"
```

成功すると以下のようなログが出力されます。

```
INFO 初回 admin ユーザーを作成しました username=admin email=admin@example.com
```

> **注意**: admin ユーザー作成後は `.env` から `SEED_ADMIN_PASSWORD` を削除することを推奨します。

---

## make コマンド一覧

| コマンド | 内容 |
|---|---|
| `make up` | バックエンド・フロントエンド・Firestore エミュレータをまとめて起動（Docker Compose） |
| `make down` | 全サービスを停止 |
| `make test` | バックエンドの全テストを実行（`go test ./...`） |
| `make lint` | `golangci-lint` を実行（未インストールの場合は自動インストール） |
| `make fmt` | `gofmt` でバックエンドのコードをフォーマット |
| `make hooks` | git フックを有効化（初回クローン後に一度だけ実行） |
| `make swagger` | Swagger UI を起動（実装予定） |
| `make build` | 本番用 Docker イメージをビルド |

---

## ディレクトリ構成

```
.
├── backend/
│   ├── cmd/                     # エントリポイント（main.go）
│   └── internal/
│       ├── domain/              # エンティティ層: ビジネスエンティティ・ルール（依存なし）
│       ├── usecase/             # ユースケース層: ビジネスロジック（domain のみ依存可）
│       ├── interface/           # インターフェース層: ハンドラ・DTO・Repository インターフェース
│       └── infrastructure/      # インフラ層: Firestore・GCS・認証・ルーティングの具体実装
├── frontend/
│   └── src/                     # React / TypeScript ソースコード
├── api/
│   └── openapi.yaml             # OpenAPI 3.0 仕様書（API 定義の単一管理元）
├── scripts/                     # シードデータ生成スクリプト
├── tickets/                     # チケット管理（backlog/ in-progress/ done/）
├── .github/workflows/
│   └── ci.yml                   # CI パイプライン（テスト・リント）
├── docker-compose.yml
├── Makefile
└── CLAUDE.md                    # AI アシスタント向け開発ガイドライン
```

---

## アーキテクチャ概要

クリーンアーキテクチャを採用しています。依存の方向は **外側から内側への一方向のみ** 許可します。

```
[Infrastructure] --> [Interface] --> [Usecase] --> [Domain]
```

| 層 | 役割 |
|---|---|
| Domain | ビジネスエンティティとルールを定義。外部への依存を持たない |
| Usecase | ビジネスロジックを実装。Domain のみに依存 |
| Interface | HTTP ハンドラ・DTO・Repository インターフェースを定義 |
| Infrastructure | Firestore・GCS・JWT 認証などの具体実装。Interface のインターフェースを実装（DI） |

層の境界は interface で定義し、具体実装は Infrastructure 層に置きます。これにより、実装の差し替え（テスト時のモック化など）が容易になります。

---

## API ドキュメント

API の仕様は `api/openapi.yaml` で一元管理しています（API First 原則）。

主なリソースエンドポイント:

| リソース | ベースパス |
|---|---|
| 認証 | `/api/v1/auth` |
| ユーザー管理 | `/api/v1/users` |
| チーム管理 | `/api/v1/teams` |
| 問題管理 | `/api/v1/teams/{team_id}/questions` |
| ノート管理 | `/api/v1/teams/{team_id}/notes` |
| タグ管理 | `/api/v1/teams/{team_id}/tags` |
| 招待 | `/api/v1/invitations` |

すべてのレスポンスは以下の統一フォーマットで返します。

```json
{
  "data": { ... },
  "error": null
}
```

認証が必要なエンドポイントには `Authorization: Bearer <token>` ヘッダーを付与してください。ログイン（`POST /api/v1/auth/login`）で取得したトークンを使用します。

---

## 本番環境（Cloud Run）へのデプロイ

CI/CD は GitHub Actions で管理しており、`main` ブランチへのマージをトリガーにテストが実行されます。デプロイは Cloud Run を対象としています。

### 初回デプロイ時の admin ユーザー作成

```bash
gcloud run deploy cert-study-backend \
  --set-env-vars SEED_ADMIN_USERNAME=admin \
  --set-env-vars SEED_ADMIN_EMAIL=admin@example.com \
  --update-secrets SEED_ADMIN_PASSWORD=seed-admin-password:latest \
  --update-secrets JWT_SECRET=jwt-secret-key:latest \
  --update-secrets GCS_BUCKET=gcs-bucket-name:latest
```

admin ユーザー作成後は `SEED_ADMIN_PASSWORD` を環境変数から削除してください。

```bash
gcloud run services update cert-study-backend \
  --remove-env-vars SEED_ADMIN_PASSWORD
```

本番環境のシークレット（`JWT_SECRET`・`SEED_ADMIN_PASSWORD` など）は **GCP Secret Manager** で管理します。

---

## コントリビューション

### ブランチ戦略

`main` ブランチへの直接 push は禁止です。必ず feature ブランチを作成して PR を出してください。

```bash
git checkout -b feature/your-feature-name
```

### PR ルール

- PR タイトルはコミットメッセージ規約（Semantic Commit Messages）に準拠してください
- `main` への PR には **1 名の承認が必須**です（セルフマージ禁止）
- コミットメッセージの形式: `<type>(<scope>): <件名（日本語）>`

| type | 用途 |
|---|---|
| `feat` | 新機能追加 |
| `fix` | バグ修正 |
| `docs` | ドキュメント変更 |
| `refactor` | リファクタリング |
| `test` | テストの追加・修正 |
| `chore` | ビルドツール・タスクランナーの更新 |

例: `feat(question): 問題作成 API を追加`

### 開発ガイドライン

- 新しい API を追加・変更する場合は `api/openapi.yaml` を先に更新してから実装してください（API First 原則）
- 新機能追加時はユニットテストを合わせて作成してください（特にユースケース層は必須）
- コードは `gofmt` / `golangci-lint` を通してからコミットしてください

---

## ライセンス

MIT © [kalKun24](https://github.com/kalKun24)
