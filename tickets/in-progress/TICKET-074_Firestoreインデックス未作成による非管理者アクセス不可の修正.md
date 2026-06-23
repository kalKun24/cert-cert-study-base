# TICKET-074 Firestore インデックス未作成による非管理者アクセス不可の修正

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-074 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-23 |
| 着手日 | 2026-06-23 |
| 完了日 | - |
| ブランチ名 | `fix/firestore-missing-indexes` |
| PR番号 | #72 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/72 |

---

## 概要

GCP 本番環境で非管理者ユーザーが問題の閲覧・作成・編集を行えない不具合を修正する。
Firestore のコレクショングループインデックスおよび複合インデックスが未作成のため、
非管理者向けの `GET /api/v1/teams` が 500 エラーとなり、チーム一覧が取得できないことが原因。

---

## 背景・目的

- `ListByOwnerOrMember` が `collectionGroup("members").Where("user_id", ...)` を使用するが、
  コレクショングループインデックスが本番 Firestore に未作成
- Firestore エミュレータはインデックスを強制しないためローカルでは問題が出なかった
- 管理者は `List()` を使うためインデックス不要で問題なし → GCP のみ非管理者が詰まる

不足インデックス:
1. `collectionGroup("members").where("user_id")` — チーム一覧（非 admin）
2. `collectionGroup("tags").where("id")` — タグ ID 検索
3. `invitations.(team_id + invitee_user_id + status)` — 重複招待チェック

---

## 受け入れ条件

- [ ] `firestore.indexes.json` がリポジトリルートに作成されている
- [ ] CD ワークフローで Firestore インデックスが自動デプロイされる
- [x] GCP 本番環境で非管理者ユーザーがチーム一覧を取得できる（インデックス直接作成で解消確認済み）
- [ ] GCP 本番環境で非管理者ユーザーが問題の閲覧・作成・編集ができる（PR マージ後に確認）

---

## サブチケット（コミット単位）

- [x] `fix(firestore): コレクショングループ・複合インデックスを定義する firestore.indexes.json を追加`
- [x] `fix(cd): CD ワークフローに Firestore インデックス自動デプロイステップを追加`
- [x] `fix(cd): 単一フィールド COLLECTION_GROUP インデックスを REST API で作成するよう修正`
- [x] `fix(usecase): ListMemberStats で孤立 members レコードをスキップするよう修正`

---

## 関連情報

- 関連チケット: TICKET-073（ローカル・GCP 環境差異の修正）
- 該当コード: `backend/internal/infrastructure/firestore/team_repository.go:167`
- 該当コード: `backend/internal/infrastructure/firestore/tag_repository.go:67`
- 該当コード: `backend/internal/infrastructure/firestore/invitation_repository.go:111`
- 備考: インデックスは非同期で構築される（数分かかることがある）
