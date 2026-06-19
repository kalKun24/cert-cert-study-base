# TICKET-030 チーム詳細・作成・編集画面UI改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-030 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-030` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

チーム詳細画面（TeamDetailPage）・チーム作成画面（TeamCreatePage）・チーム編集画面（TeamEditPage）にスタイルを適用し、メンバーテーブルのデザイン・削除確認 Modal（window.confirm 置き換え）・メンバー招待フォームの UX を改善する。

---

## 背景・目的

チーム詳細画面はチーム名・説明文・メンバー一覧（テーブル）・メンバー招待フォームを含む最も複雑な画面の一つ。現状はスタイルがなく、テーブルが裸の HTML テーブル状態。削除確認に `window.confirm` を使用しており置き換えが必要。またメンバーテーブルに表示名ではなく userId（UUID）が表示されており視認性が低い。

---

## 受け入れ条件

### チーム詳細 (TeamDetailPage)
- [ ] チームヘッダー（名前・説明・アクションボタン）が整然と表示される
  - チーム名: h1（page-title スタイル）
  - 説明文: セカンダリテキスト
  - 「編集」「削除」ボタン: 右端に横並び（owner/admin のみ）
- [ ] メンバー一覧テーブルにスタイルが適用される
  - テーブルヘッダー（薄いグレー背景）
  - 行のボーダー区切り
  - hover 時の行の背景色変化
  - 操作列（除外ボタン）は右端に固定
- [ ] メンバーテーブルはモバイルで横スクロール可能なラッパー（`.user-table-wrapper` を流用）で包まれている
- [ ] 削除確認に `window.confirm` を使わず ConfirmModal を使用する（TICKET-026 で作成済みのコンポーネントを利用）
- [ ] MemberInvite フォームがメンバーセクション下部にカードで表示される
  - 招待成功時にインライン成功メッセージが表示される
  - 招待エラー時にインラインエラーメッセージが表示される（現状確認が必要）
- [ ] 戻るリンク（「チーム一覧に戻る」）がスタイルされている

### チーム作成・編集 (TeamCreatePage / TeamEditPage)
- [ ] TeamForm コンポーネント（チーム名・説明フォーム）にスタイルが適用されている
  - フォームがカード（card クラス）で包まれている
  - 「保存」「キャンセル」ボタンが横並び
- [ ] 未保存変更がある状態でキャンセルした場合の確認ダイアログが表示される（シンプルな `window.confirm` または ConfirmModal）
- [ ] ローディング・エラー状態がスタイルされている

---

## サブチケット（コミット単位）

- [ ] `feat(team-detail): チームヘッダーと説明のスタイルを適用`
- [ ] `feat(team-detail): メンバーテーブルのスタイルを適用（レスポンシブ対応含む）`
- [ ] `fix(team-detail): window.confirmをConfirmModalに置き換え`
- [ ] `feat(team-detail): MemberInviteフォームのスタイルと成功フィードバックを改善`
- [ ] `feat(team-form): TeamFormコンポーネントのスタイルを適用`

---

## 関連情報

### メンバーテーブルのスタイル（TICKET-022 の .user-table を流用）

```css
/* チームメンバーテーブルは user-table クラスのスタイルを流用 */
.team-members-table {
  /* user-table と同じスタイルを適用 */
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
}
```

または `team-members-table` クラスに `user-table` と同じ CSS を当てる。

### MemberInvite の成功フィードバック改善

現状の MemberInvite コンポーネントは招待成功時に `onInvited` コールバックを呼ぶのみで、ユーザーへのフィードバックがない可能性がある。コンポーネント内部に成功メッセージ用の state を追加する。

```tsx
// MemberInvite.tsx に追加
const [inviteSuccess, setInviteSuccess] = useState('');

// 成功時
setInviteSuccess('メンバーを招待しました');
onInvited();

// JSX
{inviteSuccess && (
  <div className="alert alert-success" role="status">{inviteSuccess}</div>
)}
```

### TeamDetailPage の削除確認置き換え

```tsx
// 変更前
const confirmed = window.confirm(t('team.detail.deleteConfirm', { name: team.name }));

// 変更後: ConfirmModal を使用
const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);

<ConfirmModal
  isOpen={isDeleteConfirmOpen}
  title="チームを削除"
  message={t('team.detail.deleteConfirm', { name: team.name })}
  confirmLabel="削除する"
  cancelLabel="キャンセル"
  onConfirm={handleDelete}
  onCancel={() => setIsDeleteConfirmOpen(false)}
  isDangerous
/>
```

### チーム作成・編集フォームのスタイル

```css
.team-form-page { max-width: 600px; }

/* TeamForm コンポーネントに card クラスを適用 */
/* または team-form-page 配下に card スタイルを当てる */
```

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 関連チケット: TICKET-026（ConfirmModal コンポーネント）が作成済みであること
- 関連チケット: TICKET-029（チーム一覧）と合わせてチーム関連 UI を統一する
