# TICKET-046 GCPインフラ整備・CDパイプライン（優先度低）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-046 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-20 |
| 着手日 | 2026-06-22 |
| 完了日 | 2026-06-23 |
| ブランチ名 | 複数PR |
| PR番号 | #65 #69 #72 #74 #75 |
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

- [x] ~~Terraform で Cloud Run・GCS バケット・Secret Manager を管理できる~~ → Terraform 不採用。gcloud + setup-dev-gcp.sh で管理。GCS は Firestore 移行済みにつき不要
- [x] `main` ブランチへのマージをトリガーに CI（テスト）→ ビルド → Cloud Run デプロイが自動実行される
- [x] シークレット（JWT 秘密鍵等）が Secret Manager 経由で管理されている
- [x] Cloud Run のリージョン（asia-northeast1）が決定・反映されている
- [x] dev / prod の2環境が分離されている（cert-study-base-dev / cert-study-base）

---

## サブチケット（コミット単位）

- [x] GCP 初回インフラ構築・手動デプロイ（PR #65）
- [x] GitHub Actions CD パイプライン構築（PR #69 #72）
- [x] dev/prod マルチ環境・ブランチ戦略3層化（PR #74 #75）

---

## 関連情報

- 備考: ローカル開発でのエミュレーター方式は TICKET-016 で実装済み。本チケットは本番環境の整備のみ対象
