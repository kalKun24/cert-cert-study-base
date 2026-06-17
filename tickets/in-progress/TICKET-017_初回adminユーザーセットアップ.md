# TICKET-017 初回 admin ユーザーセットアップ

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-017 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-17 |
| 着手日 | 2026-06-18 |
| 完了日 | - |
| ブランチ名 | `feature/admin-seed` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

アプリ初回起動時に admin ユーザーが存在しない場合、環境変数から初期 admin を自動生成する仕組みを実装する。これがないとログインできるユーザーが存在しない状態になる。

---

## 背景・目的

TICKET-002 で認証 API は完成しているが、ユーザーを作成できるのは admin のみという鶏と卵の問題がある。初回起動時に admin を安全にブートストラップする手段が必要。

---

## 方針

サーバー起動時に GCS 上のユーザーストアが空の場合、環境変数で指定された初期 admin を自動作成する。

```
SEED_ADMIN_USERNAME=admin
SEED_ADMIN_PASSWORD=（起動前に Secret Manager から注入）
SEED_ADMIN_EMAIL=admin@example.com
SEED_ADMIN_DISPLAY_NAME=Administrator
```

- 環境変数が未設定の場合は seed をスキップする（通常運用時は設定しない）
- ユーザーストアに 1 件でもユーザーが存在する場合は seed をスキップする（冪等）
- パスワードは bcrypt でハッシュ化してから保存する

---

## 受け入れ条件

- [ ] 環境変数 `SEED_ADMIN_*` が設定された状態でサーバー起動すると、ユーザーストアが空の場合のみ admin が作成される
- [ ] ユーザーが 1 件以上存在する場合は seed をスキップし、エラーにならない（冪等）
- [ ] 環境変数未設定時は seed 処理をスキップし、通常起動する
- [ ] seed されたユーザーで `POST /api/v1/auth/login` が成功する
- [ ] `docker-compose.yml` と `.env.example` に `SEED_ADMIN_*` の定義が追加されている
- [ ] Cloud Run の環境変数設定手順が README に記載されている

---

## サブチケット（コミット単位）

- [x] `feat(infrastructure): 起動時 admin seed 処理を実装`
- [x] `chore(config): SEED_ADMIN_* 環境変数を docker-compose.yml・.env.example に追加`
- [x] `docs(readme): 初回セットアップ手順（admin 作成）を記載`

---

## 関連情報

- 関連チケット: TICKET-002（認証 API）、TICKET-010（GCP インフラ / Secret Manager）
- 備考: 本番環境では `SEED_ADMIN_PASSWORD` を Secret Manager から注入し、初回起動後は環境変数を削除する運用を推奨
