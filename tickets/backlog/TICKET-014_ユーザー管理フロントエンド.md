# TICKET-014 ユーザー管理フロントエンド（admin 向け）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-014 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-17 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/frontend-user-admin` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

TICKET-002（ユーザー管理 API）に対応する admin 向けユーザー管理画面を実装する。ユーザーの作成・編集・停止・削除を GUI で操作できるようにする。

---

## 前提条件

- TICKET-002（認証 API）が完了していること（✅ 完了済み）
- TICKET-011（フロントエンド認証基盤）が完了していること
- ログインユーザーのロールが `admin` であること（admin のみアクセス可）

---

## 実装スコープ

### 画面一覧

| 画面 | パス | 対象ロール |
|---|---|---|
| ユーザー一覧 | `/admin/users` | admin のみ |
| ユーザー作成 | `/admin/users/new` | admin のみ |
| ユーザー編集 | `/admin/users/:id/edit` | admin のみ |

### 主要コンポーネント

- `UserTable`：ユーザー一覧テーブル（username・display_name・role・is_active・作成日）
- `UserStatusBadge`：有効 / 停止中のバッジ
- `UserForm`：ユーザー作成・編集フォーム（username・display_name・email・role・password）
- `UserToggleStatusButton`：有効化 / 停止切り替えボタン（確認ダイアログ付き）
- `UserDeleteButton`：削除ボタン（確認ダイアログ付き）

---

## 受け入れ条件

- [ ] admin のみ `/admin/users` にアクセスできる（他ロールはアクセス時に 403 / リダイレクト）
- [ ] ユーザー一覧が表示され、有効 / 停止中の状態が視覚的に区別できる
- [ ] ユーザー作成フォームから新規ユーザーを作成できる
- [ ] ユーザー情報（display_name / email / role）を編集できる
- [ ] 有効化・停止の切り替えができる（確認ダイアログあり）
- [ ] ユーザー削除ができる（確認ダイアログあり）

---

## サブチケット（コミット単位）

- [ ] `feat(frontend): ユーザー一覧画面を実装`
- [ ] `feat(frontend): ユーザー作成・編集フォームを実装`
- [ ] `feat(frontend): ユーザー有効化・停止・削除操作を実装`
- [ ] `feat(frontend): admin ルートガード（非 admin のアクセス制御）を実装`

---

## 関連情報

- 関連チケット: TICKET-002（ユーザー管理 API・✅ 完了済み）、TICKET-011（フロントエンド認証基盤）
- 備考: TICKET-002 が完了済みのため、TICKET-011 完了後すぐに着手可能
