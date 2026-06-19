# TICKET-004 問題CRUD API実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-004 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-17 |
| 着手日 | 2026-06-18 |
| 完了日 | 2026-06-18 |
| ブランチ名 | `feature/question-crud-api` |
| PR番号 | #7 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/7 |

---

## 概要

勉強会の問題・解答・解説・議論点メモをMarkdown形式で作成・取得・更新・削除するREST APIを実装する。データはGCS（Google Cloud Storage）に永続化する。

---

## 背景・目的

本アプリケーションのコア機能。CISSP・情報処理安全確保支援士の問題をMarkdown形式で蓄積・管理できるようにする。クリーンアーキテクチャに従い、domain → usecase → interface → infrastructure の順で実装する。

---

## 受け入れ条件

- [x] `POST /api/v1/questions` で問題を作成し、GCSに保存できる
- [x] `GET /api/v1/questions` で問題一覧を取得できる（ページネーション対応はTICKET-009で行う）
- [x] `GET /api/v1/questions/{id}` で問題詳細（Markdownテキスト含む）を取得できる
- [x] `PUT /api/v1/questions/{id}` で問題を更新できる
- [x] `DELETE /api/v1/questions/{id}` で問題を削除できる（紐づくコメントもカスケード削除する）
- [x] 問題エンティティは `id` / `title` / `body`（Markdown） / `answer`（Markdown） / `explanation`（Markdown） / `memo`（Markdown） / `tags` / `status`（`draft` / `private` / `published`） / `visibility_scope`（`all` / `team`） / `published_team_ids`（チームIDの配列） / `created_by` / `created_at` / `updated_at` を持つ
- [x] 作成時のデフォルトは `status: draft`、`visibility_scope: all`（TICKET-006と連携）
- [x] 認証済みユーザー（`user`以上）のみ操作可能（TICKET-002の認証ミドルウェアを適用）
- [x] `openapi.yaml` に問題CRUDエンドポイントのSwagger定義が存在する
- [x] ユースケース層のユニットテストが作成されている

---

## サブチケット（コミット単位）

- [x] `docs(api): 問題CRUDエンドポイントをopenapi.yamlに追加`
- [x] `feat(domain): Questionエンティティとバリデーションを実装`
- [x] `feat(usecase): 問題CRUDユースケースを実装`
- [x] `feat(interface): 問題CRUDハンドラとDTOを実装`
- [x] `feat(infrastructure): 問題のGCSリポジトリ実装`
- [x] `test(usecase): 問題CRUDユースケースのユニットテストを作成`

---

## 関連情報

- 関連チケット: TICKET-001（前提）、TICKET-002（認証ミドルウェア適用のため前提）、TICKET-005（タグはQuestionエンティティに含まれる）、TICKET-009（フロントエンド連携）、TICKET-006（statusフィールドの詳細定義・ステータス管理ロジック）
- 参考: CLAUDE.md「クリーンアーキテクチャ」「REST API」セクション
- 備考: GCSのローカルエミュレータ方式が未決定のため、Repositoryインターフェースとモックでユニットテストをパスできるよう設計する
