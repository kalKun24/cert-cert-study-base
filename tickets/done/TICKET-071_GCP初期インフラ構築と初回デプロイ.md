# TICKET-071 GCP初期インフラ構築と初回デプロイ

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-071 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | 2026-06-22 |
| ブランチ名 | `feature/gcp-initial-deploy` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

GCP 上に本番環境を初回構築し、バックエンド・フロントエンドを Cloud Run にデプロイする。
Firestore（Native mode）・Artifact Registry・Secret Manager を手動セットアップし、動作確認まで完了させる。

---

## 背景・目的

ローカルでの品質確認が完了し、本番デプロイの前提条件が揃った。
まず手動で初期インフラを構築し、サービスが正常に動くことを確認する。
CD 自動化は TICKET-072 で後続対応する。

---

## 前提確認事項

着手前に以下を手元で確認しておくこと。

- [ ] `gcloud` CLI がインストール済みで、対象プロジェクトにログイン済み
- [ ] GCP プロジェクト ID が決定済み（例: `cert-study-prod`）
- [ ] デプロイ先リージョンが決定済み（推奨: `asia-northeast1`）
- [ ] 本番用の `JWT_SECRET` 値（32文字以上の安全なランダム文字列）を生成済み

---

## 受け入れ条件

- [ ] バックエンドが Cloud Run に正常デプロイされ、`/health` が 200 を返す
- [ ] フロントエンドが Cloud Run に正常デプロイされ、ブラウザでアクセスできる
- [ ] フロントエンドから API を叩いてログインが成功する
- [ ] `JWT_SECRET` が Secret Manager 経由で管理されている（コード・環境変数に直書きなし）
- [ ] Cloud Run の IAM が最小権限原則に従って設定されている

---

## サブチケット（コミット単位）

### フェーズ 1: GCP インフラ構築（コード変更なし・手動作業）

以下のコマンドを順番に実行する。コミットは不要。

```bash
# 1. 必要 API を有効化
gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  firestore.googleapis.com \
  secretmanager.googleapis.com \
  iam.googleapis.com \
  --project=<PROJECT_ID>

# 2. Artifact Registry リポジトリ作成
gcloud artifacts repositories create cert-study \
  --repository-format=docker \
  --location=asia-northeast1 \
  --project=<PROJECT_ID>

# 3. Firestore データベース作成（Native mode）
gcloud firestore databases create \
  --location=asia-northeast1 \
  --type=firestore-native \
  --project=<PROJECT_ID>

# 4. Secret Manager にシークレット登録
echo -n "<JWT_SECRET_VALUE>" | gcloud secrets create jwt-secret \
  --data-file=- \
  --project=<PROJECT_ID>

# 5. Cloud Run 用サービスアカウント作成
gcloud iam service-accounts create cert-study-backend \
  --display-name="cert-study backend" \
  --project=<PROJECT_ID>

# 6. Firestore アクセス権付与
gcloud projects add-iam-policy-binding <PROJECT_ID> \
  --member="serviceAccount:cert-study-backend@<PROJECT_ID>.iam.gserviceaccount.com" \
  --role="roles/datastore.user"

# 7. Secret Manager アクセス権付与
gcloud secrets add-iam-policy-binding jwt-secret \
  --member="serviceAccount:cert-study-backend@<PROJECT_ID>.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor" \
  --project=<PROJECT_ID>
```

---

### フェーズ 2: コード修正

- [x] `fix(frontend): nginx を Cloud Run 対応に修正（BACKEND_URL 環境変数サポート）`

  **問題**: 現在の nginx config は `http://backend:8080` にプロキシ（Docker Compose のサービス名）。
  Cloud Run では解決不能。
  
  **修正内容**: `Dockerfile` の nginx config 生成部分を `envsubst` テンプレート方式に変更し、
  起動時に `BACKEND_URL` 環境変数（Cloud Run のバックエンドサービス URL）を展開する。
  `PORT` 環境変数（Cloud Run が 8080 を注入）も反映する。

  ```dockerfile
  # 変更後のイメージ
  COPY nginx.conf.template /etc/nginx/templates/default.conf.template
  # Cloud Run は PORT 環境変数を注入する。nginx-alpine は /etc/nginx/templates/ を
  # 起動時に envsubst 展開して /etc/nginx/conf.d/ に配置する機能を持つ
  EXPOSE 8080
  ```

  nginx テンプレートの要点:
  ```nginx
  server {
      listen ${PORT};           # Cloud Run が注入する PORT（default 8080）
      location /api/ {
          proxy_pass ${BACKEND_URL};   # バックエンドの Cloud Run URL
      }
  }
  ```

---

### フェーズ 3: 初回デプロイ（手動）

- [ ] `chore(infra): バックエンドイメージをビルド・push・Cloud Run に初回デプロイ`

  ```bash
  # ビルド & push
  docker build -t asia-northeast1-docker.pkg.dev/<PROJECT_ID>/cert-study/backend:latest ./backend
  docker push asia-northeast1-docker.pkg.dev/<PROJECT_ID>/cert-study/backend:latest

  # Cloud Run デプロイ（バックエンド）
  gcloud run deploy cert-study-backend \
    --image=asia-northeast1-docker.pkg.dev/<PROJECT_ID>/cert-study/backend:latest \
    --region=asia-northeast1 \
    --service-account=cert-study-backend@<PROJECT_ID>.iam.gserviceaccount.com \
    --set-env-vars=GCP_PROJECT_ID=<PROJECT_ID> \
    --set-secrets=JWT_SECRET=jwt-secret:latest \
    --allow-unauthenticated \
    --port=8080 \
    --project=<PROJECT_ID>
  ```

- [ ] `chore(infra): フロントエンドイメージをビルド・push・Cloud Run に初回デプロイ`

  ```bash
  # バックエンドの URL を取得
  BACKEND_URL=$(gcloud run services describe cert-study-backend \
    --region=asia-northeast1 --format='value(status.url)' --project=<PROJECT_ID>)

  # ビルド & push
  docker build -t asia-northeast1-docker.pkg.dev/<PROJECT_ID>/cert-study/frontend:latest ./frontend
  docker push asia-northeast1-docker.pkg.dev/<PROJECT_ID>/cert-study/frontend:latest

  # Cloud Run デプロイ（フロントエンド）
  gcloud run deploy cert-study-frontend \
    --image=asia-northeast1-docker.pkg.dev/<PROJECT_ID>/cert-study/frontend:latest \
    --region=asia-northeast1 \
    --set-env-vars=BACKEND_URL=${BACKEND_URL},PORT=8080 \
    --allow-unauthenticated \
    --port=8080 \
    --project=<PROJECT_ID>
  ```

- [ ] 本番動作確認

  - バックエンド `GET /health` → `{"data":{"status":"ok"}}`
  - フロントエンドブラウザアクセス → ログイン画面表示
  - `admin / Admin1234!`（または本番用 admin）でログイン成功
  - チーム作成・問題作成・タグフィルタが正常動作

---

## 環境変数まとめ

### バックエンド（Cloud Run）

| 変数名 | 値 | 管理方法 |
|---|---|---|
| `PORT` | `8080` | Cloud Run 自動注入 |
| `GCP_PROJECT_ID` | GCP プロジェクト ID | 環境変数 |
| `JWT_SECRET` | ランダム文字列（32文字以上） | Secret Manager |
| `CORS_ALLOWED_ORIGINS` | フロントエンドの Cloud Run URL | 環境変数（フロントデプロイ後に更新） |

### フロントエンド（Cloud Run / nginx）

| 変数名 | 値 | 用途 |
|---|---|---|
| `PORT` | `8080` | Cloud Run 自動注入。nginx の listen ポート |
| `BACKEND_URL` | バックエンドの Cloud Run URL | nginx のプロキシ先 |

---

## 関連情報

- 関連チケット: TICKET-046（本チケットで実質的に代替。Terraform 化は別途検討）、TICKET-072（CD 自動化）
- 備考: CORS_ALLOWED_ORIGINS はフロントエンドのデプロイ後に確定するため、バックエンドを 2 回更新する手順になる
