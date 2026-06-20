# TICKET-052 TeamUseCaseコンストラクタ一本化

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-052 |
| ステータス | 🟢 完了 |
| 作成日 | 2026-06-20 |
| 着手日 | 2026-06-20 |
| 完了日 | 2026-06-20 |
| ブランチ名 | `refactor/team-usecase-constructor` |
| PR番号 | #33 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/33 |

---

## 概要

TICKET-042 の実装で `NewTeamUseCase`（旧）と `NewTeamUseCaseWithStats`（新）の2つのコンストラクタが並存する状態になった。将来の混乱・バグリスクを防ぐため、コンストラクタを一本化するリファクタリングを行う。

---

## 背景・目的

TICKET-042 で `ListMemberStats` ユースケース追加のために `NewTeamUseCaseWithStats` を新設したが、旧コンストラクタ `NewTeamUseCase` が残存している。呼び出し元が増えるにつれて誤った方を使うリスクが高まる。

---

## 受け入れ条件

- [x] `NewTeamUseCase` と `NewTeamUseCaseWithStats` が一本化され、コンストラクタが1つになっている
- [x] 既存のテストがすべて通る（`go test ./...`）
- [x] `golangci-lint` が通る

---

## サブチケット（コミット単位）

- [x] `refactor(usecase): TeamUseCase コンストラクタを一本化`

---

## 関連情報

- 関連チケット: TICKET-042（コンストラクタ並存が発生した起点）
- 備考: TICKET-042 の PR #32 で「次スプリント検討事項」として積み残したリファクタリング
