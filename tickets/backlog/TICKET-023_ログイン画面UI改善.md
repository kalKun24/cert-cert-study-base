# TICKET-023 ログイン画面UI改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-023 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-023` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

ログイン画面（LoginPage）に視覚的なスタイルを適用し、パスワード表示/非表示トグル・ローディングスピナー・エラー時フォーカス移動を追加することで、UX を改善する。あわせてパスワード変更成功後のリダイレクト先（LoginPage）でロケーションステートの `message` を表示する対応も含む。

---

## 背景・目的

現状のログイン画面は `.login-container` / `.login-card` の class 名が付与されているがスタイルが存在せず、視覚的なフィードバックが最低限となっている。TICKET-022 のデザインシステム構築後に、ログイン画面固有の改善を行う。

また ProfileEditPage でパスワード変更成功後に `navigate('/login', { state: { message: ... } })` でメッセージを渡しているが、LoginPage 側では受け取っていないため、変更成功の旨が表示されない問題がある。

---

## 受け入れ条件

- [ ] `.login-container` が viewport 縦中央配置（flexbox）になっている
- [ ] `.login-card` が max-width: 400px のカードデザインで表示される（shadow・border-radius 含む）
- [ ] ログインボタンはカード幅いっぱい（width: 100%）
- [ ] パスワード入力フィールドに表示/非表示トグルボタンが追加されている
  - トグルボタンは `type="button"` で `aria-label="パスワードを表示"` / `"パスワードを隠す"` を動的に切り替える
- [ ] ローディング中にボタン内でローディング表示（テキスト切り替えまたはスピナー）が行われる
- [ ] API エラー発生時にエラーバナーへフォーカスが移動する（`errorRef.current?.focus()`）
- [ ] ProfileEditPage からパスワード変更成功後にリダイレクトされた場合、ロケーションステートの `message` が LoginPage に成功メッセージとして表示される
- [ ] デスクトップ・モバイル両方で正しくレイアウトされる（モバイルはカードが margin: 16px で全幅に近い形）
- [ ] WCAG AA 準拠のコントラスト比・フォーカスインジケーターが適用されている
- [ ] `autoComplete="username"` / `autoComplete="current-password"` は維持する

---

## サブチケット（コミット単位）

- [ ] `feat(login): ログイン画面にスタイルを適用（カード・ボタン・フォーム）`
- [ ] `feat(login): パスワード表示/非表示トグルを追加`
- [ ] `fix(login): APIエラー時にエラーバナーへフォーカスを移動する`
- [ ] `fix(login): パスワード変更成功後のリダイレクト時にsuccessメッセージを表示する`

---

## 関連情報

### パスワードトグルの実装方針

```tsx
// LoginPage.tsx 内部状態に追加
const [showPassword, setShowPassword] = useState(false);

// input type を動的に切り替え
<input
  type={showPassword ? 'text' : 'password'}
  // ...
/>
<button
  type="button"
  aria-label={showPassword ? 'パスワードを隠す' : 'パスワードを表示'}
  onClick={() => setShowPassword(prev => !prev)}
>
  {/* アイコン */}
</button>
```

### ロケーションステート受け取りの実装方針

```tsx
// LoginPage.tsx に追加
const location = useLocation(); // 既に import 済み
const successMessage = (location.state as { message?: string })?.message ?? '';

// JSX に追加（error バナーの上に配置）
{successMessage && (
  <div className="alert alert-success" role="status">
    {successMessage}
  </div>
)}
```

### ビジュアルデザイン仕様

```css
/* TICKET-022 の global.css に含まれるため追記不要 */
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--color-bg-page);
  padding: var(--space-4);
}
.login-card {
  width: 100%;
  max-width: 400px;
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-xl);    /* 12px */
  box-shadow: var(--shadow-lg);
  padding: var(--space-10) var(--space-8);
}
```

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 関連チケット: TICKET-021（パスワード変更UI）で実装された ProfileEditPage のリダイレクト処理に対応する
