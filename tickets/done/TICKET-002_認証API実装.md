# TICKET-002 認証API実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-002 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-17 |
| 着手日 | 2026-06-17 |
| 完了日 | 2026-06-17 |
| ブランチ名 | `feature/auth-api` |
| PR番号 | #2 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/2 |

---

## 概要

ID/パスワード認証とJWTトークン発行・検証のAPIを実装する。ロール（admin / teamowner / user）によるアクセス制御ミドルウェアを合わせて整備し、以降のAPIエンドポイントが認可を適用できる状態にする。

---

## 背景・目的

勉強会の問題・解説データは参加者のみがアクセスする想定のため、認証・認可の仕組みが必要。bcryptによるパスワードハッシュ化とJWT認証をInfrastructure層のミドルウェアとして実装し、セキュリティ要件を満たす。

---

## Userエンティティ定義

| フィールド | 型 | 説明 |
|---|---|---|
| `id` | string (UUID) | ユーザID |
| `username` | string | ログインID（一意） |
| `display_name` | string | UI表示名 |
| `email` | string | メールアドレス（一意） |
| `password_hash` | string | bcryptハッシュ済みパスワード（平文保存禁止） |
| `role` | string | `admin` / `teamowner` / `user` |
| `is_active` | bool | 有効ユーザかどうか（`false` = 停止中） |
| `created_at` | datetime | 作成日時 |
| `updated_at` | datetime | 更新日時 |

## ロール定義

| ロール | 権限 |
|---|---|
| `admin` | 全機能・全チーム管理・ユーザ停止 |
| `teamowner` | チーム作成・自チームのメンバー管理 |
| `user` | チーム参加・自身の問題CRUD |

---

## 受け入れ条件

- [x] `POST /api/v1/auth/login` でusername/パスワードを受け取り、JWTトークンを返す
- [x] `is_active: false` のユーザはログイン時に401を返す
- [x] `POST /api/v1/auth/logout` でトークンを無効化できる（またはクライアント側削除の仕様を明記）
- [x] パスワードはbcryptでハッシュ化して永続化し、平文は一切保存しない
- [x] JWTミドルウェアが保護エンドポイントでトークンを検証し、不正なら401を返す
- [x] 停止済みユーザ（`is_active: false`）のJWTは有効期限内でも403を返す
- [x] `admin` / `teamowner` / `user` ロール判定ミドルウェアが実装されており、権限外操作は403を返す
- [x] ユーザCRUD（`admin`のみ）: `GET/POST /api/v1/users`、`GET/PUT/DELETE /api/v1/users/{id}`
- [x] ユーザ停止・再有効化（`admin`のみ）: `PATCH /api/v1/users/{id}/status` で `is_active` を切り替えられる
- [x] `openapi.yaml` に認証エンドポイント・ユーザ管理エンドポイントのSwagger定義が存在する
- [x] ユースケース層のユニットテストが作成されている

---

## サブチケット（コミット単位）

- [x] `docs(api): 認証・ユーザー管理エンドポイントをopenapi.yamlに追加`
- [x] `feat(domain): Userエンティティとロール定義（admin / teamowner / user）を実装`
- [x] `feat(usecase): 認証ユースケース（ログイン・ユーザーCRUD・停止管理）を実装`
- [x] `feat(interface): 認証ハンドラとDTOを実装`
- [x] `feat(infrastructure): JWTミドルウェアとbcryptパスワード管理を実装`
- [x] `feat(infrastructure): ユーザーデータのGCSリポジトリ実装`
- [x] `test(usecase): 認証ユースケースのユニットテストを作成`

---

## 関連情報

- 関連チケット: TICKET-001（前提）、TICKET-003〜009（本チケット完了後に認可を適用）、TICKET-003（teamownerロールの利用）
- 参考: CLAUDE.md「認証」セクション
- 備考: 認証用ユーザーデータの永続化先（GCS上のJSONファイル or 別途DB）はTODO。本チケットではRepositoryインターフェースを定義し、GCS実装を作成するが、未決定事項として設計方針を記録する
