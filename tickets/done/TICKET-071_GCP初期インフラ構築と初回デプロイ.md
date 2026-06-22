# TICKET-071 GCP初期インフラ構築と初回デプロイ

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-071 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | 2026-06-23 |
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

- [x] バックエンドが Cloud Run に正常デプロイされ、`/health` が 200 を返す
- [x] フロントエンドが Cloud Run に正常デプロイされ、ブラウザでアクセスできる
- [x] フロントエンドから API を叩いてログインが成功する
- [x] `JWT_SECRET` が Secret Manager 経由で管理されている（コード・環境変数に直書きなし）
- [x] Cloud Run の IAM が最小権限原則に従って設定されている

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

- [x] `chore(infra): バックエンドイメージをビルド・push・Cloud Run に初回デプロイ`

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

- [x] `fix(frontend): nginx Host ヘッダー修正・再ビルド・再デプロイ`

  **背景**: nginx の `proxy_set_header Host $host` がフロントエンドのホスト名を
  バックエンドへ送っていたため Cloud Run が 502 を返していた。
  `BACKEND_HOST` 環境変数で正しいバックエンドのホスト名を明示する方式に修正済み。

  ```bash
  # バックエンドの URL・ホスト名を取得
  BACKEND_URL=$(gcloud run services describe cert-study-backend \
    --region=asia-northeast1 --format='value(status.url)' --project=<PROJECT_ID>)
  BACKEND_HOST=$(echo $BACKEND_URL | sed 's|https://||')

  # ビルド & push（タグ番号は適宜変える）
  docker build -t asia-northeast1-docker.pkg.dev/<PROJECT_ID>/cert-study/frontend:v5 ./frontend
  docker push asia-northeast1-docker.pkg.dev/<PROJECT_ID>/cert-study/frontend:v5

  # Cloud Run デプロイ（PORT は Cloud Run が自動注入するため --set-env-vars に含めない）
  gcloud run deploy cert-study-frontend \
    --image=asia-northeast1-docker.pkg.dev/<PROJECT_ID>/cert-study/frontend:v5 \
    --region=asia-northeast1 \
    --set-env-vars="BACKEND_URL=${BACKEND_URL},BACKEND_HOST=${BACKEND_HOST}" \
    --allow-unauthenticated \
    --port=8080 \
    --project=<PROJECT_ID>
  ```

- [x] `chore(infra): バックエンドに SEED_ADMIN_* 環境変数を設定して admin ユーザーを作成`

  seed.go の仕様: `SEED_ADMIN_USERNAME` / `SEED_ADMIN_PASSWORD` / `SEED_ADMIN_EMAIL`
  の 3 つが設定されており、かつ Firestore にユーザーが 0 件のときのみ admin を作成する。
  パスワードは Secret Manager (`seed-admin-password`) に登録し、`--update-secrets` で注入する方式を採用。

  ```bash
  # 1. Secret Manager にパスワードを登録
  echo -n '<PASSWORD>' | gcloud secrets create seed-admin-password \
    --data-file=- --project=<PROJECT_ID>

  # 2. バックエンド SA に読み取り権限付与
  gcloud secrets add-iam-policy-binding seed-admin-password \
    --member="serviceAccount:cert-study-backend@<PROJECT_ID>.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor" --project=<PROJECT_ID>

  # 3. 環境変数（非機密）と Secret（パスワード）をまとめて設定
  gcloud run services update cert-study-backend \
    --region=asia-northeast1 \
    --update-env-vars="GCP_PROJECT_ID=<PROJECT_ID>,SEED_ADMIN_USERNAME=CertBaseAdmin,SEED_ADMIN_EMAIL=grenouille24hi@gmail.com,SEED_ADMIN_DISPLAY_NAME=Administrator" \
    --remove-env-vars="SEED_ADMIN_PASSWORD" \
    --update-secrets="SEED_ADMIN_PASSWORD=seed-admin-password:latest" \
    --project=<PROJECT_ID>
  ```

  > 起動後に seed が実行され、ログに
  > `"初回 admin ユーザーを作成しました" username=CertBaseAdmin email=...` が出ることを確認。
  > 確認後、SEED_ADMIN_* 変数は削除しても問題ない（冪等のため再実行は無害）。

- [x] `chore(infra): バックエンドの CORS_ALLOWED_ORIGINS をフロントエンド URL に設定`

  ```bash
  FRONTEND_URL=$(gcloud run services describe cert-study-frontend \
    --region=asia-northeast1 --format='value(status.url)' --project=<PROJECT_ID>)

  gcloud run services update cert-study-backend \
    --region=asia-northeast1 \
    --set-env-vars="CORS_ALLOWED_ORIGINS=${FRONTEND_URL}" \
    --project=<PROJECT_ID>
  ```

- [x] 本番動作確認

  - [x] バックエンド `GET /health` → `{"data":{"status":"ok"}}`
  - [x] フロントエンドブラウザアクセス → ログイン画面表示
  - [x] seed で設定した admin 認証情報でログイン成功
  - [ ] チーム作成・問題作成・タグフィルタが正常動作（ログイン後の動作確認は別途）

---

## 環境変数まとめ

### バックエンド（Cloud Run）

| 変数名 | 値 | 管理方法 |
|---|---|---|
| `PORT` | `8080` | Cloud Run 自動注入 |
| `GCP_PROJECT_ID` | GCP プロジェクト ID | 環境変数 |
| `JWT_SECRET` | ランダム文字列（32文字以上） | Secret Manager |
| `CORS_ALLOWED_ORIGINS` | フロントエンドの Cloud Run URL | 環境変数（フロントデプロイ後に更新） |

### バックエンド（初回のみ・seed 用）

| 変数名 | 値 | 管理方法 |
|---|---|---|
| `SEED_ADMIN_USERNAME` | `admin` | 環境変数（seed 後は削除可） |
| `SEED_ADMIN_PASSWORD` | 強いパスワード | 環境変数（seed 後は削除可） |
| `SEED_ADMIN_EMAIL` | `grenouille24hi@gmail.com` | 環境変数（seed 後は削除可） |
| `SEED_ADMIN_DISPLAY_NAME` | `Administrator` | 環境変数（seed 後は削除可） |

### フロントエンド（Cloud Run / nginx）

| 変数名 | 値 | 用途 |
|---|---|---|
| `PORT` | `8080` | Cloud Run 自動注入。nginx の listen ポート |
| `BACKEND_URL` | バックエンドの Cloud Run URL（`https://...`） | nginx のプロキシ先 URL |
| `BACKEND_HOST` | バックエンドのホスト名（`https://` なし） | nginx の `Host` ヘッダー設定 |

---

## 関連情報

- 関連チケット: TICKET-046（本チケットで実質的に代替。Terraform 化は別途検討）、TICKET-072（CD 自動化）
- 備考: CORS_ALLOWED_ORIGINS はフロントエンドのデプロイ後に確定するため、バックエンドを 2 回更新する手順になる
