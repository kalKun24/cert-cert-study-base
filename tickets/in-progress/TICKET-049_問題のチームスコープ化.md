# TICKET-049 問題のチームスコープ化

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-049 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-20 |
| 着手日 | 2026-06-20 |
| 完了日 | - |
| ブランチ名 | `feature/team-scoped-questions` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

問題（Question）をチームスコープに変更する。各問題は1つのチームに属し、
チームメンバー以外（admin含む）はアクセス不可とする。

---

## 背景・目的

タグがチームスコープ化（TICKET-048）されたが、問題はまだグローバル管理のため
別チームのユーザーにも見えてしまっている。問題もチームごとに管理する。

タグとの違い: admin も対象チームのメンバーでなければ 403 を返す（タグは admin 特権スキップあり）。

---

## 仕様

- `Question` に `team_id` フィールドを追加
- `visibility_scope` / `published_team_ids` を廃止（チームメンバーシップで代替）
- `status` (draft/published/private) は維持（チーム内でのみ意味を持つ）
- API を `/api/v1/teams/{team_id}/questions` に移行
- コメント API も `/api/v1/teams/{team_id}/questions/{id}/comments` に移行
- 全ユーザー（admin含む）がチームメンバーでなければ 403

---

## サブチケット（コミット単位）

- [ ] `refactor(question): VisibilityScope廃止・TeamID追加・リポジトリIF変更（domain層）`
- [ ] `refactor(question): チームスコープ対応・usecase層更新（テスト含む）`
- [ ] `refactor(question): チームスコープ対応・handler/dto/repository/routing更新`
- [ ] `refactor(question): フロントエンドをチームスコープAPI対応に更新`
- [ ] `docs(api): openapi.yamlの問題エンドポイントをチームスコープに更新`

---

## 関連情報

- 関連チケット: TICKET-048（タグのチームスコープ化）
- 備考: 既存 GCS データに team_id がない場合は空文字として読み込まれる（手動マイグレーション別途）
