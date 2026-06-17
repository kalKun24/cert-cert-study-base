# TICKET-010 GCPインフラ構築（Terraform）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-010 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-17 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/infra-terraform` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

Cloud Run へのデプロイに必要な GCP リソース一式を Terraform で構築する。
人間が手動でやるべき作業（課金設定・シークレット値の入力）を明確に分離し、Claude が Terraform コードを書いてその他を自動化する。

---

## 背景・目的

- アプリを Cloud Run で稼働させるために GCP 側のリソースが必要
- 手作業による設定ミスを防ぐため、インフラをコードで管理（IaC）する
- 環境再現性を高め、将来の本番/ステージング分離にも対応できる基盤を作る

---

## ⚠️ 人間が必ず行う作業（Claude には実行不可）

このチケットには **人間だけが実施できる作業** が含まれる。
Claude はそのタイミングで作業依頼を行い、完了確認を得てから次のステップへ進む。

### STEP-H1（着手前）: GCP プロジェクト作成 ＋ 課金アカウント紐付け

> **Claude が着手する前に完了させてください**

1. [Google Cloud Console](https://console.cloud.google.com/) でプロジェクトを新規作成
2. プロジェクト ID を控える（例: `cert-study-prod-123456`）
3. 「お支払い」からプロジェクトに課金アカウントを紐付ける
4. 以下の値を Claude に伝える:
   - `PROJECT_ID`（例: `cert-study-prod-123456`）
   - `BILLING_ACCOUNT_ID`（課金アカウントID）

### STEP-H2（Terraform apply 前）: 設計値の確定

> **Claude が Terraform コードを書いた後、apply 前に確認が必要**

以下の値を決めて Claude に伝える（CLAUDE.md の未決定事項）:

| 項目 | 例 |
|---|---|
| Cloud Run リージョン | `asia-northeast1`（東京）|
| GCS バケット名 | `cert-study-data-prod` |
| GCS バケット リージョン | `asia-northeast1` |
| Artifact Registry リポジトリ名 | `cert-study-backend` |

### STEP-H3（Terraform apply 後）: シークレット値の登録

> **Terraform apply で Secret Manager のキーが作成されたら、値を登録する**

Secret Manager コンソールから以下の値を手動で入力:

- `jwt-secret-key`：JWT 署名用の秘密鍵（openssl rand -base64 32 などで生成）

> **理由**: シークレットの値は Terraform の state ファイルに残るとリスクがあるため、値の登録は人間が手動で行う

---

## 受け入れ条件

- [ ] `terraform plan` がエラーなく通る
- [ ] `terraform apply` で全リソースが正常に作成される
- [ ] GitHub Actions から Workload Identity Federation で GCP 認証できる
- [ ] Cloud Run に手動デプロイ（初回確認用）が成功する
- [ ] GCS バケットへの読み書きが Cloud Run から動作する

---

## サブチケット（コミット単位）

> ※ STEP-H1〜H3 の人間作業完了後にそれぞれ着手する

- [ ] `chore(infra): Terraform ディレクトリ構成と provider 設定を追加`
- [ ] `feat(infra): API 有効化・Artifact Registry・GCS バケットを Terraform 化`
- [ ] `feat(infra): サービスアカウントと IAM ロール付与を Terraform 化`
- [ ] `feat(infra): Workload Identity Federation を Terraform 化`
- [ ] `feat(infra): Secret Manager キーの定義を Terraform 化`
- [ ] `feat(infra): Cloud Run サービスの初期設定を Terraform 化`
- [ ] `docs(infra): README にインフラ構築手順（人間作業含む）を記載`

---

## 作業フロー（Claude ↔ 人間）

```
[人間] STEP-H1: GCPプロジェクト作成 ＋ 課金設定
    ↓ PROJECT_ID を Claude に連絡
[Claude] Terraform コード実装（STEP-H2 の値を仮置きでも着手可）
    ↓
[人間] STEP-H2: 設計値を確定して Claude に連絡
    ↓
[Claude] terraform plan を実行・確認してもらう
    ↓
[人間] plan の内容を確認し、apply を承認
    ↓
[Claude] terraform apply を実行
    ↓
[人間] STEP-H3: Secret Manager にシークレット値を登録
    ↓
[Claude] 動作確認・README 更新
```

---

## 関連情報

- 関連チケット: TICKET-001（プロジェクト基盤構築）
- 参考: [Cloud Run 公式ドキュメント](https://cloud.google.com/run/docs)
- 参考: [Workload Identity Federation for GitHub Actions](https://cloud.google.com/blog/products/identity-security/enabling-keyless-authentication-from-github-actions)
- 備考: Terraform state は GCS バケットで管理する（バックエンド設定を含む）
- 備考: 本番・ステージング環境分離は本チケットのスコープ外。まず本番1環境を構築する
