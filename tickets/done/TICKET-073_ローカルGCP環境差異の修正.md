# TICKET-073 ローカル・GCP 環境差異の修正

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-073 |
| ステータス | 🟢 完了 |
| 作成日 | 2026-06-23 |
| 着手日 | 2026-06-23 |
| 完了日 | 2026-06-23 |
| ブランチ名 | `fix/local-gcp-env` |
| PR番号 | #69 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/69 |

---

## 概要

GCP デプロイ対応時に加えた変更が原因でローカル環境が起動しなくなっていた問題と、その調査で発見した追加不具合を修正する。

---

## 背景・目的

PR #65（GCP 初回デプロイ）で nginx に `resolver 8.8.8.8` をハードコードしたため、Docker Compose 内部 DNS が使えなくなりローカルの API プロキシが 502 を返すようになった。合わせて全体を調査し、レート制限・CORS・CD パイプラインの不具合も修正する。

---

## 受け入れ条件

- [x] `make up` でローカル環境が正常起動し、フロントエンドから API が疎通する
- [x] Cloud Run へのデプロイ時も nginx が正常に動作する
- [x] レート制限が Cloud Run 経由のリクエストでも正しく機能する
- [x] 本番バックエンドの CORS が許可オリジンのみを返す
- [x] CD パイプラインが `GCP_PROJECT_ID` を明示的に注入する

---

## サブチケット（コミット単位）

- [x] `fix(nginx): resolver を環境変数化してローカルと GCP で切り替え可能にする`
- [x] `fix(middleware): X-Forwarded-For のパース処理を修正し Cloud Run でも正しく IP を取得する`
- [x] `fix(cd): バックエンドに GCP_PROJECT_ID・CORS_ALLOWED_ORIGINS を明示的に注入する`
- [x] `chore(docker-compose): obsolete な version フィールドを削除`
- [x] `docs(env): CORS_ALLOWED_ORIGINS の設定例を .env.example に追記`

---

## 関連情報

- 関連チケット: TICKET-071（GCP 初回デプロイ）、TICKET-072（CD パイプライン）
- 備考: `FRONTEND_URL` を GitHub Actions repository variable に登録済み（値: `https://cert-study-frontend-z3wvc2r7zq-an.a.run.app`）
