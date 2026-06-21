# TICKET-056 ノートドメイン定義とリポジトリインターフェース

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-056 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-21 |
| 着手日 | 2026-06-21 |
| 完了日 | - |
| ブランチ名 | feature/note-domain |
| PR番号 | - |
| PRリンク | - |

---

## 概要

知識共有ノート機能のドメイン層を定義する。`domain.Note` エンティティ・ステータス型・ドメインエラー・`NoteRepository` インターフェースを追加し、後続のユースケース層・インフラ層が依存できる土台を整える。

---

## 背景・目的

問題（Question）機能と同様、ノートもチームスコープで管理し、下書き・非公開・公開の3段階のステータスを持つ。問題機能との構造的な類似を維持しつつ、ノートとしての意味論（本文・議論点・メモ）を独自のエンティティで表現する。ドメイン層を先に確定させることで、バックエンド実装（TICKET-057）とフロントエンド実装（TICKET-059）を並行して進められる状態にする。

---

## 受け入れ条件

- [ ] `backend/internal/domain/note.go` が追加されている
- [ ] `NoteStatus` 型（`draft` / `private` / `published`）と `IsValid()` メソッドが定義されている
- [ ] `Note` エンティティが以下のフィールドを持つ: `ID`・`TeamID`・`Title`・`Body`（Markdown）・`DiscussionPoints`（Markdown）・`Memo`（Markdown）・`Tags`（`[]string`）・`Status`・`CreatedBy`・`CreatedAt`・`UpdatedAt`
- [ ] `ErrNoteNotFound`・`ErrInvalidNoteStatus` のドメインエラーが定義されている
- [ ] `NoteRepository` インターフェースが定義されており、以下のメソッドを持つ: `FindByID`・`ListByTeam`・`SearchByTeam`・`Save`・`Delete`・`FindByTagID`
- [ ] `NoteSearchFilter` が `TagIDs []string` と `Keyword string` を持つ
- [ ] `NoteComment` エンティティが `backend/internal/domain/note.go`（または同パッケージの `note_comment.go`）に定義されている。フィールド: `ID`・`NoteID`・`Body`・`CreatedBy`・`CreatedAt`・`UpdatedAt`
- [ ] `ErrNoteCommentNotFound` のドメインエラーが定義されている
- [ ] `NoteCommentRepository` インターフェースが定義されており、以下のメソッドを持つ: `FindByID`・`ListByNoteID`・`Save`・`Delete`
- [ ] ユニットテストが不要なことを確認した（ドメイン層は純粋な型定義のみ）
- [ ] `golangci-lint` を通過する

---

## サブチケット（コミット単位）

- [ ] `feat(domain): Noteエンティティ・ステータス型・ドメインエラーを定義`
- [ ] `feat(domain): NoteRepositoryインターフェースとNoteSearchFilterを定義`
- [ ] `feat(domain): NoteCommentエンティティ・ドメインエラー・NoteCommentRepositoryインターフェースを定義`

---

## 関連情報

- 関連チケット: TICKET-057（ユースケース・API）、TICKET-058（ノートコメント）、TICKET-059（フロントエンド）
- 参考: `backend/internal/domain/question.go`（Question エンティティをほぼ踏襲する）
- 備考:
  - ノートの `Body` フィールドは問題の `Body` に相当する（本文）
  - `DiscussionPoints` は問題の `Answer` / `Explanation` に相当する議論点フィールド
  - `Memo` は問題の `Memo` と同じ扱い（自由記述メモ）
  - **設計決定（2026-06-21）**: コメントは `NoteComment` を新規エンティティとして定義する。既存の `domain.Comment`（`QuestionID` フィールドを持つ）は変更しない。`NoteComment` は `NoteID` フィールドを持ち、`domain.Comment` とは完全に独立した型とする
