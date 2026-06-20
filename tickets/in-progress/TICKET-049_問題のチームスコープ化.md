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

問題（Question）をチームスコープに変更する。各問題は1つのチームに属し、そのチームのメンバー以外からは参照不可とする。admin であっても、アクティブチームを切り替えない限り別チームの問題は閲覧不可。

---

## 背景・目的

現在の問題はグローバル（`VisibilityScope=all` がデフォルト）で管理されており、チームをまたいで問題が見えてしまう。タグがチームスコープ化されたのと同様に、問題もチームごとに管理する。

---

## 仕様

### 権限ルール
- チームメンバー全員：所属チームの問題を一覧・詳細閲覧・作成・編集・削除可能
- admin：**アクティブチームの問題のみ**閲覧・操作可能（別チームの問題は team_id を指定したアクセスが必要。admin特権で全チームにアクセス可能だが、UI上はアクティブチームの問題のみ表示）
- 非メンバー（別チームのユーザー）：403

### API 変更
- `GET    /api/v1/questions`               → `GET    /api/v1/teams/{team_id}/questions`
- `POST   /api/v1/questions`               → `POST   /api/v1/teams/{team_id}/questions`
- `GET    /api/v1/questions/{id}`          → `GET    /api/v1/teams/{team_id}/questions/{id}`
- `PUT    /api/v1/questions/{id}`          → `PUT    /api/v1/teams/{team_id}/questions/{id}`
- `DELETE /api/v1/questions/{id}`          → `DELETE /api/v1/teams/{team_id}/questions/{id}`
- `PUT    /api/v1/questions/{id}/visibility` → `PUT /api/v1/teams/{team_id}/questions/{id}/visibility`

### ドメインモデル変更
- `Question` に `TeamID string` を追加
- 既存の `VisibilityScope` / `PublishedTeamIDs` フィールドは**削除**（チームスコープで代替）
- `IsVisibleTo` ロジックをチームメンバーシップチェックに簡略化

### データ移行
- 既存の問題データ（`team_id` なし）はいったん**非表示**にする（`team_id` が空の問題は誰にも表示しない）

---

## 受け入れ条件

- [ ] チームAのメンバーがチームAの問題を作成・閲覧・編集・削除できる
- [ ] チームBのメンバーがチームAの問題にアクセスすると 403 になる
- [ ] admin がアクティブチームAの問題を閲覧でき、チームBに切り替えるとチームBの問題が表示される
- [ ] フロントエンドの全問題ページ（一覧・詳細・作成・編集）がアクティブチームのスコープで動作する
- [ ] `openapi.yaml` がチームスコープAPI定義に更新されている

---

## サブチケット（コミット単位）

- [ ] `feat(question): 問題をチームスコープ化（TICKET-049）`
- [ ] `test(e2e): 問題のチームスコープ化E2Eテスト（TICKET-049）`

---

## 関連情報

- 関連チケット: TICKET-048（タグのチームスコープ化）— 同じ設計パターンを踏襲
- 備考: `VisibilityScope` / `PublishedTeamIDs` はタグ同様に削除。チームスコープが公開制御の唯一の軸となる。

