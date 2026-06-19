# TICKET-046 GCPインフラ整備・CDパイプライン（優先度低）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-046 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-20 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/infra-setup` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

ローカルで納得のいくものができた後に着手。Terraform で GCP インフラを管理し、
GitHub Actions の CD パイプラインで Cloud Run への自動デプロイを整備する。

---

## 背景・目的

現状は手動デプロイ。コードがマージされたら自動でデプロイされる体制を整える。

---

## 受け入れ条件

- [ ] Terraform で Cloud Run・GCS バケット・Secret Manager を管理できる
- [ ] `main` ブランチへのマージをトリガーに CI（テスト）→ ビルド → Cloud Run デプロイが自動実行される
- [ ] シークレット（JWT 秘密鍵・GCS 認証情報等）が Secret Manager 経由で管理されている
- [ ] Cloud Run のリージョン・スペック・スケーリング設定が決定・反映されている
- [ ] GCS バケット名・環境ごとの命名規則が決定・反映されている

---

## サブチケット（コミット単位）

- [ ] `chore(infra): Terraform でGCS バケット・IAM を管理`
- [ ] `chore(infra): Terraform で Cloud Run サービスを管理`
- [ ] `chore(infra): Terraform で Secret Manager を管理`
- [ ] `chore(ci): GitHub Actions に CD ジョブ（ビルド→デプロイ）を追加`
- [ ] `docs: インフラ構成・デプロイ手順を README に記載`

---

## 関連情報

- 備考: ローカル開発でのエミュレーター方式は TICKET-016 で実装済み。本チケットは本番環境の整備のみ対象
