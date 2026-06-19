# TICKET-031 ユーザー管理画面UI改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-031 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-031` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

ユーザー管理画面（UserListPage / UserCreatePage / UserEditPage）にスタイルを適用し、UserTable のテーブルデザイン・ステータスバッジ・ユーザー作成・編集フォームの UI を改善する。

---

## 背景・目的

ユーザー管理画面は admin ユーザーのみがアクセスできる管理機能。現状はスタイルがなく、UserTable が裸の HTML テーブル状態。ユーザーステータス（有効/停止中）のバッジや、ロールバッジの視覚的区別がない。ユーザー作成・編集フォーム（UserForm）にもスタイルの適用が必要。

---

## 受け入れ条件

### ユーザー一覧 (UserListPage)
- [ ] ページヘッダー（タイトル + 「ユーザーを作成」ボタン）が横並びで表示される
- [ ] UserTable が TICKET-022 の `.user-table` スタイルで表示される
  - テーブルヘッダー: 薄いグレー背景・大文字スモール・セカンダリ色
  - データ行: hover 時に背景色変化
  - `overflow-x: auto` ラッパーでモバイル横スクロール対応
- [ ] ユーザーのステータス（有効/停止中）がバッジで視覚的に区別される
  - 有効: グリーン系バッジ（`.status-badge-active`）
  - 停止中: グレー系バッジ（`.status-badge-inactive`）
- [ ] ロールバッジ（管理者/チームオーナー/ユーザー）が色で区別される
  - 管理者: 赤系（`data-role="admin"` スタイル）
  - その他: 青系
- [ ] 操作列のボタン（編集・ステータストグル・削除）が整然と並ぶ
- [ ] 空状態・ローディング・エラー状態がスタイルされている

### ユーザー作成・編集 (UserCreatePage / UserEditPage)
- [ ] UserForm コンポーネントにスタイルが適用されている
  - フォームがカード（または max-width: 500px のコンテナ）で包まれている
  - ユーザー名・表示名・メール・ロール選択・パスワードフィールドが整然と並ぶ
  - 「保存」「キャンセル」ボタンが横並び
- [ ] パスワードフィールドに表示/非表示トグルが追加されている（TICKET-023 で作成の PasswordInput コンポーネントを利用）
- [ ] バリデーションエラーが各フィールド下に赤テキストで表示される
- [ ] ローディング・エラー状態がスタイルされている

---

## サブチケット（コミット単位）

- [ ] `feat(user-list): UserTableのスタイルを適用（テーブル・バッジ・操作ボタン）`
- [ ] `feat(user-list): ステータスバッジとロールバッジのスタイルを適用`
- [ ] `feat(user-form): UserFormコンポーネントのスタイルを適用`
- [ ] `feat(user-form): パスワードフィールドに表示/非表示トグルを追加（PasswordInputコンポーネント利用）`

---

## 関連情報

### UserTable の変更点

`UserTable.tsx` コンポーネントは現在 `UserStatusBadge`, `UserToggleStatusButton`, `UserDeleteButton` を使用している。各コンポーネントの出力に適切なクラスを付与する必要がある。

UserStatusBadge の変更例:

```tsx
// UserStatusBadge.tsx
// 現状の実装を確認して status に応じた class 名を付与する
<span
  className={`status-badge ${user.is_active ? 'status-badge-active' : 'status-badge-inactive'}`}
>
  {user.is_active ? '有効' : '停止中'}
</span>
```

### PasswordInput コンポーネントの活用

TICKET-023 で作成した `PasswordInput` コンポーネントを UserForm 内のパスワードフィールドに適用する。

```tsx
// UserForm.tsx 内（create mode のパスワードフィールド）
import PasswordInput from './PasswordInput';

<PasswordInput
  id="user-password"
  label={t('user.form.passwordLabel')}
  value={password}
  onChange={setPassword}
  autoComplete="new-password"
  disabled={isSubmitting}
/>
```

### スタイル仕様

```css
/* ユーザーフォームページ */
.user-form-page { max-width: 560px; }

/* ユーザーフォームをカードで包む */
.user-form-page .user-form-card {
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  box-shadow: var(--shadow-card);
}

/* .user-table は TICKET-022 の global.css で定義済み */
/* .status-badge は TICKET-022 の global.css で定義済み */
```

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 関連チケット: TICKET-023（PasswordInput コンポーネント）が作成済みであること
