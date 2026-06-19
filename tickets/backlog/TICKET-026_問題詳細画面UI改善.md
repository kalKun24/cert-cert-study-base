# TICKET-026 問題詳細画面UI改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-026 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-026` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

問題詳細画面（QuestionDetailPage）にスタイルを適用し、コンテンツセクションの視覚的区別・解答の折りたたみ表示・削除確認 Modal（window.confirm の置き換え）・コメントセクションのデザイン改善を実装する。

---

## 背景・目的

問題詳細画面は問題文・解答・解説・議論点メモの4セクションが並んでいるが、スタイルがなく区別がつかない。また解答が問題文の直下に表示されるため、ユーザーが解答を考える前に目に入るリスクがある。学習効果向上のため解答セクションをデフォルトで折りたたみ、「解答を見る」ボタンで展開するパターンを採用する。

さらに `window.confirm` による削除確認はモバイルでの操作体験が悪く、スタイル統一もできないため、インライン確認 UI または Modal に置き換える。

---

## 受け入れ条件

- [ ] 問題タイトルがページタイトルとして視覚的に大きく（h1: 24px 以上）表示される
- [ ] メタ情報（作成日・ステータスバッジ・タグバッジ）がタイトル下に横並びで表示される
  - ステータスバッジは `data-status` 属性で draft(グレー)/private(オレンジ)/published(グリーン) を区別
- [ ] 「問題文」「解答」「解説」「議論点・メモ」の各セクションに見出し（h2）が表示され、下線区切りで視覚的に区別される
- [ ] 解答セクションはデフォルトで折りたたまれ、「解答を見る」ボタンで展開できる
  - 折りたたみは `aria-expanded` / `aria-controls` を使用
  - Escape キーで閉じる必要はないが、再クリックで閉じることができる
- [ ] Markdown コンテンツ（`.markdown-content`）にスタイルが適用される（コード・リスト・引用符 含む）
- [ ] 削除確認に `window.confirm` を使わず、削除ボタンクリック後に「本当に削除しますか？確認する / キャンセル」の二段階確認 UI を表示する
  - 実装は ConfirmModal コンポーネント（新規作成）または削除ボタンのインライン展開方式のどちらでも可
  - `role="dialog"` / `aria-modal="true"` のアクセシビリティ属性を付与すること
- [ ] コメントセクション（CommentSection）にスタイルが適用され、コメントカードとして表示される
  - 自分のコメントは左ボーダー（primary 色）で区別
  - コメント投稿フォームのプレビュートグル（書く / プレビュー）がタブデザインになっている
- [ ] 戻るリンクが「問題一覧に戻る」テキストでスタイルされている
- [ ] デスクトップ・モバイル両方で正しくレイアウトされる
- [ ] max-width: 860px でコンテンツ幅が制限されている

---

## サブチケット（コミット単位）

- [ ] `feat(question-detail): ConfirmModalコンポーネントを作成`
- [ ] `feat(question-detail): 問題詳細ヘッダー・メタ情報のスタイルを適用`
- [ ] `feat(question-detail): 解答セクションの折りたたみUI（CollapsibleSection）を実装`
- [ ] `feat(question-detail): Markdownコンテンツのスタイルを適用`
- [ ] `feat(question-detail): window.confirmをConfirmModalに置き換え`
- [ ] `feat(question-detail): コメントセクションのスタイルを適用`

---

## 関連情報

### ConfirmModal コンポーネントの設計

```tsx
// frontend/src/components/ConfirmModal.tsx (新規)
interface ConfirmModalProps {
  isOpen: boolean;
  title: string;
  message: string;
  confirmLabel: string;
  cancelLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
  isDangerous?: boolean;
}

// 使用例 (QuestionDetailPage.tsx)
<ConfirmModal
  isOpen={isDeleteConfirmOpen}
  title="問題を削除"
  message="この問題を削除すると元に戻せません。本当に削除しますか？"
  confirmLabel="削除する"
  cancelLabel="キャンセル"
  onConfirm={handleDelete}
  onCancel={() => setIsDeleteConfirmOpen(false)}
  isDangerous
/>
```

### CollapsibleSection コンポーネントの設計

```tsx
// frontend/src/components/CollapsibleSection.tsx (新規)
interface CollapsibleSectionProps {
  title: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
  revealLabel?: string;  // 「解答を見る」のような折りたたみ専用ラベル
  headingId: string;
}
```

### ステータスバッジの data 属性付与

```tsx
// QuestionDetailPage.tsx の既存コードを変更
// 変更前:
<span className="question-status">{t(`question.status.${question.status}`)}</span>

// 変更後:
<span className="question-status" data-status={question.status}>
  {t(`question.status.${question.status}`)}
</span>
```

### Markdown コンテンツのスタイル（主要仕様）

```css
.markdown-content { font-size: 0.9375rem; line-height: 1.75; color: #1e2433; }
.markdown-content code { font-family: monospace; padding: 2px 6px;
  background: #f0f2f5; border: 1px solid #e1e5ec; border-radius: 4px; }
.markdown-content pre { background: #1e2433; color: #e2e8f0;
  padding: 16px; border-radius: 6px; overflow-x: auto; }
.markdown-content blockquote { border-left: 3px solid #8daacc;
  padding: 8px 16px; background: #eef2f7; border-radius: 0 4px 4px 0; }
```

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 備考: `window.confirm` は QuestionDetailPage のみを対象とする。TagManagePage・TeamDetailPage・CommentSection の `window.confirm` 置き換えは各画面のチケット（TICKET-028/030）で対応する
