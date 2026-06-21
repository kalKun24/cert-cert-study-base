# TICKET-057 ノートCRUD API実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-057 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-21 |
| 着手日 | 2026-06-21 |
| 完了日 | - |
| ブランチ名 | feature/note-crud-api |
| PR番号 | - |
| PRリンク | - |

---

## 概要

ノートの CRUD 操作（作成・一覧取得・詳細取得・更新・削除・公開設定変更・キーワード/タグ検索）を提供する REST API をバックエンドに実装する。`api/openapi.yaml` の更新を先行し、ユースケース層・GCS リポジトリ実装・ハンドラを順に実装する。

---

## 背景・目的

問題（Question）機能と同じアーキテクチャで、チームスコープのノートを管理できるようにする。チームメンバーが学習ノート・知識共有資料を投稿・管理できる基盤を提供する。

---

## 受け入れ条件

- [x] `api/openapi.yaml` に以下のエンドポイントが追加されている（API First）
  - `POST   /api/v1/teams/{team_id}/notes`
  - `GET    /api/v1/teams/{team_id}/notes`（ページネーション・キーワード・タグ検索対応）
  - `GET    /api/v1/teams/{team_id}/notes/{note_id}`
  - `PUT    /api/v1/teams/{team_id}/notes/{note_id}`
  - `PATCH  /api/v1/teams/{team_id}/notes/{note_id}/visibility`
  - `DELETE /api/v1/teams/{team_id}/notes/{note_id}`
- [x] レスポンス形式が `{ "data": ..., "error": ... }` の統一フォーマットに準拠している
- [x] `NoteUseCase` が以下のメソッドを持つ: `CreateNote`・`SearchNotes`・`ListNotes`・`GetNote`・`UpdateNote`・`UpdateNoteVisibility`・`DeleteNote`
- [x] 可視性ルールが問題と同じ: `draft` は作成者本人または admin のみ、`published`/`private` はチームメンバー全員
- [x] 編集権限: チームオーナー・admin・作成ユーザーのみが更新・削除可能。チームオーナーの判定は `teamRepo.IsOwner(ctx, teamID, callerID)` で行う（JWT クレームの `is_team_owner` フラグは使わない）
- [x] GCS リポジトリ実装が `backend/internal/infrastructure/repository/note_repository.go` に追加されている
- [x] GCS 上のオブジェクトパスが `teams/{team_id}/notes.json`（チーム別ファイル）に統一されている
- [x] ハンドラが `backend/internal/interface/handler/note_handler.go` に追加されている
- [x] ルーターにノート系エンドポイントが登録されている
- [x] `NoteUseCase` のユニットテストが追加されている（`usecase/note_test.go`）
- [x] `golangci-lint` を通過する

---

## サブチケット（コミット単位）

- [x] `docs(api): openapi.yamlにノートCRUDエンドポイントを追加`
- [x] `feat(usecase): NoteUseCaseを実装（CRUD・可視性ルール・ページネーション）`
- [x] `test(usecase): NoteUseCaseのユニットテストを追加`
- [x] `feat(infra): GCSノートリポジトリ実装を追加`
- [x] `feat(handler): ノートハンドラを追加しルーターに登録`

---

## 関連情報

- 関連チケット: TICKET-056（ドメイン定義、先行必須）、TICKET-058（ノートコメント）、TICKET-059（フロントエンド）
- 参考:
  - `backend/internal/usecase/question.go`
  - `backend/internal/infrastructure/repository/question_repository.go`
  - `api/openapi.yaml`（questions セクション）
- 備考:
  - 編集権限は「チームオーナー・admin・作成ユーザー」。チームオーナーの判定は `teamRepo.IsOwner(ctx, teamID, callerID)` でGCS（DB）にアクセスして行う。JWT クレームの `is_team_owner` フラグは使用しない（設計決定 #1）
  - GCS 上のパスは `teams/{team_id}/notes.json`（チーム別ファイル）に決定した（設計決定 #2）。これはノートとの一貫性確保のため問題（Question）も同様にチーム別へ移行する（TICKET-061）
  - 検索対象フィールドはタイトル・本文（`Body`）・議論点（`DiscussionPoints`）・メモ（`Memo`）とする
  - `NoteUseCase` はチームオーナー判定のために `TeamRepository` を依存として受け取る（DI）
