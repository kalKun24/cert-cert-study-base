# TICKET-025 問題一覧画面UI改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-025 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-025` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

問題一覧画面（QuestionListPage）にスタイルを適用し、問題カードのデザイン・フィルターパネルのレイアウト・空状態メッセージ・キーワード検索のデバウンスを改善する。

---

## 背景・目的

現状の問題一覧は class 名が付与されているがスタイルがなく、リスト形式・カード形式のどちらとも判断できない状態。また onChange ごとにキーワード検索 API が呼ばれるパフォーマンス問題、検索結果ゼロ時にフィルターをリセットする手段がない UX 問題が存在する。

---

## 受け入れ条件

- [ ] 問題一覧の各行がカードデザイン（border・shadow・hover エフェクト）で表示される
  - タイトル（セミボールド）・作成日（小さめテキスト）・タグバッジが整然と並んでいる
  - hover 時に border-color が primary 色に変化し、shadow が強調され translateY(-1px) される
- [ ] フィルターパネルが白いカード（border + border-radius）で包まれ、キーワード input とタグチップが縦に並ぶ
- [ ] キーワード検索に 300ms のデバウンスが適用され、入力のたびに API 呼び出しが起きない
- [ ] 検索条件があり結果が 0 件の場合、「フィルターをリセット」ボタン付きの空状態メッセージを表示する
- [ ] タグチップが未選択・選択状態で視覚的に区別できる（選択中は primary 色で塗りつぶし）
- [ ] 問題カードのタイトルが 2 行を超える場合は `-webkit-line-clamp: 2` で切り捨てられる
- [ ] ページネーションが「前ページ」「次ページ」のボタンとページ情報（例: 2 / 5）で表示される
- [ ] デスクトップ・モバイル両方で正しくレイアウトされる
- [ ] ローディング状態・エラー状態が既存の表示から改善される（スタイルが適用される）
- [ ] WCAG AA 準拠: タグチップが `role="checkbox"` を保持し、フィルターグループに `aria-label` がある

---

## サブチケット（コミット単位）

- [ ] `feat(question-list): キーワード検索のデバウンス (useDebounce hook) を追加`
- [ ] `feat(question-list): 問題カードのスタイルを適用`
- [ ] `feat(question-list): フィルターパネルのスタイルを適用`
- [ ] `feat(question-list): 空状態メッセージとフィルターリセット機能を追加`
- [ ] `feat(question-list): ページネーションのスタイルを適用`

---

## 関連情報

### useDebounce hook の実装

```ts
// frontend/src/utils/useDebounce.ts (新規)
import { useState, useEffect } from 'react';

export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);
  return debouncedValue;
}
```

```tsx
// QuestionListPage.tsx での使用
const debouncedKeyword = useDebounce(keyword, 300);

// useEffect の依存配列を keyword → debouncedKeyword に変更
useEffect(() => {
  syncSearchParams(page, debouncedKeyword, selectedTagNames);
  const cleanup = loadQuestions(page, debouncedKeyword, selectedTagNames);
  return cleanup;
}, [page, debouncedKeyword, selectedTagNames, loadQuestions, syncSearchParams]);
```

### フィルターリセット

```tsx
// QuestionListPage.tsx に追加
const hasActiveFilters = keyword.trim() !== '' || selectedTagNames.length > 0;

const handleReset = () => {
  setKeyword('');
  setSelectedTagNames([]);
  setPage(1);
};

// 空状態メッセージ
{questions.length === 0 && hasActiveFilters && (
  <div className="question-list-empty-filtered">
    <p>検索条件に一致する問題が見つかりませんでした。</p>
    <button type="button" className="btn btn-secondary" onClick={handleReset}>
      フィルターをリセット
    </button>
  </div>
)}
```

### タグチップのスタイル（TICKET-022 参照）

選択状態の class 切り替えは TagChip コンポーネントの `selected` prop で制御済み。
`selected` が true のとき `.tag-chip--selected` class を付与し、primary 色で塗りつぶす。

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 関連チケット: TICKET-019（ページネーションUI）で実装済みの Paginator コンポーネントにスタイルを適用
