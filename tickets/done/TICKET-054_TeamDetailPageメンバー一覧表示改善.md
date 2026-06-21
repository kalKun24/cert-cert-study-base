# TICKET-054 TeamDetailPage メンバー一覧表示改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-054 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-21 |
| 着手日 | 2026-06-21 |
| 完了日 | 2026-06-21 |
| ブランチ名 | `feature/TICKET-054` |
| PR番号 | #39 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/39 |

---

## 概要

チーム詳細画面（`TeamDetailPage`）のメンバー一覧を、ユーザーID（UUID）表示から `GET /api/v1/teams/{id}/members` のレスポンス（`TeamMemberStatsDTO`）を利用したリッチな表示に変更する。

---

## 背景・目的

現在のチーム詳細画面ではメンバー一覧にUUIDしか表示されず、ユーザーにとって誰がメンバーなのかが全く判別できない。`TeamMemberListPage`（メンバー統計一覧画面）と同じデータ形式を用いることで、表示名・ロール・問題数・コメント数・最終ログイン日時を確認できるようにする。また、チームオーナーまたは admin のみがロール変更・メンバー除名ボタンを表示できるように権限分岐を実装する。

---

## 受け入れ条件

- [x] チーム詳細画面のメンバー一覧が `GET /api/v1/teams/{id}/members` を呼び出して取得したデータを表示する
- [x] 表示項目: 表示名（display_name）、ロール（role）、問題数（question_count）、コメント数（comment_count）、最終ログイン日時（last_login_at）
- [x] チームオーナーまたは admin ロールのユーザーのみ、ロール変更ボタン・メンバー除名ボタンが表示される
- [x] 一般メンバーには編集ボタン類が非表示（閲覧のみ）
- [x] UI文字列はすべて `frontend/src/locales/ja.json` で管理されている
- [x] `TeamMemberListPage` の参考実装と一貫したスタイルで表示される

---

## サブチケット（コミット単位）

- [x] `feat(team-detail): TeamDetailPage のメンバー一覧を TeamMemberStatsDTO で表示`
- [x] `feat(team-detail): 権限に応じたロール変更・メンバー除名ボタンの表示分岐`
- [x] `feat(locales): TeamDetailPage メンバー一覧の日本語文字列を ja.json に追加`

---

## 関連情報

- 関連チケット: TICKET-051（TeamDetailPage からメンバー一覧への導線追加）
- 参考実装: `frontend/src/pages/TeamMemberListPage.tsx`
- API:
  - `GET /api/v1/teams/{id}/members` → `TeamMemberStatsDTO[]`
  - `PATCH /api/v1/teams/{id}/members/{user_id}/role`
  - `DELETE /api/v1/teams/{id}/members/{user_id}`
- 備考: 文字列は必ず `frontend/src/locales/ja.json` で管理すること
