# TICKET-058 ノートコメントAPI実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-058 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-21 |
| 着手日 | 2026-06-21 |
| 完了日 | 2026-06-22 |
| ブランチ名 | feature/note-comment-api |
| PR番号 | #45 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/45 |

---

## 概要

ノートへのコメント投稿・一覧取得・編集・削除 API をバックエンドに実装する。既存の問題コメント（`CommentUseCase`）との共有化、または専用ユースケースの追加を検討し、チームメンバーがノートにコメントできる仕組みを提供する。

---

## 背景・目的

ノート機能でも問題と同様にチーム内でコメントを投稿して議論できる。`domain.Comment` エンティティには現在 `QuestionID` フィールドがあり、ノートコメントへの流用可否を判断してから実装する。

---

## 受け入れ条件

- [x] `api/openapi.yaml` に以下のエンドポイントが追加されている
  - `POST   /api/v1/teams/{team_id}/notes/{note_id}/comments`
  - `GET    /api/v1/teams/{team_id}/notes/{note_id}/comments`
  - `PUT    /api/v1/teams/{team_id}/notes/{note_id}/comments/{comment_id}`
  - `DELETE /api/v1/teams/{team_id}/notes/{note_id}/comments/{comment_id}`
- [x] コメント投稿者本人のみ編集可能、削除は投稿者本人または admin のみ可能（問題コメントと同じルール）
- [x] チームメンバーであり、ノートの閲覧権限（`draft` ならば作成者本人または admin）を持つユーザーのみコメント可能
- [x] コメント一覧が投稿日時の昇順で返される
- [x] コメント本文の空チェック・最大長チェックが行われる
- [x] ユースケースのユニットテストが追加されている
- [x] `golangci-lint` を通過する

---

## サブチケット（コミット単位）

- [x] `docs(api): openapi.yamlにノートコメントエンドポイントを追加`
- [x] `feat(usecase): NoteCommentUseCaseを実装（NoteCommentエンティティを使用）`
- [x] `test(usecase): NoteCommentUseCaseのユニットテストを追加`
- [x] `feat(infra): GCSノートコメントリポジトリ実装を追加（NoteCommentRepositoryインターフェースを実装）`
- [x] `feat(handler): ノートコメントハンドラを追加しルーターに登録`

---

## 関連情報

- 関連チケット: TICKET-056（ドメイン定義、先行必須）、TICKET-057（ノートCRUD、先行必須）、TICKET-059（フロントエンド）
- 参考:
  - `backend/internal/domain/comment.go`（参考のみ。変更しない）
  - `backend/internal/usecase/comment.go`（参考のみ。変更しない）
  - `backend/internal/infrastructure/repository/comment_repository.go`（GCS パス: `questions/{questionID}/comments/{commentID}.json`）
- 備考:
  - **設計決定（2026-06-21）**: `NoteComment` を新規エンティティとして実装する。既存の `domain.Comment`（`QuestionID` フィールドを持つ）は変更しない。`Comment` との共用は行わない
  - `NoteComment` エンティティおよび `NoteCommentRepository` インターフェースの定義は TICKET-056 で行う。このチケットはユースケース層以降の実装を担当する
  - GCS オブジェクトパスは `teams/{team_id}/notes/{noteID}/comments/{commentID}.json` を想定（ノートがチーム別パス `teams/{team_id}/` 配下に格納されるため、コメントも同じプレフィックスに揃える）
