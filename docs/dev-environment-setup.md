# GCP dev 環境セットアップ手順書

## 概要

cert-study-base は prod / dev の2環境に分離して運用します。

| 項目 | prod | dev |
|---|---|---|
| GCP プロジェクト | `cert-study-base` | `cert-study-base-dev` |
| デプロイトリガー | `main` ブランチへのマージ | `develop` ブランチへのマージ |
| ワークフロー | `cd.yml` | `cd-dev.yml` |

dev 環境の目的は、`main` へマージする前に機能・バグ修正を安全に検証することです。本番データへの影響なしに統合テストを行えます。

---

## 前提条件

以下がすべて満たされていることを確認してください。

- `gcloud` CLI がインストールされ、`gcloud auth login` で認証済みであること
- GCP の課金アカウントが用意されていること
- GitHub リポジトリ（`kalKun24/cert-study-base`）の管理者権限があること

---

## ステップ1: 課金アカウントのリンク（手動）

> **注意**: 課金アカウントのリンクは自動化できないため、手動で実施してください。
> これを行わないと、ステップ2の API 有効化が失敗します。

### 1-1. 課金アカウント ID を確認する

```bash
gcloud billing accounts list
```

出力例:
```
ACCOUNT_ID            NAME                OPEN  MASTER_ACCOUNT_ID
012345-ABCDEF-789012  My Billing Account  True
```

### 1-2. dev プロジェクトに課金アカウントをリンクする

```bash
gcloud billing projects link cert-study-base-dev \
  --billing-account=BILLING_ACCOUNT_ID
```

`BILLING_ACCOUNT_ID` は上記で確認した値（例: `012345-ABCDEF-789012`）に置き換えてください。

---

## ステップ2: 自動セットアップスクリプトの実行

課金アカウントのリンク完了後、以下のスクリプトを実行してください。

```bash
chmod +x scripts/setup-dev-gcp.sh
./scripts/setup-dev-gcp.sh
```

スクリプトが実行する内容:
1. GCP プロジェクト `cert-study-base-dev` の作成
2. 必要な API の有効化
3. Artifact Registry リポジトリの作成
4. Firestore データベースの作成
5. Cloud Run サービスのプレースホルダ作成
6. Workload Identity Federation の設定
7. サービスアカウントへの IAM ロール付与
8. Secret Manager シークレットの作成（値は空）

スクリプト完了後、以下のような出力が表示されます。次のステップで使用するため、値を控えてください。

```
========================================
GitHub Secrets に設定する値:
========================================
GCP_WORKLOAD_IDENTITY_PROVIDER_DEV:
  projects/XXXXXXXXXX/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider

GCP_SERVICE_ACCOUNT_DEV:
  github-actions-sa@cert-study-base-dev.iam.gserviceaccount.com
========================================
```

---

## ステップ3: GitHub Variables/Secrets の設定（手動）

GitHub リポジトリの **Settings > Secrets and variables > Actions** を開き、以下を追加してください。

### Variables（Repository variables）

| 変数名 | 値 |
|---|---|
| `GCP_PROJECT_ID_DEV` | `cert-study-base-dev` |
| `FRONTEND_URL_DEV` | デプロイ後に Cloud Run フロントエンドの URL を設定（例: `https://cert-study-frontend-xxxx-an.a.run.app`） |

> **注意**: `FRONTEND_URL_DEV` は初回デプロイ後に Cloud Run コンソールで URL を確認してから設定してください。初回は空のままデプロイし、その後 URL を設定して再デプロイする運用で問題ありません。

### Secrets（Repository secrets）

| シークレット名 | 値 |
|---|---|
| `GCP_WORKLOAD_IDENTITY_PROVIDER_DEV` | スクリプト出力の `GCP_WORKLOAD_IDENTITY_PROVIDER_DEV` の値 |
| `GCP_SERVICE_ACCOUNT_DEV` | スクリプト出力の `GCP_SERVICE_ACCOUNT_DEV` の値 |

---

## ステップ4: Secret Manager のシークレット値設定（手動）

スクリプトはシークレットの「箱（リソース）」を作成しますが、実際の値は手動で設定する必要があります。

### JWT 秘密鍵を設定する

```bash
echo -n "YOUR_JWT_SECRET_VALUE" | \
  gcloud secrets versions add jwt-secret \
    --data-file=- \
    --project=cert-study-base-dev
```

### 初期管理者パスワードを設定する

```bash
echo -n "YOUR_SEED_ADMIN_PASSWORD" | \
  gcloud secrets versions add seed-admin-password \
    --data-file=- \
    --project=cert-study-base-dev
```

> **セキュリティ注意**: dev 環境のシークレットであっても、推測されやすい値は使用しないでください。prod とは異なる値を設定することを推奨します。

---

## 自動化されていない操作の一覧

以下の操作はスクリプトで自動化されていないため、手動での実施が必要です。

| 操作 | 理由 |
|---|---|
| 課金アカウントのリンク | GCP の制約により API での自動化が困難 |
| GitHub Variables/Secrets の追加 | セキュリティ上、スクリプトからの自動設定を避けるため |
| Secret Manager の値の設定 | 平文の秘密情報をスクリプト内に記述しないようにするため |
| Cloud Run サービスへの公開アクセス設定 | 初回デプロイ後に IAM で `allUsers` に `roles/run.invoker` を付与する必要がある |

### Cloud Run 公開アクセスの設定方法

初回デプロイ完了後、以下のコマンドで各サービスを一般公開します。

```bash
# バックエンドを公開
gcloud run services add-iam-policy-binding cert-study-backend \
  --region=asia-northeast1 \
  --member=allUsers \
  --role=roles/run.invoker \
  --project=cert-study-base-dev

# フロントエンドを公開
gcloud run services add-iam-policy-binding cert-study-frontend \
  --region=asia-northeast1 \
  --member=allUsers \
  --role=roles/run.invoker \
  --project=cert-study-base-dev
```
