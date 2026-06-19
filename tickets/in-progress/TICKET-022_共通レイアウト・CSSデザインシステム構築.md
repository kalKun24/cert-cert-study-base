# TICKET-022 共通レイアウト・CSSデザインシステム構築

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-022 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-19 |
| 着手日 | 2026-06-19 |
| 完了日 | - |
| ブランチ名 | `feature/profile-edit` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

全画面共通の CSS デザインシステム（デザイントークン・グローバルスタイル）を構築し、NavBar とサイドバーのレイアウト・ナビゲーション構造を改善する。現状はすべての画面に class 名が付与されているがスタイルが存在しない状態であるため、このチケットで視覚的な基盤を整える。

---

## 背景・目的

現在 `frontend/src/` 配下のコンポーネントには `.btn`, `.btn-primary`, `.navbar`, `.sidebar-menu` 等の class 名が付与されているが、対応するCSSファイルが存在しない。すべての画面のUIを整備する前提として、まずデザイントークンとグローバルスタイルを一元管理するCSSファイルを作成する必要がある。

また NavBar とサイドバーにナビゲーションリンクが重複しており（ホーム・問題一覧・タグ一覧が両方に存在）、さらにサイドバーに `/teams` リンクが欠落している。構造を整理してナビゲーションを一本化する。

---

## 受け入れ条件

- [x] `frontend/src/styles/global.css` が作成され、`main.tsx` で import されている
- [x] CSS カスタムプロパティ（デザイントークン）が `:root` に定義されている
  - カラーシステム: Happy Hues Palette 10 ベース（ダークティール + ゴールドアクセント）
  - タイポグラフィ: フォントファミリー・フォントサイズ・ウェイト・行高
  - スペーシング: 4pxグリッドベースの `--space-*` 変数
  - ボーダー・シャドウ・トランジション・フォーカスリング
- [x] グローバルリセットスタイル（`box-sizing: border-box`・body のデフォルトスタイル）が適用されている
- [x] アプリレイアウト（`.app-layout` / `.app-body` / `.main-content`）がデスクトップ・タブレット・モバイルで正しく表示される
  - デスクトップ: NavBar(sticky, 56px) + Sidebar(240px) + Main の3ペイン
  - モバイル: サイドバー非表示、NavBar にハンバーガーボタンを追加してサイドバーをオーバーレイ表示
- [x] NavBar はブランド名・ユーザー情報（表示名・ロールバッジ）・プロフィール/ログアウトボタンのみに整理される（ナビリンクはサイドバーに集約）
- [x] サイドバーに `/teams` リンクが追加され、ナビゲーションリンクが一元管理されている（ホーム・問題一覧・タグ管理・チーム一覧・ユーザー管理）
- [x] サイドバーのリンクに `NavLink` の `active` クラスが機能しアクティブ状態が視覚的に区別できる
- [x] `role-badge` に `data-role` 属性を付与し、admin は赤系・その他は青系でスタイルが異なる
- [x] モバイル時に NavBar のハンバーガーボタン（☰）をタップするとサイドバーがオーバーレイ表示され、外側タップ・ESC キーで閉じる
- [x] WCAG AA 準拠: ボタン色(#f9bc60)と背景(#004643)のコントラスト比 4.5:1 以上
- [x] `@media (prefers-reduced-motion: reduce)` によるモーション抑制が適用されている
- [x] `.sr-only` クラスが定義されている

---

## サブチケット（コミット単位）

- [x] `feat(styles): CSSデザイントークンとグローバルスタイルを追加`
- [x] `feat(layout): アプリレイアウト・NavBar・サイドバーのスタイルを追加`
- [x] `refactor(layout): NavBarのナビリンクを削除しサイドバーに集約、/teams リンクを追加`
- [x] `feat(layout): モバイル用ハンバーガーメニューとサイドバーオーバーレイを実装`
- [x] `feat(styles): ボタン・フォーム・アラート・カードのグローバルスタイルを追加`
- [x] `feat(styles): バッジ・ステータス・ページタイトルのスタイルを追加`

---

## 関連情報

### CSSファイル構成

```
frontend/src/styles/
└── global.css   # デザイントークン + 全コンポーネントのグローバルスタイル
```

`main.tsx` に `import './styles/global.css';` を追加するだけで全画面に反映される。

### カラーパレット（主要トークン）

Happy Hues Palette 10 ( https://www.happyhues.co/palettes/10 ) をベースに採用。

```css
/* 背景・サーフェス */
--color-bg-page:        #004643;  /* アプリ全体の背景: ダークティール */
--color-bg-card:        #0f3433;  /* カード・パネル: やや暗いティール */
--color-bg-sidebar:     #001e1d;  /* サイドバー: 最も暗いティール */

/* テキスト */
--color-text-primary:   #fffffe;  /* 見出し: ほぼ白 */
--color-text-body:      #abd1c6;  /* 本文: ライトミント */
--color-text-muted:     #abd1c6;  /* サブテキスト */

/* アクセント（ボタン・リンク） */
--color-accent:         #f9bc60;  /* ゴールド: メインアクション */
--color-accent-text:    #001e1d;  /* アクセント上のテキスト */

/* セマンティック */
--color-error:          #e16162;  /* エラー: コーラルレッド */
--color-success:        #abd1c6;  /* 成功: ミント */
--color-border-default: #0f3433;  /* ボーダー */
```

### レイアウト仕様

```
デスクトップ (1024px+):
  NavBar:  height: 56px, sticky top: 0, z-index: 100
  Sidebar: width: 240px, sticky top: 56px, height: calc(100vh - 56px)
  Main:    flex: 1, padding: 32px 40px

タブレット (768px - 1023px):
  Sidebar: width: 200px

モバイル (〜767px):
  Sidebar: display: none（ハンバーガータップ時にオーバーレイとして表示）
  NavBar:  ハンバーガーボタン（☰）を左端に追加
  Main:    padding: 16px
```

### NavBar 改善方針

現状: NavBar にナビリンク（ホーム・問題一覧・タグ一覧・チーム一覧）+ サイドバーに重複リンク
改善: NavBar はブランド名・ユーザーセクションのみ。ナビリンクはサイドバーに集約。

```tsx
// Layout.tsx: <Link> を NavLink に変更
// NavBar.tsx: .navbar-links セクション全体を削除
// Sidebar: /teams リンク追加
```

### ボタン・フォーム・アラートのデザイン仕様

- `btn-primary`: background #f9bc60, color #001e1d, hover: brightness 1.08
- `btn-secondary`: border #abd1c6, color #abd1c6, 背景透明
- `btn-danger`: background #e16162, color #fffffe
- `form-input`: background #0f3433, border #abd1c6, color #fffffe, focus: border #f9bc60 + フォーカスリング
- `alert-error`: 背景 rgba(#e16162, 0.15), テキスト #e16162, ボーダー #e16162
- `alert-success`: 背景 rgba(#abd1c6, 0.15), テキスト #abd1c6, ボーダー #abd1c6

### 関連チケット

- 関連チケット: このチケットが完了した後、TICKET-023〜032 の各画面 UI チケットを着手する
- 備考: Tailwind CSS は導入しない。既存の class 名を活かし CSS 変数で一元管理する方針。
