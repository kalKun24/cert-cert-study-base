# TICKET-076 GCP マルチ環境構築とブランチ戦略変更

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-076 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-23 |
| 着手日 | 2026-06-23 |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-076` |
| PR番号 | #74 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/74 |

---

## 概要

GCP プロジェクトを prod / dev の2環境に分離し、ブランチ戦略を3層（`main` / `develop` / `feature/*`）に変更する。
`feature/*` → `develop` → dev 環境、`develop` → `main` → prod 環境という2ステージのデプロイフローを整備する。

---

## 背景・目的

現在は `main` / `feature/*` の2層構成で、唯一の GCP プロジェクト（`cert-study-base`）に直接デプロイしている。
開発中の変更を本番に影響させずに動作確認できる dev 環境を用意し、品質を担保した上で本番リリースできるようにする。

---

## 受け入れ条件

- [ ] CLAUDE.md のブランチ戦略セクションが3層構成の内容に更新されている
- [ ] CI ワークフロー（ci.yml）が `develop` ブランチのプッシュ/PR にも動作する
- [ ] `cd.yml` が `main` ブランチトリガー（prod 専用）として整理されている
- [ ] `cd-dev.yml` が新規作成され、`develop` ブランチトリガーで dev プロジェクトにデプロイする
- [ ] GCP dev プロジェクト（`cert-study-base-dev`）のセットアップ手順書が `docs/dev-environment-setup.md` に作成されている
- [ ] セットアップスクリプト（`scripts/setup-dev-gcp.sh`）が作成されており、自動化できる手順を実行できる
- [ ] `develop` ブランチが作成され、リモートにプッシュされている

---

## サブチケット（コミット単位）

- [x] `docs(claude): CLAUDE.md のブランチ戦略・CI/CD セクションを3層構成に更新`
- [x] `chore(ci): ci.yml を develop ブランチにも対応`
- [x] `chore(cd): cd.yml を prod 専用に整理`
- [x] `chore(cd): cd-dev.yml を新規作成（develop トリガー・dev 環境）`
- [x] `docs(infra): GCP dev 環境セットアップ手順書を作成`
- [x] `chore(infra): GCP dev 環境セットアップスクリプトを作成`

---

## 関連情報

- 関連チケット: TICKET-046（GCPインフラ整備・CDパイプライン）
- 参考: 既存 prod プロジェクト `cert-study-base`、既存 `cd.yml` の設定
- 備考:
  - 課金アカウントのリンク・一部 IAM 権限付与は手動手順書に記載する
  - GitHub Variables/Secrets の追加（`GCP_PROJECT_ID_DEV`, `FRONTEND_URL_DEV`, `GCP_WORKLOAD_IDENTITY_PROVIDER_DEV`, `GCP_SERVICE_ACCOUNT_DEV`）も手順書に記載する
  - PR は `develop` ブランチ向けに作成する（CLAUDE.md 更新・develop ブランチ作成自体は main へのコミットで問題ない）
