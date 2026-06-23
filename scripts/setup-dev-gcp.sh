#!/bin/bash
# cert-study-base dev 環境 GCP セットアップスクリプト
#
# このスクリプトは cert-study-base-dev GCP プロジェクトに必要なリソースを作成します。
# 実行前に以下が完了していることを確認してください:
#   1. gcloud auth login で認証済み
#   2. GCP コンソールで課金アカウントを cert-study-base-dev にリンク済み
#
# 使い方:
#   chmod +x scripts/setup-dev-gcp.sh
#   ./scripts/setup-dev-gcp.sh

set -euo pipefail

# ============================================================
# 変数定義
# ============================================================
PROJECT_ID="cert-study-base-dev"
REGION="asia-northeast1"
REPO_NAME="cert-study"
GITHUB_REPO="kalKun24/cert-study-base"
SERVICE_ACCOUNT_NAME="github-actions-sa"
SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
POOL_NAME="github-actions-pool"
PROVIDER_NAME="github-actions-provider"

# ============================================================
# 関数定義
# ============================================================

log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

# ステップ1: GCP プロジェクト作成
create_project() {
  log "ステップ1: GCP プロジェクトを作成します..."

  if gcloud projects describe "${PROJECT_ID}" --quiet > /dev/null 2>&1; then
    log "  プロジェクト ${PROJECT_ID} はすでに存在します。スキップします。"
  else
    gcloud projects create "${PROJECT_ID}" \
      --name="cert-study-base (dev)"
    log "  プロジェクト ${PROJECT_ID} を作成しました。"
  fi
}

# ステップ2: 必要な API を有効化
enable_apis() {
  log "ステップ2: 必要な API を有効化します..."

  gcloud services enable \
    run.googleapis.com \
    firestore.googleapis.com \
    artifactregistry.googleapis.com \
    secretmanager.googleapis.com \
    iam.googleapis.com \
    iamcredentials.googleapis.com \
    cloudresourcemanager.googleapis.com \
    --project="${PROJECT_ID}"

  log "  API の有効化が完了しました。"
}

# ステップ3: Artifact Registry リポジトリ作成
create_artifact_registry() {
  log "ステップ3: Artifact Registry リポジトリを作成します..."

  if gcloud artifacts repositories describe "${REPO_NAME}" \
    --location="${REGION}" \
    --project="${PROJECT_ID}" \
    --quiet > /dev/null 2>&1; then
    log "  リポジトリ ${REPO_NAME} はすでに存在します。スキップします。"
  else
    gcloud artifacts repositories create "${REPO_NAME}" \
      --repository-format=docker \
      --location="${REGION}" \
      --project="${PROJECT_ID}" \
      --description="cert-study-base Docker イメージ（dev 環境）"
    log "  Artifact Registry リポジトリ ${REPO_NAME} を作成しました。"
  fi
}

# ステップ4: Firestore データベース作成
create_firestore() {
  log "ステップ4: Firestore データベースを作成します..."

  if gcloud firestore databases describe \
    --project="${PROJECT_ID}" \
    --quiet > /dev/null 2>&1; then
    log "  Firestore データベースはすでに存在します。スキップします。"
  else
    gcloud firestore databases create \
      --project="${PROJECT_ID}" \
      --location="${REGION}" \
      --type=firestore-native
    log "  Firestore データベースを作成しました。"
  fi
}

# ステップ5: Cloud Run サービスのプレースホルダ作成
create_cloudrun_placeholders() {
  log "ステップ5: Cloud Run サービスのプレースホルダを作成します..."

  # バックエンドサービスのプレースホルダ
  if gcloud run services describe cert-study-backend \
    --region="${REGION}" \
    --project="${PROJECT_ID}" \
    --quiet > /dev/null 2>&1; then
    log "  cert-study-backend サービスはすでに存在します。スキップします。"
  else
    gcloud run deploy cert-study-backend \
      --image=gcr.io/cloudrun/hello \
      --region="${REGION}" \
      --project="${PROJECT_ID}" \
      --no-allow-unauthenticated \
      --quiet
    log "  cert-study-backend サービスのプレースホルダを作成しました。"
  fi

  # フロントエンドサービスのプレースホルダ
  if gcloud run services describe cert-study-frontend \
    --region="${REGION}" \
    --project="${PROJECT_ID}" \
    --quiet > /dev/null 2>&1; then
    log "  cert-study-frontend サービスはすでに存在します。スキップします。"
  else
    gcloud run deploy cert-study-frontend \
      --image=gcr.io/cloudrun/hello \
      --region="${REGION}" \
      --project="${PROJECT_ID}" \
      --no-allow-unauthenticated \
      --quiet
    log "  cert-study-frontend サービスのプレースホルダを作成しました。"
  fi
}

# ステップ6: Workload Identity Federation の設定
setup_workload_identity() {
  log "ステップ6: Workload Identity Federation を設定します..."

  # プロジェクト番号を取得
  PROJECT_NUMBER=$(gcloud projects describe "${PROJECT_ID}" \
    --format="value(projectNumber)")

  # Workload Identity プールの作成
  if gcloud iam workload-identity-pools describe "${POOL_NAME}" \
    --location=global \
    --project="${PROJECT_ID}" \
    --quiet > /dev/null 2>&1; then
    log "  Workload Identity プール ${POOL_NAME} はすでに存在します。スキップします。"
  else
    gcloud iam workload-identity-pools create "${POOL_NAME}" \
      --location=global \
      --project="${PROJECT_ID}" \
      --display-name="GitHub Actions Pool（dev）"
    log "  Workload Identity プール ${POOL_NAME} を作成しました。"
  fi

  # Workload Identity プロバイダの作成
  if gcloud iam workload-identity-pools providers describe "${PROVIDER_NAME}" \
    --workload-identity-pool="${POOL_NAME}" \
    --location=global \
    --project="${PROJECT_ID}" \
    --quiet > /dev/null 2>&1; then
    log "  プロバイダ ${PROVIDER_NAME} はすでに存在します。スキップします。"
  else
    gcloud iam workload-identity-pools providers create-oidc "${PROVIDER_NAME}" \
      --workload-identity-pool="${POOL_NAME}" \
      --location=global \
      --project="${PROJECT_ID}" \
      --issuer-uri="https://token.actions.githubusercontent.com" \
      --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
      --attribute-condition="attribute.repository==\"${GITHUB_REPO}\""
    log "  プロバイダ ${PROVIDER_NAME} を作成しました。"
  fi

  # サービスアカウントの作成
  if gcloud iam service-accounts describe "${SERVICE_ACCOUNT_EMAIL}" \
    --project="${PROJECT_ID}" \
    --quiet > /dev/null 2>&1; then
    log "  サービスアカウント ${SERVICE_ACCOUNT_EMAIL} はすでに存在します。スキップします。"
  else
    gcloud iam service-accounts create "${SERVICE_ACCOUNT_NAME}" \
      --project="${PROJECT_ID}" \
      --display-name="GitHub Actions サービスアカウント（dev）"
    log "  サービスアカウント ${SERVICE_ACCOUNT_EMAIL} を作成しました。"
  fi

  # サービスアカウントへの IAM ロール付与
  log "  サービスアカウントに IAM ロールを付与します..."

  for ROLE in \
    roles/run.admin \
    roles/iam.serviceAccountUser \
    roles/artifactregistry.writer \
    roles/datastore.user \
    roles/secretmanager.secretAccessor \
    roles/logging.logWriter; do
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
      --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
      --role="${ROLE}" \
      --quiet
    log "    ${ROLE} を付与しました。"
  done

  # Workload Identity バインディングの設定
  POOL_RESOURCE="projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_NAME}"
  gcloud iam service-accounts add-iam-policy-binding "${SERVICE_ACCOUNT_EMAIL}" \
    --project="${PROJECT_ID}" \
    --role=roles/iam.workloadIdentityUser \
    --member="principalSet://iam.googleapis.com/${POOL_RESOURCE}/attribute.repository/${GITHUB_REPO}" \
    --quiet
  log "  Workload Identity バインディングを設定しました。"
}

# ステップ7: Secret Manager シークレット作成
create_secrets() {
  log "ステップ7: Secret Manager のシークレットを作成します（値は空）..."

  for SECRET_NAME in jwt-secret seed-admin-password; do
    if gcloud secrets describe "${SECRET_NAME}" \
      --project="${PROJECT_ID}" \
      --quiet > /dev/null 2>&1; then
      log "  シークレット ${SECRET_NAME} はすでに存在します。スキップします。"
    else
      gcloud secrets create "${SECRET_NAME}" \
        --project="${PROJECT_ID}" \
        --replication-policy=automatic
      log "  シークレット ${SECRET_NAME} を作成しました。値は手動で設定してください。"
    fi
  done
}

# ステップ8: 完了メッセージと GitHub Secrets 設定値の出力
print_summary() {
  PROJECT_NUMBER=$(gcloud projects describe "${PROJECT_ID}" \
    --format="value(projectNumber)")

  PROVIDER_RESOURCE="projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_NAME}/providers/${PROVIDER_NAME}"

  echo ""
  echo "========================================"
  echo "セットアップが完了しました！"
  echo "========================================"
  echo ""
  echo "次のステップ:"
  echo "  1. Secret Manager のシークレット値を手動で設定してください:"
  echo "     - jwt-secret"
  echo "     - seed-admin-password"
  echo ""
  echo "  2. GitHub リポジトリの Settings > Secrets and variables > Actions に"
  echo "     以下を追加してください:"
  echo ""
  echo "========================================"
  echo "GitHub Secrets に設定する値:"
  echo "========================================"
  echo ""
  echo "GCP_WORKLOAD_IDENTITY_PROVIDER_DEV:"
  echo "  ${PROVIDER_RESOURCE}"
  echo ""
  echo "GCP_SERVICE_ACCOUNT_DEV:"
  echo "  ${SERVICE_ACCOUNT_EMAIL}"
  echo "========================================"
  echo ""
  echo "  3. GitHub Variables に以下を追加してください:"
  echo "     GCP_PROJECT_ID_DEV = ${PROJECT_ID}"
  echo "     FRONTEND_URL_DEV   = （初回デプロイ後に Cloud Run の URL を設定）"
  echo ""
  echo "  4. 初回デプロイ後に Cloud Run サービスの公開アクセスを設定してください:"
  echo "     docs/dev-environment-setup.md を参照してください。"
}

# ============================================================
# メイン処理
# ============================================================

log "cert-study-base dev 環境のセットアップを開始します..."
log "対象プロジェクト: ${PROJECT_ID}"
echo ""

create_project
enable_apis
create_artifact_registry
create_firestore
create_cloudrun_placeholders
setup_workload_identity
create_secrets
print_summary
