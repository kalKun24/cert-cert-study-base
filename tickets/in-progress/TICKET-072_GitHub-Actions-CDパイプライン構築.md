# TICKET-072 GitHub Actions CD パイプライン構築

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-072 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-23 |
| 完了日 | - |
| ブランチ名 | `feature/cd-pipeline` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

`main` ブランチへのマージをトリガーに、Docker イメージをビルド → Artifact Registry へプッシュ → Cloud Run へデプロイする CD ワークフローを GitHub Actions に追加する。
認証には Workload Identity Federation を使用し、サービスアカウントキーをリポジトリに保存しない。

---

## 背景・目的

TICKET-071 で手動デプロイが完了した後、毎回手動でビルド・デプロイするのを自動化する。
`main` マージ後に自動で本番反映されることで、開発サイクルを短縮する。

---

## 前提条件

- [ ] TICKET-071 が完了済み（Cloud Run サービス・Artifact Registry が存在する）
- [ ] GCP プロジェクト ID が確定済み

---

## 受け入れ条件

- [ ] `main` へのマージ後、CI（既存）→ CD（新規）の順でパイプラインが実行される
- [ ] CD ジョブがサービスアカウントキーなし（Workload Identity Federation）で GCP 認証できる
- [ ] バックエンド・フロントエンド両方が自動デプロイされる
- [ ] デプロイ完了後に Cloud Run のサービス URL がジョブのサマリーに出力される
- [ ] シークレット（JWT_SECRET 等）はリポジトリの Secrets に平文で保存しない

---

## サブチケット（コミット単位）

- [ ] `chore(infra): Workload Identity Federation を設定（GCP 側）`

  ```bash
  # Workload Identity Pool 作成
  gcloud iam workload-identity-pools create github-actions \
    --location=global \
    --project=<PROJECT_ID>

  # GitHub Actions 用 Provider 作成
  gcloud iam workload-identity-pools providers create-oidc github \
    --location=global \
    --workload-identity-pool=github-actions \
    --issuer-uri=https://token.actions.githubusercontent.com \
    --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
    --project=<PROJECT_ID>

  # サービスアカウントに権限付与（Artifact Registry push + Cloud Run deploy）
  gcloud iam service-accounts add-iam-policy-binding \
    cert-study-backend@<PROJECT_ID>.iam.gserviceaccount.com \
    --member="principalSet://iam.googleapis.com/projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/github-actions/attribute.repository/kalKun24/cert-study-base" \
    --role="roles/iam.workloadIdentityUser" \
    --project=<PROJECT_ID>
  ```

  GitHub Secrets に登録する値:
  - `GCP_PROJECT_ID` : GCP プロジェクト ID
  - `GCP_WORKLOAD_IDENTITY_PROVIDER` : Workload Identity Provider のリソース名
  - `GCP_SERVICE_ACCOUNT` : サービスアカウントのメールアドレス

- [ ] `chore(ci): GitHub Actions に CD ワークフロー（.github/workflows/cd.yml）を追加`

  ワークフローの骨格:

  ```yaml
  name: CD

  on:
    push:
      branches: [main]

  jobs:
    deploy-backend:
      name: バックエンドデプロイ
      runs-on: ubuntu-latest
      permissions:
        id-token: write
        contents: read
      steps:
        - uses: actions/checkout@v4
        - uses: google-github-actions/auth@v2
          with:
            workload_identity_provider: ${{ secrets.GCP_WORKLOAD_IDENTITY_PROVIDER }}
            service_account: ${{ secrets.GCP_SERVICE_ACCOUNT }}
        - uses: google-github-actions/setup-gcloud@v2
        - name: Docker 認証
          run: gcloud auth configure-docker asia-northeast1-docker.pkg.dev
        - name: バックエンドビルド & push
          run: |
            docker build -t asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/cert-study/backend:${{ github.sha }} ./backend
            docker push asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/cert-study/backend:${{ github.sha }}
        - name: Cloud Run デプロイ（バックエンド）
          uses: google-github-actions/deploy-cloudrun@v2
          with:
            service: cert-study-backend
            region: asia-northeast1
            image: asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/cert-study/backend:${{ github.sha }}

    deploy-frontend:
      name: フロントエンドデプロイ
      runs-on: ubuntu-latest
      needs: deploy-backend
      permissions:
        id-token: write
        contents: read
      steps:
        # ... バックエンドと同様のステップ
        # deploy-backend の出力 URL を BACKEND_URL として渡す
  ```

- [ ] `docs: CD パイプラインの設定手順を README に追記`

---

## 注意事項

- Workload Identity Federation の設定は GCP コンソール or gcloud で **1 回限りの手動作業**
- サービスアカウントには最小権限のみ付与する:
  - `roles/run.developer`（Cloud Run デプロイ）
  - `roles/artifactregistry.writer`（イメージ push）
  - `roles/iam.serviceAccountUser`（サービスアカウント利用）
- フロントエンドのデプロイでは `BACKEND_URL` を動的に解決する必要がある（バックエンドの deploy job の output を参照）

---

## 関連情報

- 関連チケット: TICKET-071（本チケットの前提）、TICKET-046（本チケットで代替）
- 参考: [google-github-actions/deploy-cloudrun](https://github.com/google-github-actions/deploy-cloudrun)
- 参考: [Workload Identity Federation - GitHub Actions](https://cloud.google.com/blog/products/identity-security/enabling-keyless-authentication-from-github-actions)
