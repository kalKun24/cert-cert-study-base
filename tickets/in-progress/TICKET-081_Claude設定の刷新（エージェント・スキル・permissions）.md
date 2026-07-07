# TICKET-081 Claude設定の刷新（エージェント・スキル・permissions）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-081 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-07-07 |
| 着手日 | 2026-07-07 |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-081` |
| PR番号 | （PR作成後に記入） |
| PRリンク | （PR作成後に記入） |

---

## 概要

`.claude/` 配下のエージェント・permissions・CLAUDE.md を監査した結果に基づき、実態と乖離した記述の修正、汎用ボイラープレートのプロジェクト特化、権限の整理、チケット管理スキルの追加を行う。

---

## 背景・目的

- 専門エージェント6件が汎用テンプレートのままで、本プロジェクトに存在しない技術（PostgreSQL・Redis・RabbitMQ・マイクロサービス等）の指示を含み、CLAUDE.md と矛盾していた
- CLAUDE.md の技術スタック（永続化=GCS、swaggo/swag）が Firestore 移行後の実態と乖離していた
- Dev Team / Design Team のトリガー（「TICKET-XXX を実装して」）が競合し、ルーティングが曖昧だった
- `Bash(rm *)` の包括許可など、permissions が過剰に広かった
- チケットライフサイクル操作が両オーケストレーターに重複記載されていた

---

## 受け入れ条件

- [ ] CLAUDE.md の技術スタック・Swagger・make コマンド表・TODO が実態（Firestore・Go 1.25・オフセットページネーション等）と一致している
- [ ] 専門エージェント6件が本プロジェクトの技術スタックのみを参照し、CLAUDE.md と矛盾しない
- [ ] Code Reviewer / Security Engineer / UI Designer / UX Architect が読み取り専用ツール構成になっている
- [ ] Design Team は提案フェーズ専任、実装は Dev Team に一本化されている
- [ ] project-setup.md が削除されている
- [ ] チケットライフサイクル操作が `.claude/skills/ticket/` に一元化されている
- [ ] `.claude/settings.json` から `Bash(rm *)` が除去され、表記がコロン形式に統一されている
- [ ] 編集時整形の PostToolUse hook が動作する（gofmt / prettier）

---

## サブチケット（コミット単位）

- [x] `docs(claude): CLAUDE.mdを実態に合わせて更新（Firestore・make表・TODO整理）`
- [x] `refactor(claude): 専門エージェント6件をプロジェクト特化にスリム化`
- [ ] `refactor(claude): Design Teamを提案専任化しDev Teamに実装を一本化`
- [ ] `chore(claude): project-setupエージェントを削除`
- [ ] `feat(claude): チケット管理スキルを追加`
- [ ] `chore(claude): permissionsの正規化と編集時整形hooksの追加`

---

## 関連情報

- 関連チケット: なし
- 参考: Claude Code サブエージェント/スキル/permissions のベストプラクティス監査（2026-07-07 実施）
- 備考: `~/.claude/settings.json`（グローバル）とメモリディレクトリの修正はリポジトリ外のためこのチケットのスコープ外（同日ローカルで実施）
