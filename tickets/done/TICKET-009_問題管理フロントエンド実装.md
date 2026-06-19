# TICKET-009 問題管理フロントエンド実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-009 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-17 |
| 着手日 | 2026-06-18 |
| 完了日 | 2026-06-18 |
| ブランチ名 | `feature/question-frontend` |
| PR番号 | #12 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/12 |

---

## 概要

問題の作成・閲覧・編集・削除を行うフロントエンドUIをReact + TypeScriptで実装する。Markdownエディタとプレビューを搭載し、タグの付与・一覧表示にも対応する。

---

## 背景・目的

バックエンドAPIが整備されたあと、ユーザーが実際に問題を管理するためのUIが必要。Markdownで問題・解答・解説・議論点メモを直感的に作成・編集できる体験を提供する。

---

## 受け入れ条件

- [x] 問題一覧ページ: 問題のタイトル・タグ・作成日を一覧表示し、タグフィルタ・キーワード検索ができる
- [x] 問題詳細ページ: Markdownをレンダリングして表示する（body / answer / explanation / memo）
- [x] 問題作成ページ: Markdownエディタで入力し、タグを複数選択して保存できる
- [x] 問題編集ページ: 既存データを読み込んでMarkdownエディタで編集・保存できる
- [x] 問題削除: 確認ダイアログ付きで削除できる
- [ ] 未認証ユーザーはログイン画面にリダイレクトされる
- [x] ESLint / Prettier がエラーなく通る

---

## サブチケット（コミット単位）

- [x] `chore(frontend): APIクライアント（fetch wrapper）とルーティング基盤を実装`
- [x] `feat(frontend): 問題一覧ページを実装（タグフィルタ・キーワード検索含む）`
- [x] `feat(frontend): 問題詳細ページ（Markdownレンダリング）を実装`
- [x] `feat(frontend): 問題作成・編集ページ（Markdownエディタ + タグ選択）を実装`
- [x] `feat(frontend): 問題削除機能（確認ダイアログ）を実装`
- [ ] `feat(frontend): 認証状態管理とログイン画面・認証ガードを実装`

---

## 関連情報

- 関連チケット: TICKET-001（前提）、TICKET-002（認証API）、TICKET-004（問題CRUD API）、TICKET-005（タグAPI）、TICKET-008（検索API）
- 参考: CLAUDE.md「React」セクション（関数コンポーネント + Hooks、PascalCase.tsx）
- 備考: フロントエンドの状態管理ライブラリは未選定（CLAUDE.md TODO）。着手前に方針を確認する。MarkdownエディタはOSSライブラリを選定して使用する
