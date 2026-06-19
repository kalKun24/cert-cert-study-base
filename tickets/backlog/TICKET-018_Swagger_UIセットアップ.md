# TICKET-018 Swagger UIセットアップ

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-018 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/swagger-ui` |
| PR番号 | - |
| PRリンク | （PR作成後に記入） |

---

## 概要

`make swagger` が現在 TODO のままになっている。`api/openapi.yaml` を Swagger UI で閲覧できるようにし、開発時の API 確認・手動テストを効率化する。

---

## 背景・目的

CLAUDE.md に「Swagger UIは開発環境の `/swagger/` で確認」と定義されているが、`Makefile` の `swagger` ターゲットが `TODO` のままであり未実装。

---

## 受け入れ条件

- [ ] `http://localhost:8080/swagger/` を開くと `api/openapi.yaml` の内容が表示される
- [ ] Swagger UI 上から `Authorize` ボタンで JWT を設定し、保護されたエンドポイントを実行できる
- [ ] `make swagger` を実行すると Swagger UI が起動する

---

## サブチケット（コミット単位）

- [ ] `chore(backend): /swagger/ エンドポイントで openapi.yaml を静的 serve する`
- [ ] `chore(makefile): make swagger ターゲットを実装`

---

## 関連情報

- 関連チケット: TICKET-001（プロジェクト基盤構築）、TICKET-002（認証API）
- 備考: `swaggo/swag` のコード生成は使わず、既存の `api/openapi.yaml` を静的に serve する方式で実装する
