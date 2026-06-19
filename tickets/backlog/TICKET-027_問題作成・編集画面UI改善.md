# TICKET-027 問題作成・編集画面UI改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-027 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-027` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

問題作成・編集画面（QuestionCreatePage / QuestionEditPage）にスタイルを適用し、Markdown エディタのタブデザイン・タグ選択エリア・公開ステータス選択・フォームアクション（保存・キャンセル）の UX を改善する。また未保存の変更がある状態でページを離れようとした場合の警告を追加する。

---

## 背景・目的

問題作成・編集画面はエディタタブ切り替え方式で問題文・解答・解説・議論点メモを編集する。現状はスタイルがなく、タブの選択状態・エディタエリアの境界・タグ選択エリアの見た目がわからない。またキャンセルボタン押下時に入力内容を確認なく破棄してしまう問題がある。

---

## 受け入れ条件

- [ ] フォームの各セクション（タイトル・エディタ・タグ・ステータス・アクション）がカード内に整然と配置される
- [ ] エディタタブ（問題文・解答・解説・議論点メモ）がタブデザインになっている
  - アクティブタブは primary 色のテキスト・下ボーダーなし（エディタエリアと繋がる外観）
  - 非アクティブタブは hover 時に背景色が変化する
  - `aria-selected` / `role="tab"` が正しく機能する（現状実装済み、視覚的スタイルのみ追加）
- [ ] タグ選択エリア（`fieldset`）が bordered なグループとして表示される
  - タグが多い場合に最大高さ（max-height: 200px）でスクロール可能になる
- [ ] 公開ステータス select がカスタムスタイルの select として表示される（TICKET-022 の `.form-select` 適用）
- [ ] 「保存」「キャンセル」ボタンがフォーム下部に横並びで表示される
  - 保存: primary ボタン
  - キャンセル: secondary ボタン
- [ ] フォームに変更が加えられた状態（dirty 状態）でキャンセルボタンをクリックした場合、確認ダイアログを表示する
  - React Router の `useBlocker` またはシンプルな ConfirmModal を使用する
- [ ] バリデーションエラーメッセージが該当フィールドの下に赤テキストで表示される
- [ ] 送信中はフォーム全体がローディング状態（ボタン disabled + テキスト変更）になる
- [ ] QuestionCreatePage と QuestionEditPage のフォーム部分が共通コンポーネント（QuestionForm）に整理されていること（任意: 実装の重複が大きい場合）
- [ ] デスクトップ・モバイル両方で正しくレイアウトされる
- [ ] WCAG AA 準拠のフォーカスインジケーターが適用されている

---

## サブチケット（コミット単位）

- [ ] `feat(question-form): エディタタブのスタイルを適用`
- [ ] `feat(question-form): タイトル・タグ選択・ステータス選択のスタイルを適用`
- [ ] `feat(question-form): フォームアクション（保存・キャンセル）のスタイルと dirty ガードを追加`
- [ ] `refactor(question-form): QuestionCreatePage と QuestionEditPage の共通フォームを QuestionForm コンポーネントに整理 (任意)`

---

## 関連情報

### Dirty ガードの実装方針

```tsx
// QuestionCreatePage.tsx / QuestionEditPage.tsx に追加
const [isDirty, setIsDirty] = useState(false);

// フォーム変更時に isDirty を true にする
const handleFormChange = (newValues: Partial<FormValues>) => {
  setIsDirty(true);
  setForm(prev => prev ? { ...prev, ...newValues } : prev);
};

// キャンセルボタン押下時
const handleCancel = () => {
  if (isDirty) {
    if (!window.confirm('入力内容を破棄してキャンセルしますか？')) return;
    // ConfirmModal が TICKET-026 で作成済みの場合はそれを使用
  }
  navigate('/questions');
};
```

### タグ選択エリアのスタイル

```css
/* タグが多い場合にスクロール可能にする */
.tag-checkbox-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  max-height: 200px;
  overflow-y: auto;
  padding: var(--space-3);
  background-color: var(--color-bg-page);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-md);
}
```

### エディタタブのスタイル（TICKET-022 参照）

```css
.editor-tabs {
  display: flex;
  border-bottom: 1px solid var(--color-border-default);
}
.editor-tab {
  height: 36px;
  padding: 0 var(--space-4);
  border: 1px solid transparent;
  border-bottom: none;
  border-radius: var(--radius-md) var(--radius-md) 0 0;
  position: relative;
  bottom: -1px;
}
.editor-tab--active {
  color: var(--color-primary-600);
  background-color: var(--color-bg-card);
  border-color: var(--color-border-default);
  border-bottom-color: var(--color-bg-card);
}
```

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 関連チケット: TICKET-026（ConfirmModal コンポーネント）が作成済みであれば dirty ガードで使用可能
