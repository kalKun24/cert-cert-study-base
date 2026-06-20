# TICKET-042 チームメンバー一覧画面

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-042 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-20 |
| 着手日 | 2026-06-20 |
| 完了日 | 2026-06-20 |
| ブランチ名 | `feature/team-member-list` |
| PR番号 | #32 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/32 |

---

## 概要

チームに所属するメンバーの一覧と活動統計（問題数・コメント数・最終ログイン）を表示する画面を新設する。

---

## 背景・目的

ホーム画面の「チームメンバー数」クリック時の遷移先として必要。チームの活動状況を把握するため。

---

## 受け入れ条件

- [x] `GET /api/v1/teams/{id}/members` がメンバーごとの統計を返す
  - フィールド: `user_id`, `display_name`, `role`, `is_team_owner`, `question_count`, `comment_count`, `last_login_at`
- [x] ユーザーエンティティに `last_login_at` フィールドが追加され、ログイン時に更新される
- [x] フロントエンドに `/teams/{id}/members` ページが存在する
- [x] 一覧には表示名・権限・問題数・コメント数・最終ログイン日時が表示される
- [x] `api/openapi.yaml` が更新されている

---

## サブチケット（コミット単位）

- [x] `feat(domain): ユーザーエンティティに last_login_at フィールドを追加`
- [x] `feat(usecase): 認証時に last_login_at を更新するロジックを追加`
- [x] `feat(usecase): チームメンバー統計取得ユースケースを実装`
- [x] `feat(interface): GET /api/v1/teams/{id}/members エンドポイントを追加`
- [x] `docs(api): openapi.yaml を更新`
- [x] `feat(page): チームメンバー一覧ページを作成`

---

## 関連情報

- 関連チケット: TICKET-041（ダッシュボード遷移元）
- 備考: テーブル・バッジ等のスタイルは `global.css` の CSS カスタムプロパティと既存の一覧画面のパターンに従う
