# TICKET-007 問題コメント機能実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-007 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-17 |
| 着手日 | 2026-06-18 |
| 完了日 | 2026-06-18 |
| ブランチ名 | `feature/question-comment` |
| PR番号 | #10 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/10 |

---

## 概要

公開済み問題に対して、閲覧権限を持つユーザがコメントを投稿・閲覧できる機能を実装する。問題本文はRead Onlyだが、コメントは誰でも記入でき、投稿者の`display_name`を表示する。

---

## 背景・目的

勉強会での議論をアプリ内に残せるようにする。問題ページにコメント欄を設けることで、解釈の違いや補足情報をメンバー間で共有できる。

---

## Commentエンティティ定義

| フィールド | 型 | 説明 |
|---|---|---|
| `id` | string (UUID) | コメントID |
| `question_id` | string | 対象の問題ID |
| `body` | string | コメント本文（Markdown） |
| `created_by` | string | 投稿者のユーザID |
| `created_at` | datetime | 投稿日時 |
| `updated_at` | datetime | 更新日時 |

レスポンス時は `created_by` に対応する `display_name` を含めて返す。

---

## 受け入れ条件

- [x] `POST /api/v1/questions/{id}/comments` でコメントを投稿できる
- [x] `GET /api/v1/questions/{id}/comments` でコメント一覧を投稿日時の昇順で取得できる
- [x] `PUT /api/v1/questions/{id}/comments/{comment_id}` で自分のコメントを編集できる
- [x] `DELETE /api/v1/questions/{id}/comments/{comment_id}` で自分のコメントを削除できる（`admin` は全コメント削除可）
- [x] コメントの投稿・閲覧・編集・削除は、その問題の閲覧権限を持つユーザのみ可能
  - `published + all`: 全ログインユーザ
  - `published + team`: 指定チームのメンバーのみ
  - `draft` / `private`: 作成者本人のみ（実質コメント不可）
- [x] コメント一覧のレスポンスに投稿者の `display_name` を含む
- [x] `openapi.yaml` にコメントエンドポイントのSwagger定義が存在する
- [x] ユースケース層のユニットテストが作成されている
- [x] フロントエンドの問題詳細ページにコメント欄を追加する
  - コメント一覧（投稿者名・本文・投稿日時）を表示
  - コメント入力フォームを表示（Markdownプレビュー対応）
  - 自分のコメントに編集・削除ボタンを表示

---

## サブチケット（コミット単位）

- [x] `docs(api): コメントエンドポイントをopenapi.yamlに追加`
- [x] `feat(domain): Commentエンティティとバリデーションを実装`
- [x] `feat(usecase): コメントCRUDユースケースと閲覧権限チェックを実装`
- [x] `feat(interface): コメントハンドラとDTOを実装`
- [x] `feat(infrastructure): コメントのGCSリポジトリ実装`
- [x] `feat(frontend): 問題詳細ページにコメント欄を追加`
- [x] `test(usecase): コメントユースケースのユニットテストを作成`

---

## 関連情報

- 関連チケット: TICKET-002（認証・ユーザ情報取得）、TICKET-004（問題エンティティ）、TICKET-006（閲覧権限の判定ロジックを再利用）、TICKET-003（チームメンバーシップ判定）、TICKET-009（フロントエンド問題詳細ページの拡張）
- 備考: 問題削除時（TICKET-004）に紐づくコメントはカスケード削除する（問題と同時に全コメントを削除）
