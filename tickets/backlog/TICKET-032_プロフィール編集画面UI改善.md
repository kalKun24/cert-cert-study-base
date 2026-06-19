# TICKET-032 プロフィール編集画面UI改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-032 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-032` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

プロフィール編集画面（ProfileEditPage）にスタイルを適用し、2つのカードフォーム（表示名変更・パスワード変更）を整然としたデザインにする。パスワードフィールドへの表示/非表示トグル追加と、パスワード変更セクションへの警告メッセージ追加を含む。

---

## 背景・目的

プロフィール編集画面は2つのカード（`.card`）で構成されており、TICKET-022 のデザインシステムを適用することで即座に見栄えが整う。追加でパスワードフィールドの表示/非表示トグル（LoginPage と共通の PasswordInput コンポーネントを活用）と、パスワード変更後に再ログインが必要なことを事前に伝える警告メッセージを追加する。

---

## 受け入れ条件

- [ ] ページタイトル「プロフィール編集」が `.page-title` スタイルで表示される
- [ ] 「表示名の変更」カードと「パスワードの変更」カードがそれぞれ `.card` スタイルで表示される（border・shadow・border-radius）
- [ ] 各カードの見出し（H2）が `.card-title` スタイルで表示される（下線区切り含む）
- [ ] 表示名フォームの成功メッセージが `.alert-success` スタイルで表示される
- [ ] フォームのエラーメッセージが `.alert-error` スタイルで表示される
- [ ] パスワード変更セクションに「パスワードを変更すると自動的にログアウトされます」という情報バナー（`.alert-info`）が表示される
- [ ] パスワードフィールド（現在パスワード・新パスワード・確認パスワード）に表示/非表示トグルが追加されている（PasswordInput コンポーネントを利用）
- [ ] 各フォームの「保存」「変更」ボタンが `.btn-primary` スタイルで表示される
- [ ] ローディング中はボタンが disabled になり「保存中...」等のテキストが表示される（現状実装済み、スタイルのみ）
- [ ] max-width: 600px でコンテンツ幅が制限されている
- [ ] デスクトップ・モバイル両方で正しくレイアウトされる

---

## サブチケット（コミット単位）

- [ ] `feat(profile): 表示名変更カードのスタイルを適用`
- [ ] `feat(profile): パスワード変更カードに警告メッセージとPasswordInputトグルを追加`

---

## 関連情報

### パスワード変更セクションへの情報バナー追加

```tsx
// ProfileEditPage.tsx の「パスワードの変更」カード内に追加
<div className="alert alert-info" role="note">
  {t('profile.password.logoutWarning')}
  {/* 例: 「パスワードを変更すると、セキュリティのため自動的にログアウトされます。」 */}
</div>
```

ja.json に追加:
```json
{
  "profile": {
    "password": {
      "logoutWarning": "パスワードを変更すると、セキュリティのため自動的にログアウトされます。"
    }
  }
}
```

### PasswordInput コンポーネントの活用

TICKET-023 で作成した `PasswordInput` コンポーネントをプロフィール編集の3つのパスワードフィールドに適用する。

```tsx
// ProfileEditPage.tsx 変更前
<input
  id="current-password"
  type="password"
  value={currentPassword}
  onChange={(e) => setCurrentPassword(e.target.value)}
  disabled={isPasswordSubmitting}
  autoComplete="current-password"
/>

// 変更後
<PasswordInput
  id="current-password"
  label={t('profile.password.currentLabel')}
  value={currentPassword}
  onChange={(value) => setCurrentPassword(value)}
  disabled={isPasswordSubmitting}
  autoComplete="current-password"
/>
```

### スタイル仕様

```css
/* プロフィール編集ページ */
.profile-edit-page {
  max-width: 600px;
}

/* .card と .card-title は TICKET-022 の global.css で定義済み */
/* .alert-* は TICKET-022 の global.css で定義済み */
/* .btn-primary は TICKET-022 の global.css で定義済み */
/* .form-group, .form-group label, .form-group input は TICKET-022 で定義済み */
```

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 関連チケット: TICKET-023（PasswordInput コンポーネント）が作成済みであること
- 関連チケット: TICKET-021（パスワード変更・プロフィール編集UI）で実装された ProfileEditPage の既存コードに対してスタイルを追加する
