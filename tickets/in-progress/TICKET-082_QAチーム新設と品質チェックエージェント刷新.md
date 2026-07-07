# TICKET-082 QAチーム新設と品質チェックエージェント刷新

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-082 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-07-07 |
| 着手日 | 2026-07-07 |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-082` |
| PR番号 | - |
| PRリンク | （PR作成後に記入） |

---

## 概要

品質チェック専任のエージェントチーム「QA Team」を新設する。汎用テンプレートのまま残っていた `testing-api-tester.md` / `testing-reality-checker.md` をプロジェクト固有の定義に全面書き換えし、既存の Code Reviewer / Security Engineer と合わせた4エージェントを束ねるオーケストレーター `qa-team.md` を追加する。

---

## 背景・目的

- TICKET-081 で `.claude/` 構成を刷新したが、`testing-*` の2エージェントは英語の汎用ボイラープレート（Laravel・Playwright スクリーンショット前提）のままで、本プロジェクトの実態（Go + Firestore + REST API）と不整合だった
- openapi.yaml には約49オペレーションあるが、httptest は note/question の2リソースのみで E2E / API 契約検証が存在しない。実行ベースの API 検証（API Tester）でこの空白を埋める
- Dev Team の Phase 3（Code Reviewer ‖ Security Engineer）は実装フローの軽量レビューとして残し、QA Team は PR 前の総合検査として独立起動できるようにする

---

## 受け入れ条件

- [ ] `.claude/settings.json` の permissions に `Bash(curl:*)` / `Bash(docker compose ps:*)` / `Bash(docker compose logs:*)` が追加されている
- [ ] `testing-api-tester.md` が日本語・プロジェクト固有（openapi.yaml 契約検証・実行手順・読み取り専用 tools）に書き換えられている
- [ ] `testing-reality-checker.md` が日本語・証拠ベース最終判定役（デフォルト NEEDS WORK・make test / make lint 実行・抜き取り検証）に書き換えられている
- [ ] `qa-team.md` オーケストレーターが新設され、Phase 2 並列検証（Code Reviewer ‖ Security Engineer ‖ API Tester）→ Phase 3 Reality Checker 最終判定のフローが定義されている
- [ ] API Tester の実行手順（make up → /health → admin ログイン）が実環境で通ることを確認済み
- [ ] QA Team を実起動し、並列検証→最終判定→報告のフローが機能することを確認済み

---

## サブチケット（コミット単位）

- [ ] `chore(claude): QA検証用にcurl・docker compose診断系のpermissionsを追加`
- [ ] `feat(claude): API Testerをプロジェクト固有のAPI実挙動検証エージェントに刷新`
- [ ] `feat(claude): Reality Checkerを証拠ベース最終判定エージェントに刷新`
- [ ] `feat(claude): QA Teamオーケストレーターを追加`

---

## 関連情報

- 関連チケット: TICKET-081（Claude設定の刷新）
- 参考: `.claude/agents/dev-team.md`（オーケストレーターのパターン参照元）、`.claude/agents/engineering-code-reviewer.md` / `engineering-security-engineer.md`（読み取り専用エージェントのスタイル参照元）
- 備考: QA Team は読み取り専用チーム（ファイルを一切変更しない）。修正が必要な場合は Dev Team へ委譲する
