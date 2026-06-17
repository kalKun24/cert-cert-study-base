# TICKET-015 CD パイプライン（Cloud Run 自動デプロイ）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-015 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-17 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/cd-pipeline` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

`main` ブランチへのマージをトリガーに、Docker イメージのビルド → Artifact Registry へのプッシュ → Cloud Run へのデプロイ を自動化する GitHub Actions ワークフローを実装する。

---

## 背景・目的

CLAUDE.md に「main マージトリガーに、テスト → ビルド → Cloud Run デプロイ」と定義されているが、現状 CI（テスト・lint）しか実装されていない。TICKET-010（Terraform）でインフラが整備された後、デプロイの自動化が必要。

---

## 前提条件

- TICKET-010（GCP インフラ Terraform）が完了していること
  - Artifact Registry リポジトリが存在すること
  - Cloud Run サービスが初期作成されていること
  - Workload Identity Federation が設定されていること

---

## ワークフロー設計

```
トリガー: push to main

ジョブ1: test（既存 CI と統合）
  - go test / golangci-lint
  - frontend lint / build

ジョブ2: build-and-push（test 成功後）
  - Workload Identity Federation で GCP 認証
  - docker build（バックエンド）
  - docker push → Artifact Registry

ジョブ3: deploy（build-and-push 成功後）
  - gcloud run deploy --image <新イメージ>
  - デプロイ完了確認（ヘルスチェック URL を叩く）
```

---

## 受け入れ条件

- [ ] `main` へのマージ後、自動で Cloud Run にデプロイされる
- [ ] テストが失敗した場合、ビルド・デプロイは実行されない
- [ ] Workload Identity Federation によるキーレス認証で GCP にアクセスできる（サービスアカウントキーの JSON を GitHub Secrets に置かない）
- [ ] デプロイ後にヘルスチェック（`/health`）を叩いて正常を確認する
- [ ] デプロイ失敗時に GitHub Actions 上でエラーが明確に確認できる

---

## サブチケット（コミット単位）

- [ ] `chore(ci): CD ワークフロー（build-and-push ジョブ）を追加`
- [ ] `chore(ci): CD ワークフロー（Cloud Run deploy ジョブ）を追加`
- [ ] `chore(ci): デプロイ後ヘルスチェックステップを追加`
- [ ] `docs(readme): デプロイフロー・ロールバック手順を記載`

---

## 関連情報

- 関連チケット: TICKET-001（CI 基盤）、TICKET-010（GCP インフラ Terraform）
- 備考: フロントエンドの静的ファイルの配信方法（Cloud Run で同梱 or 別途 CDN）は本チケット着手前に確認すること
