# TICKET-035 チームオーナー権限システム（バックエンド）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-035 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-20 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/team-owner-role-backend` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

`team_owner` 権限を新設し、管理者がユーザーへ付与・剥奪できるようにする。
チームオーナーは「作成できるチーム数上限」を持ち、上限内でチームを作成できる。

---

## 背景・目的

現状のロールは `admin` / `user` の2種のみ。チームを自律的に管理できる中間権限が必要。

---

## 受け入れ条件

- [ ] ユーザーエンティティに `is_team_owner: bool` と `max_teams: int` フィールドが追加されている
- [ ] 管理者が `PATCH /api/v1/admin/users/{id}/role` で `is_team_owner` と `max_teams` を設定できる
- [ ] チーム作成 API（`POST /api/v1/teams`）が `is_team_owner: true` のユーザーのみ実行可能
- [ ] チーム作成時に `max_teams` 上限チェックが行われ、超過時は 403 を返す
- [ ] `api/openapi.yaml` が更新されている
- [ ] ユースケース層のユニットテストが作成されている

---

## サブチケット（コミット単位）

- [ ] `feat(domain): ユーザーエンティティに is_team_owner・max_teams フィールドを追加`
- [ ] `feat(usecase): チームオーナー権限付与・剥奪ユースケースを実装`
- [ ] `feat(usecase): チーム作成時の team_owner 権限・上限チェックを追加`
- [ ] `feat(interface): 管理者向けロール変更エンドポイントを追加`
- [ ] `docs(api): openapi.yaml を更新`
- [ ] `test(usecase): チームオーナー権限ユースケースのユニットテストを作成`

---

## 関連情報

- 関連チケット: TICKET-036（招待機能）、TICKET-038（権限付与UI）
- 備考: `is_team_owner` は `role: "admin"` とは独立したフラグ。管理者は常にすべてのチームにアクセス可能
