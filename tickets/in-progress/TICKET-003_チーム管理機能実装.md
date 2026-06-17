# TICKET-003 チーム管理機能実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-003 |
| ステータス | 🟡 進行中 |
| 作成日 | 2026-06-17 |
| 着手日 | 2026-06-18 |
| 完了日 | - |
| ブランチ名 | `feature/team-management` |
| PR番号 | - |
| PRリンク | （PR作成後に記入） |

---

## 概要

チームの作成・管理・メンバー招待機能を実装する。チーム単位でMarkdown問題の公開範囲を制御するための基盤となる。

---

## 背景・目的

勉強会を複数のチーム（例: 資格種別・社内グループ）で運用できるようにする。`admin` または `teamowner` ロールを持つユーザがチームを自由に作成し、メンバーを招待することで、問題の公開範囲をチーム単位で制御できるようにする（TICKET-006と連携）。

---

## ロール定義（TICKET-002 と共通）

| ロール | 権限 |
|---|---|
| `admin` | 全機能・全チームの管理 |
| `teamowner` | チームの作成・自チームのメンバー管理 |
| `user` | チームへの参加・問題CRUD（自身のコンテンツ） |

---

## エンティティ定義

### Team

| フィールド | 型 | 説明 |
|---|---|---|
| `id` | string (UUID) | チームID |
| `name` | string | チーム名（一意） |
| `description` | string | チームの説明（任意） |
| `owner_id` | string | 作成者ユーザID |
| `created_at` | datetime | 作成日時 |
| `updated_at` | datetime | 更新日時 |

### TeamMember

| フィールド | 型 | 説明 |
|---|---|---|
| `team_id` | string | チームID |
| `user_id` | string | メンバーのユーザID |
| `joined_at` | datetime | 参加日時 |

---

## 受け入れ条件

- [ ] `POST /api/v1/teams` でチームを作成できる（`admin` / `teamowner` のみ）
- [ ] `GET /api/v1/teams` で自分が所属するチームの一覧を取得できる（`admin` は全チーム取得可）
- [ ] `GET /api/v1/teams/{id}` でチーム詳細・メンバー一覧を取得できる（メンバーまたは `admin` のみ）
- [ ] `PUT /api/v1/teams/{id}` でチーム情報を更新できる（`owner_id` 本人または `admin` のみ）
- [ ] `DELETE /api/v1/teams/{id}` でチームを削除できる（`owner_id` 本人または `admin` のみ）
- [ ] `POST /api/v1/teams/{id}/members` で指定ユーザをチームに招待できる（`owner_id` 本人または `admin` のみ）
- [ ] `DELETE /api/v1/teams/{id}/members/{user_id}` でメンバーを除外できる（`owner_id` 本人または `admin` のみ）
- [ ] `openapi.yaml` にチーム管理エンドポイントのSwagger定義が存在する
- [ ] ユースケース層のユニットテストが作成されている

---

## サブチケット（コミット単位）

- [ ] `docs(api): チーム管理エンドポイントをopenapi.yamlに追加`
- [ ] `feat(domain): Team・TeamMemberエンティティとバリデーションを実装`
- [ ] `feat(usecase): チームCRUD・メンバー管理ユースケースを実装`
- [ ] `feat(interface): チームハンドラとDTOを実装`
- [ ] `feat(infrastructure): Team・TeamMemberのGCSリポジトリ実装`
- [ ] `test(usecase): チーム管理ユースケースのユニットテストを作成`

---

## 関連情報

- 関連チケット: TICKET-002（`teamowner` ロール追加）、TICKET-006（チーム単位の公開範囲制御）、TICKET-009（チーム管理UI）
- 備考: チーム削除時にそのチーム宛ての `published_team_ids` をどう扱うか（問題を自動的に `private` に戻すか）は着手前に設計を確認する
