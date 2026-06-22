# TICKET-067 NoteCommentUseCase の userRepo=nil ガード除去とテストカバレッジ回復

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-067 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | 2026-06-22 |
| ブランチ名 | feature/TICKET-067 |
| PR番号 | #52 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/52 |

---

## 概要

`note_comment.go:49` のテスト用 nil ガードをプロダクションコードから除去し、モック UserRepository を注入したテストに置き換えることで、表示名解決パスのテストカバレッジを回復する。

---

## 背景・目的

`backend/internal/usecase/note_comment.go:49` に `if uc.userRepo == nil { return userID, nil }` という分岐が存在する。コメントにも「テスト用」と明記されているプロダクションコードへの分岐埋め込みであり、CLAUDE.md の DI 原則（「層の境界はinterfaceで定義し、具体実装はInfラstructure層に置く（DI）」）に反する。

`backend/internal/usecase/note_comment_test.go` が `userRepo=nil` を渡しているため、`resolveNoteCommentDisplayName` の実際の `FindByID` 呼び出しパス（正常系・ErrUserNotFound・その他エラー）が全くテストされておらず、CLAUDE.md の「特にユースケース層はユニットテストを必須とする」という要件を満たしていない。

---

## 受け入れ条件

- [x] `note_comment.go` の `if uc.userRepo == nil` ガードを除去する
- [x] `note_comment_test.go` に `mockUserRepo` 構造体を追加し、`nil` の代わりに注入する形に修正する
- [x] `resolveNoteCommentDisplayName` の正常系（ユーザー取得成功・display_name 返却）をテストする
- [x] `ErrUserNotFound` 時のフォールバック（userID を返す）をテストする
- [x] その他エラー時の挙動をテストする
- [x] `make test` が全てパスすることを確認する

---

## サブチケット（コミット単位）

- [x] `refactor(usecase): NoteCommentUseCase の userRepo nil ガードを除去`
- [x] `test(usecase): mockUserRepo を追加し display_name 解決パスのテストを実装`

---

## 関連情報

- 関連チケット: -
- 参考: -
- 備考: 次スプリント対応。検出エージェント: Code Reviewer（提案 #7, #8）。品質チェック日: 2026-06-22。
