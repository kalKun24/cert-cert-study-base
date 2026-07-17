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
| PR番号 | #88 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/88 |

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

- [ ] `.claude/settings.json` の permissions に curl（**localhost 限定を推奨**）/ `Bash(docker compose ps:*)` / `Bash(docker compose logs:*)` が追加されている
  - ※保留中: Claude 自身による permissions 変更は auto モード分類器に拒否されるため、ユーザーの手動対応が必要。また Security Engineer のレビューにより `Bash(curl:*)` 全開放は「.env 読取 + 任意外部送信」が揃うリスクが指摘されており、localhost 限定（例: ベースURLを強制するラッパースクリプト経由）での許可を推奨
- [ ] `testing-api-tester.md` が日本語・プロジェクト固有（openapi.yaml 契約検証・実行手順・読み取り専用 tools）に書き換えられている
- [ ] `testing-reality-checker.md` が日本語・証拠ベース最終判定役（デフォルト NEEDS WORK・make test / make lint 実行・抜き取り検証）に書き換えられている
- [ ] `qa-team.md` オーケストレーターが新設され、Phase 2 並列検証（Code Reviewer ‖ Security Engineer ‖ API Tester）→ Phase 3 Reality Checker 最終判定のフローが定義されている
- [x] API Tester の実行手順（make up → /health → admin ログイン）が実環境で通ることを確認済み
- [ ] QA Team を実起動し、並列検証→最終判定→報告のフローが機能することを確認済み
  - ※部分達成: フロー内容（Phase 2 3並列 → Phase 3 Reality Checker）は 2026-07-07 に代行実行で検証済み（下記「検証記録」）。QA Team オーケストレーター自体はセッション再起動後でないとエージェントレジストリに載らないため、真の実起動確認はマージ後の新セッションで1回行う

---

## サブチケット（コミット単位）

- [ ] `chore(claude): QA検証用にcurl・docker compose診断系のpermissionsを追加`（保留: ユーザー手動対応）
- [x] `feat(claude): API Testerをプロジェクト固有のAPI実挙動検証エージェントに刷新`
- [x] `feat(claude): Reality Checkerを証拠ベース最終判定エージェントに刷新`
- [x] `feat(claude): QA Teamオーケストレーターを追加`
- [x] `fix(claude): QAチーム初回実行レビューの指摘を反映`

---

## 検証記録（2026-07-07）

- **API 実行手順の実環境確認**: `make up` → `GET /health` 200 → `.env` の SEED_ADMIN_* で `POST /api/v1/auth/login` 成功（JWT 取得）→ 認証付き `GET /api/v1/users` 200 / トークン無し 401 を確認
- **QA フロー代行実行**: qa-team.md の記述どおり Phase 2（Code Reviewer ‖ Security Engineer ‖ API Tester の3並列）→ Phase 3（Reality Checker）を実施
  - API Tester: 主要導線スモーク 6/6 ✅（health / login / users / teams / 未認証401×2、openapi.yaml と契約一致）
  - Code Reviewer: 🔴 0件・🟡 4件（すべて反映済み）。前提事実（ポート・seed 条件・レート制限・{data,error} 形式）の実物一致を確認
  - Security Engineer: Critical 0件。High 1件は「curl:* 全開放のリスク」で permissions 保留判断に反映。Medium/Low の指摘は定義ファイルに反映済み
  - Reality Checker 最終判定: **CONDITIONAL**（ブロッカー 0・make test 成功 PASS 276 / SKIP 10（エミュレータ未起動分）・lint は CI 相当 v2.12.2 で 0 issues・抜き取り検証で矛盾ゼロ。残作業 = permissions のユーザー対応と真の実起動確認のみ）
- **範囲外の発見**: `make lint` は Makefile ピンの golangci-lint v1.62.2 が Go 1.25 と非互換でローカル実行不可（CI は v2.12.2 で正常）→ TICKET-083 として起票

---

## 関連情報

- 関連チケット: TICKET-081（Claude設定の刷新）
- 参考: `.claude/agents/dev-team.md`（オーケストレーターのパターン参照元）、`.claude/agents/engineering-code-reviewer.md` / `engineering-security-engineer.md`（読み取り専用エージェントのスタイル参照元）
- 備考: QA Team は読み取り専用チーム（ファイルを一切変更しない）。修正が必要な場合は Dev Team へ委譲する
- 受容リスク（Security Engineer レビューより）: 「読み取り専用」は frontmatter の tools 制限とプロンプト上の禁止事項による運用ルールであり、Bash を持つ以上技術的な強制ではない（既存の Code Reviewer / Security Engineer と同様の前提）。permissions は settings.json 全体に適用されエージェント単位でスコープできない
