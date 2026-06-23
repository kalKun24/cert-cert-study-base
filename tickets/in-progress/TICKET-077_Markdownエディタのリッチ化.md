# TICKET-077 Markdownエディタのリッチ化

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-077 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-23 |
| 着手日 | 2026-06-23 |
| 完了日 | - |
| ブランチ名 | feature/TICKET-077 |
| PR番号 | #79 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/79 |

---

## 概要

現在の `@uiw/react-md-editor` ベースのエディタを、Obsidian / VSCode に近いリッチな Markdown エディタへ置き換える。改行挙動の不具合・見た目の粗さ・機能不足を解消し、快適な編集体験を提供する。

---

## 背景・目的

問題・解答・解説・ノート本文など、テキスト入力の比重が高いアプリであるにもかかわらず、現在のエディタには以下の問題がある。

### 現状の問題点

#### 機能面
- **改行の挙動が不自然**: `@uiw/react-md-editor` は `textarea` + `<pre>` のシンタックスハイライトを重ね合わせる方式で、Enter キーによる改行が意図しない挙動を起こすことがある（Markdown の 2改行ルールと UI の齟齬）。
- **シンタックスハイライトが貧弱**: 見出し・コードブロック・リストなどのハイライトが視覚的に分かりにくく、長文の編集で構造が把握しにくい。
- **ツールバーの機能が最低限**: 太字・斜体・リンク程度しかなく、テーブル挿入・コードブロック挿入・チェックリストなどのショートカットがない。
- **コードブロックの言語指定 UI がない**: `` ```python `` のような言語タグ入力は手入力のみ。
- **Vim / Emacs キーバインドに非対応**: 開発者・資格勉強ユーザーがよく使うキーバインドが使えない。
- **ライブプレビューと編集エリアの境界が分かりにくい**: `preview="live"` で左右分割しているが、分割比の調整ができない。

#### 見た目・UX 面
- **エディタ全体のデザインがプロダクトの Teal テーマと不整合**: CSS オーバーライドで補正しているが、ツールバーアイコンなどに違和感が残る。
- **モバイルでプレビューが省略される**: モバイル時は `w-md-editor` の高さを固定している（global.css L2802）ため、プレビューが実質見えない。
- **フォントサイズ・行間**: エディタ内のテキストが読みにくい。

#### コード面
- `QuestionCreatePage` / `QuestionEditPage` / `NoteCreatePage` / `NoteEditPage` の 4 ファイルに `MDEditor` の呼び出しが重複しており、エディタコンポーネントが共通化されていない。
- `height="100%"` + CSS での高さ制御が複雑で、レイアウト変更時の影響範囲が大きい。

---

## 改善後のゴール

- **Enter キーで普通に改行できる**（Obsidian 相当の自然な改行体験）
- **シンタックスハイライト**: 見出し・太字・コードブロック・リストが色分けされ、Markdown の構造が一目でわかる
- **豊富なツールバー**: テーブル・コードブロック・チェックリスト・水平線の挿入ボタンを備える
- **スプリットビュー**: 編集ペインとプレビューペインを分割比調整可能にする
- **テーマ整合**: Teal カラーテーマに合ったデザイン
- **共通エディタコンポーネント化**: `MarkdownEditor.tsx` として切り出し、全エディタページで再利用

---

## 採用候補ライブラリ比較

| ライブラリ | バンドルサイズ | シンタックスHL | ツールバー | Vim対応 | プレビュー | メンテナンス状況 | 備考 |
|---|---|---|---|---|---|---|---|
| **@uiw/react-md-editor**（現行） | ~300 KB | 最低限 | 基本的 | なし | 左右分割 | 活発 | 改行不具合あり |
| **@uiw/react-codemirror** + **CodeMirror 6** | ~200 KB（必要拡張のみ） | 高品質（lang-markdown） | 要カスタム | cm-vim 拡張あり | 別途 react-markdown | 非常に活発 | Obsidian, Zenn のベース技術。柔軟性最高 |
| **Monaco Editor**（@monaco-editor/react） | ~5 MB（遅延ロードで軽減可） | VSCode 同等 | VSCode 同等 | なし（extあり） | 別途実装 | Microsoft 公式 | バンドル大きい・VSCode 体験を求めるなら最適 |
| **Milkdown** | ~150 KB | WYSIWYG | ProseMirror ベース | プラグインあり | リアルタイム | 活発 | WYSIWYG方式。Markdown との二重管理が必要 |
| **react-md-editor**（uiwと別物、npm: react-md-editor） | ~80 KB | 限定的 | 基本的 | なし | プレビュー分離 | やや低迷 | 現行と近い機能 |

### 推奨: CodeMirror 6 + @uiw/react-codemirror

- **理由**: バンドルサイズを抑えつつ、lang-markdown 拡張でシンタックスハイライトが高品質。改行・カーソル移動がネイティブに近い挙動。Vim キーバインド（@codemirror/legacy-modes または cm-vim）を後から拡張できる。プレビューは既存の `react-markdown` + `rehype-sanitize` を流用可能。
- **次点**: プロジェクトが VSCode 体験を最優先するなら Monaco Editor（ただし遅延ロード必須）。

---

## 受け入れ条件

- [ ] Enter キーで自然に改行でき、2回 Enter で段落区切りとなる Markdown の改行ルールが直感的に操作できること
- [ ] 見出し（#〜###）・太字・斜体・コードブロック・リストのシンタックスハイライトが有効なこと
- [ ] ツールバーに太字・斜体・リンク・テーブル・コードブロック・チェックリストの挿入ボタンがあること
- [ ] 編集エリアとプレビューを左右に分割表示できること（スプリットビュー）
- [ ] プレビューは既存の `react-markdown` + `rehype-sanitize` を使い、XSS 対策が維持されること
- [ ] Teal テーマとデザイン的に整合するスタイルであること（CSS カスタムプロパティ使用）
- [ ] `MarkdownEditor.tsx` として共通コンポーネントに切り出し、`QuestionCreatePage` / `QuestionEditPage` / `NoteCreatePage` / `NoteEditPage` の 4 ページすべてで利用されること
- [ ] モバイル表示でもエディタとプレビューが（タブ切り替えなどで）参照できること
- [ ] 既存の `@uiw/react-md-editor` への依存が削除されること（package.json から除去）
- [ ] ESLint / Prettier のエラーがないこと

---

## サブチケット（コミット単位）

- [x] `chore(deps): react-md-editor を削除し @uiw/react-codemirror と関連拡張を追加`
- [x] `feat(editor): MarkdownEditor 共通コンポーネントを実装（CodeMirror 6 + プレビュー分割）`
- [x] `feat(editor): MarkdownEditor にツールバー（テーブル・コードブロック・チェックリスト）を追加`
- [x] `refactor(question): QuestionCreatePage・QuestionEditPage を MarkdownEditor に切り替え`
- [x] `refactor(note): NoteCreatePage・NoteEditPage を MarkdownEditor に切り替え`
- [x] `style(editor): Teal テーマとの整合スタイルを適用し旧 MDEditor CSS オーバーライドを削除`
- [x] `test(editor): MarkdownEditor のユニットテストを追加`

---

## 関連情報

- 関連チケット: -
- 参考:
  - [@uiw/react-codemirror](https://github.com/uiwjs/react-codemirror)
  - [CodeMirror 6 lang-markdown](https://github.com/codemirror/lang-markdown)
  - [CodeMirror 6 Vim mode (cm-vim)](https://github.com/replit/codemirror-vim)
  - [Monaco Editor for React](https://github.com/suren-atoyan/monaco-react)
  - [Milkdown](https://milkdown.dev/)
- 備考:
  - 現行の `@uiw/react-md-editor@4.1.1`（global.css L2398〜2438）の CSS オーバーライドは置き換え後に削除する
  - ライブラリ最終選定は実装着手前に確認を取ること
