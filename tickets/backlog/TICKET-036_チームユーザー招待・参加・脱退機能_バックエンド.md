# TICKET-036 チームユーザー招待・参加・脱退機能（バックエンド）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-036 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-20 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/team-invitation-backend` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

チームオーナーがユーザーを招待し、招待されたユーザーが参加/拒否を選択できるようにする。
また、ユーザー自身がチームから脱退できるようにする。

---

## 背景・目的

現状はチームメンバーシップの動的な変更手段がない。招待ベースのチーム参加フローが必要。

---

## 受け入れ条件

- [ ] 招待データ（`invitations`）が GCS 上で管理される
  - フィールド: `id`, `team_id`, `invited_by`, `invitee_identifier`（UUID/ユーザーID/eメール）, `status`（pending/accepted/rejected）, `created_at`
- [ ] `POST /api/v1/teams/{id}/invitations` で招待を送信できる（チームオーナーのみ）
- [ ] `GET /api/v1/invitations/me` でログインユーザー宛の招待一覧を取得できる
- [ ] `PATCH /api/v1/invitations/{id}` で招待を受諾/拒否できる（招待されたユーザーのみ）
- [ ] `DELETE /api/v1/teams/{id}/members/{user_id}` でチームオーナーがメンバーを退会させられる
- [ ] `DELETE /api/v1/teams/{id}/members/me` でユーザー自身が脱退できる
- [ ] `api/openapi.yaml` が更新されている
- [ ] ユースケース層のユニットテストが作成されている

---

## サブチケット（コミット単位）

- [ ] `feat(domain): 招待エンティティと招待リポジトリインターフェースを定義`
- [ ] `feat(usecase): 招待送信ユースケースを実装（UUID/ユーザーID/eメール対応）`
- [ ] `feat(usecase): 招待受諾・拒否ユースケースを実装`
- [ ] `feat(usecase): チームメンバー退会・脱退ユースケースを実装`
- [ ] `feat(infrastructure): 招待リポジトリ（GCS）を実装`
- [ ] `feat(interface): 招待・脱退エンドポイントを追加`
- [ ] `docs(api): openapi.yaml を更新`
- [ ] `test(usecase): 招待・参加・脱退ユースケースのユニットテストを作成`

---

## 関連情報

- 関連チケット: TICKET-035（チームオーナー権限）、TICKET-037（フロントエンド選択フロー）
- 備考: 招待は `invitee_identifier` に UUID / ユーザーID / eメールのいずれかを指定できる。既存メンバーへの招待はエラーにする
