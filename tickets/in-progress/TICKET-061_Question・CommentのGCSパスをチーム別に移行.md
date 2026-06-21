# TICKET-061 Question・CommentのGCSパスをチーム別に移行

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-061 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-21 |
| 着手日 | 2026-06-21 |
| 完了日 | - |
| ブランチ名 | refactor/gcs-team-scoped-paths |
| PR番号 | #42 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/42 |

---

## 概要

問題（Question）のGCS格納パスを現在のグローバル単一ファイル方式（`questions.json`）からチーム別ファイル方式（`teams/{team_id}/questions.json`）に変更する。あわせて、問題コメント（Comment）のGCSパスも現在の `questions/{questionID}/comments/{commentID}.json` から `teams/{team_id}/questions/{questionID}/comments/{commentID}.json` へ移行する。これによりノート（Note）のGCSパス設計（`teams/{team_id}/notes.json`）との一貫性を確保する。

---

## 背景・目的

ノート機能（TICKET-057）でGCSパス設計を検討した結果、チームスコープのデータはすべて `teams/{team_id}/` プレフィックス配下に格納する方針に決定した（設計決定 2026-06-21）。現在の問題データはバケット直下の `questions.json`（全チーム混在の単一ファイル）に格納されており、チームフィルタリングをアプリ側で行っている。この方式はデータ量増加時のパフォーマンス低下とデータ分離の欠如という問題があるため、ノートと同じチーム別パス方式に統一する。

---

## 受け入れ条件

- [x] `GCSQuestionRepository` のGCSパスが `teams/{team_id}/questions.json` に変更されている
  - `loadQuestions` がグローバルな `questions.json` ではなく `teams/{teamID}/questions.json` を読み書きする
  - `saveQuestions` も同様に `teams/{teamID}/questions.json` へ書き込む
  - `ListByTeam`・`SearchByTeam` は teamID ごとのファイルを読むため、全件ロード後フィルタリングするロジックが不要になる（ただしファイルが存在しない場合は空リストを返す）
  - `FindByID`・`FindByTagID` は全チームを横断して検索する可能性があるため、実装方針を明示すること（例: `teamID` を引数に追加、または呼び出し元でチームを特定済みとしてルーティングする）
- [x] `GCSCommentRepository` のGCSパスが `teams/{team_id}/questions/{questionID}/comments/{commentID}.json` に変更されている
  - `commentObjectName` 関数のシグネチャを `commentObjectName(teamID, questionID, commentID string) string` に変更する
  - `commentPrefixByQuestion` 関数も `commentPrefixByQuestion(teamID, questionID string) string` に変更する
  - `domain.CommentRepository` インターフェースのメソッドシグネチャに `teamID` 引数が必要な場合は `domain/comment.go` も合わせて更新する
- [x] 既存のユースケース（`QuestionUseCase`・`CommentUseCase`）が新しいリポジトリシグネチャに対応している
- [x] 既存のハンドラがチームIDをリポジトリに渡す実装になっている
- [x] 既存のユニットテスト（`usecase/question_test.go` 等）が新しい実装に合わせて更新されている
- [x] `golangci-lint` を通過する

---

## サブチケット（コミット単位）

- [x] `refactor(infra): GCSQuestionRepositoryのパスをteams/{team_id}/questions.jsonに変更`
- [x] `refactor(infra): GCSCommentRepositoryのパスをteams/{team_id}/questions/{id}/comments/{id}.jsonに変更`
- [x] `refactor(usecase): QuestionUseCase・CommentUseCaseを新リポジトリシグネチャに対応`
- [x] `test(usecase): 既存ユニットテストを新しい実装に合わせて更新`

---

## 関連情報

- 関連チケット: TICKET-057（ノートCRUD API実装。本チケットとGCSパス設計を統一する）、TICKET-056（ノートドメイン定義）
- 参考:
  - `backend/internal/infrastructure/repository/question_repository.go`（現状: `questionsObjectName = "questions.json"`、チームフィルタリングをアプリ側で実施）
  - `backend/internal/infrastructure/repository/comment_repository.go`（現状: `questions/{questionID}/comments/{commentID}.json`）
- 備考:
  - **データマイグレーション**: 本番環境（GCS）に既存データが存在する場合、`questions.json` の全レコードをチームIDごとに分割して `teams/{team_id}/questions.json` に書き直す移行スクリプトが必要。コメントも同様に `questions/{id}/comments/{id}.json` を `teams/{team_id}/questions/{id}/comments/{id}.json` へ移行するスクリプトが必要。現時点でローカル開発は `fake-gcs-server`（インメモリ）のため本番デプロイ前に対応すること
  - `FindByID`・`FindByTagID` の全チーム横断検索の扱いは、呼び出し元のハンドラ・ユースケースで常に `teamID` が確定している（チームメンバーしかアクセスできないため）ので、引数に `teamID` を追加する方針で統一する
  - 変更後は全チームの問題が `teams/{team_id}/questions.json` に分散されるため、`ListByTeam` はファイルをそのまま全件返せばよくなり、不要なアプリ側フィルタリングがなくなる
