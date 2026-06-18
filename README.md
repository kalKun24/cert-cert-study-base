# cert-study-base

CISSPや情報処理安全確保支援士などのセキュリティ資格取得を目標とした勉強会を、効率的に運営・管理するためのWebアプリケーションです。

---

## ローカル開発環境のセットアップ

### 1. git フックを有効化（初回クローン後に一度だけ実行）

```bash
make hooks
```

コミット時に `gofmt` が自動で走り、未フォーマットのファイルを修正して再ステージします。

### 2. 環境変数ファイルを作成

```bash
cp .env.example .env
```

`.env` を編集して各値を設定してください（`.env` は git 管理外です）:

| 変数 | 説明 |
|---|---|
| `JWT_SECRET` | JWT署名用シークレットキー（`openssl rand -base64 32` で生成）|
| `GCS_BUCKET` | GCSバケット名 |
| `SEED_ADMIN_USERNAME` | 初回起動時に作成する admin のユーザー名 |
| `SEED_ADMIN_PASSWORD` | 初回起動時に作成する admin のパスワード |
| `SEED_ADMIN_EMAIL` | 初回起動時に作成する admin のメールアドレス |
| `SEED_ADMIN_DISPLAY_NAME` | 初回起動時に作成する admin の表示名（省略時: `Administrator`）|

### 2. アプリを起動

```bash
make up
```

バックエンド（:8080）、フロントエンド（:3000）、GCS エミュレータ（:4443）がまとめて起動します。

> **GCS エミュレータについて**: `make up` では自動的に `fake-gcs-server` がローカル GCS として起動します。実 GCS アカウントは不要です。バケット（`local-bucket`）はサーバー起動時に自動作成されます。

### 3. 初回 admin ユーザーの自動作成

`.env` に `SEED_ADMIN_*` が設定されている状態でサーバーが起動すると、**ユーザーストアが空の場合のみ** admin ユーザーが自動作成されます。

```
# ログ例
INFO 初回 admin ユーザーを作成しました username=admin email=admin@example.com
```

> **注意**: admin 作成後は `.env` から `SEED_ADMIN_PASSWORD` を削除することを推奨します。

---

## コマンド一覧

| コマンド | 内容 |
|---|---|
| `make up` | バックエンド・フロントエンド・GCS エミュレータをまとめて起動 |
| `make down` | 全サービスを停止 |
| `make test` | 全テストを実行 |
| `make lint` | `golangci-lint` を実行 |
| `make swagger` | Swagger UI を起動 |
| `make build` | 本番用 Docker イメージをビルド |

---

## 本番環境（Cloud Run）のセットアップ

### 初回デプロイ時の admin 作成

Cloud Run の環境変数に `SEED_ADMIN_*` を設定してデプロイします。

```bash
gcloud run deploy cert-study-backend \
  --set-env-vars SEED_ADMIN_USERNAME=admin \
  --set-env-vars SEED_ADMIN_EMAIL=admin@example.com \
  --update-secrets SEED_ADMIN_PASSWORD=seed-admin-password:latest \
  --update-secrets JWT_SECRET=jwt-secret-key:latest \
  --update-secrets GCS_BUCKET=gcs-bucket-name:latest
```

admin 作成後は `SEED_ADMIN_PASSWORD` を環境変数から削除してください:

```bash
gcloud run services update cert-study-backend \
  --remove-env-vars SEED_ADMIN_PASSWORD
```

### シークレット管理

本番環境のシークレット（`JWT_SECRET`、`SEED_ADMIN_PASSWORD` など）は **GCP Secret Manager** で管理します。詳細は TICKET-010（GCP インフラ Terraform）を参照してください。

---

## アーキテクチャ

詳細は [CLAUDE.md](CLAUDE.md) を参照してください。
