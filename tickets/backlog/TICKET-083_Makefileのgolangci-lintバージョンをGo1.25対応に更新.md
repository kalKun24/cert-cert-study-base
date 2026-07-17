# TICKET-083 Makefileのgolangci-lintバージョンをGo1.25対応に更新

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-083 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-07-07 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | - |
| PR番号 | - |
| PRリンク | （PR作成後に記入） |

---

## 概要

`make lint` がローカルで実行できない問題を解消する。Makefile がピンしている golangci-lint v1.62.2 は Go 1.23 でビルドされており、Go 1.25 のプロジェクトに対して `can't load config: the Go language version (go1.23) used to build golangci-lint is lower than the targeted Go version (1.25.0)` で失敗する。

---

## 背景・目的

- TICKET-082 の QA フロー検証中に Reality Checker が発見した
- CI（`.github/workflows/ci.yml`）は golangci-lint-action@v7 + **v2.12.2** で正常に動作しており、ローカルの `make lint` だけが壊れている
- ローカルと CI の lint 結果が一致しないと、QA Team（Reality Checker）の「make lint を証拠にする」プロセスが機能しない
- 注意: `make lint` は実行時に `go install ...@v1.62.2` を行うため、開発者の `~/go/bin/golangci-lint` を非互換版で上書きする副作用もある

---

## 受け入れ条件

- [ ] Makefile の `GOLANGCI_LINT_VERSION` が CI と同じ v2.12.2（またはそれ以降の Go 1.25 対応版）に更新されている
- [ ] `make lint` がローカルで正常終了する
- [ ] v1 → v2 で設定ファイル（.golangci.yml 等がある場合）の互換性が確認されている

---

## サブチケット（コミット単位）

- [ ] `chore(build): golangci-lintをv2.12.2に更新しmake lintのGo1.25非互換を解消`

---

## 関連情報

- 関連チケット: TICKET-082（発見元）
- 参考: `.github/workflows/ci.yml` の lint ジョブ（v2.12.2 使用）
- 備考: golangci-lint v1 → v2 は CLI・設定形式に破壊的変更があるため、`.golangci.yml` が存在する場合は `golangci-lint migrate` の要否を確認する
