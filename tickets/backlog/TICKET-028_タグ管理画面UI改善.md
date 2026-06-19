# TICKET-028 タグ管理画面UI改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-028 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-028` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

タグ管理画面（TagManagePage）にスタイルを適用し、タグ一覧のリストデザイン・インライン編集UI・クライアントサイド検索フィルター・`window.alert` / `window.confirm` の置き換えを実装する。

---

## 背景・目的

タグ管理画面は admin ユーザーがタグの作成・編集・削除を行う管理画面。現状はスタイルがなく、タグ一覧が単純なリストとして表示されており視認性が低い。また削除確認と削除エラー通知に `window.confirm` / `window.alert` を使用しており、モバイルでの操作体験が悪い。タグ数が増えると一覧が長くなるため、クライアントサイドの検索フィルターも追加する。

---

## 受け入れ条件

- [ ] タグ作成フォームがカード（border・border-radius）で包まれて表示される（admin のみ）
- [ ] タグ作成フォームの input と「追加」ボタンが横並び（`flex` レイアウト）で表示される
- [ ] タグ一覧の各行がカードアイテム（border・padding）として表示される
  - タグ名（セミボールド）+ 右端に「編集」「削除」ボタン（admin のみ）
  - hover 時にボーダー色が強調される
- [ ] インライン編集時（`editingId === tag.id`）のフォームが自然にタグ行内に収まる
  - input + 「保存」「キャンセル」ボタンが横並びになる
  - 編集開始時に input に `autoFocus` が効いている（現状実装済み）
- [ ] クライアントサイドのタグ検索 input が一覧上部に追加され、タグ名でリアルタイムフィルタリングできる
- [ ] 削除確認に `window.confirm` を使わず、ConfirmModal コンポーネント（TICKET-026 で作成）を使用する
- [ ] 削除エラー通知に `window.alert` を使わず、インラインのエラーアラート（`.alert-error`）で表示する
- [ ] タグが 0 件のときの空状態メッセージが適切にスタイルされている
- [ ] デスクトップ・モバイル両方で正しくレイアウトされる（max-width: 700px）

---

## サブチケット（コミット単位）

- [ ] `feat(tag-manage): タグ作成フォームのスタイルを適用`
- [ ] `feat(tag-manage): タグ一覧のカードリストスタイルを適用`
- [ ] `feat(tag-manage): クライアントサイド検索フィルターを追加`
- [ ] `fix(tag-manage): window.confirm/alertをConfirmModalとインラインエラーに置き換え`

---

## 関連情報

### クライアントサイド検索フィルターの実装

```tsx
// TagManagePage.tsx に追加
const [searchQuery, setSearchQuery] = useState('');

const filteredTags = useMemo(
  () => tags.filter(tag => tag.name.toLowerCase().includes(searchQuery.toLowerCase())),
  [tags, searchQuery]
);

// JSX に追加（タグ一覧の上部）
<div className="tag-search-wrapper">
  <input
    type="search"
    className="form-input"
    placeholder="タグを検索..."
    value={searchQuery}
    onChange={(e) => setSearchQuery(e.target.value)}
    aria-label="タグを検索"
  />
</div>

// ul 内の tags.map を filteredTags.map に変更
```

### 削除エラーのインライン表示

```tsx
// TagManagePage.tsx に追加
const [deleteError, setDeleteError] = useState('');

const handleDelete = async (tag: Tag) => {
  // 確認は ConfirmModal で
  setDeleteError('');
  try {
    await deleteTag(tag.id);
    setTags(prev => prev.filter(t => t.id !== tag.id));
  } catch {
    setDeleteError(t('tag.error.deleteFailed'));
  }
};

// JSX に追加
{deleteError && (
  <div className="alert alert-error" role="alert">{deleteError}</div>
)}
```

### スタイル仕様

```css
.tag-manage-page { max-width: 700px; }

.tag-create-form-wrapper {
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  padding: var(--space-5);
  margin-bottom: var(--space-6);
}

.tag-manage-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.tag-manage-item {
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4);
  transition: border-color var(--transition-fast);
}
.tag-manage-item:hover { border-color: var(--color-border-strong); }

.tag-manage-item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.tag-manage-name {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-medium);
}

.tag-manage-actions { display: flex; gap: var(--space-2); }

.tag-create-form-row,
.tag-edit-form-row {
  display: flex;
  gap: var(--space-3);
  align-items: flex-start;
}
.tag-create-form-row .form-input,
.tag-edit-form-row .form-input { flex: 1; }
```

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 関連チケット: TICKET-026（ConfirmModal コンポーネント）が作成済みであること
