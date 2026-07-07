# TICKET-078 Markdownプレビューの見出しCSSカスタマイズ

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-078 |
| ステータス | 🟢 完了 |
| 作成日 | 2026-06-23 |
| 着手日 | 2026-06-23 |
| 完了日 | 2026-06-23 |
| ブランチ名 | `feature/markdown-preview-heading-css` |
| PR番号 | #77 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/77 |

---

## 概要

Markdownプレビュー内の見出し（h1〜h3）にGitHubスタイルのCSSを適用し、文書の読みやすさを向上させる。

---

## 背景・目的

現状のプレビューは見出しにスタイルが当たっておらず、長文の問題・解説を読む際に階層構造が把握しにくい。H2に下線を引くなどのスタイルを追加し、セクション区切りを視覚的に明確にする。

---

## 受け入れ条件

- [x] Markdownプレビューの h1 にボトムボーダー（2px・グレー）が表示される
- [x] Markdownプレビューの h2 にボトムボーダー（1px・グレー）が表示される
- [x] h1・h2・h3 のフォントサイズ・ウェイトが適切に設定されている
- [x] 既存のエディタ機能・他ページの表示に影響がない
- [x] ダークモード（使用中の場合）でも視認性が保たれている

---

## サブチケット（コミット単位）

- [x] `style(markdown): プレビュー見出し（h1〜h3）にGitHubスタイルのCSSを適用`

---

## 関連情報

- 関連チケット: TICKET-077（Markdownエディタのリッチ化）
- 適用CSS（GitHubスタイル・案A）:

```css
.markdown-preview h1 {
  font-size: 1.875rem;
  font-weight: 700;
  padding-bottom: 0.4em;
  border-bottom: 2px solid #e5e7eb;
  margin-bottom: 1rem;
}
.markdown-preview h2 {
  font-size: 1.5rem;
  font-weight: 600;
  padding-bottom: 0.3em;
  border-bottom: 1px solid #e5e7eb;
  margin-bottom: 0.875rem;
}
.markdown-preview h3 {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
}
```

- 備考: `.markdown-preview` のクラス名は実際のコンポーネントに合わせて調整する
