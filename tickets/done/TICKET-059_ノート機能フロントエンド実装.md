# TICKET-059 ノート機能フロントエンド実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-059 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-21 |
| 着手日 | 2026-06-22 |
| 完了日 | 2026-06-22 |
| ブランチ名 | feature/note-frontend |
| PR番号 | #47 |
| PRリンク | - |

---

## 概要

ノート一覧・詳細・作成・編集ページをフロントエンドに実装する。問題（Question）機能の既存コンポーネント（`CommentSection`・`AccordionSection`・`TagChip`・`Paginator` 等）を最大限流用し、保守性を維持する。

---

## 背景・目的

バックエンドで提供されるノート API（TICKET-057・058）を利用して、ユーザーがブラウザ上でノートを作成・閲覧・編集できる画面を提供する。問題ページと同一のUIコンポーネントを流用することで、学習コストと保守コストを最小化する。

---

## 受け入れ条件

- [ ] `NoteListPage.tsx` が実装されており、以下の機能を持つ
  - タイトル・キーワード・タグによる検索フィルタリング（URL クエリパラメータと同期）
  - ページネーション（`Paginator` コンポーネント流用）
  - ノートカードのクリックで詳細ページへ遷移
  - タグ表示（`TagChip` コンポーネント流用）
- [ ] `NoteDetailPage.tsx` が実装されており、以下の機能を持つ
  - ノート本文・議論点・メモの Markdown レンダリング（`AccordionSection` + `ReactMarkdown` + `rehype-sanitize` 流用）
  - コメント投稿・一覧表示・編集・削除（`CommentSection` コンポーネント流用、またはノート用に拡張）
  - 編集権限を持つユーザー（チームオーナー・admin・作成ユーザー）のみ編集・削除ボタンを表示
  - 公開ステータス（下書き・非公開・公開）の表示と変更
- [ ] `NoteCreatePage.tsx` が実装されており、以下の機能を持つ
  - タイトル・本文・議論点・メモの入力フォーム（Markdown 入力）
  - タグ選択（`TagDropdown` コンポーネント流用）
  - ステータス（下書き・非公開・公開）の選択
- [ ] `NoteEditPage.tsx` が実装されており、`NoteCreatePage` と同じフォームで既存データを初期表示できる
- [ ] API ユーティリティ `frontend/src/utils/noteApi.ts` が追加されている
  - `fetchNotes`・`fetchNote`・`createNote`・`updateNote`・`deleteNote`・`updateNoteVisibility`
- [ ] `frontend/src/types/note.ts` に `Note` 型定義が追加されている
- [ ] ルーティングが追加されている（`/notes`・`/notes/:id`・`/notes/new`・`/notes/:id/edit`）
- [ ] 各ページに適切な i18n キーが追加されている（翻訳ファイル `ja.json` / `en.json` 更新）
- [ ] ESLint・Prettier を通過する

---

## サブチケット（コミット単位）

- [x] `feat(types): Note型定義とnoteApiユーティリティを追加`
- [x] `feat(page): NoteListPageを実装（一覧・検索・ページネーション）`
- [x] `feat(page): NoteDetailPageを実装（Markdown表示・コメント・権限制御）`
- [x] `feat(page): NoteCreatePage・NoteEditPageを実装`
- [x] `feat(router): ノート系ルートを追加`
- [x] `feat(i18n): ノート機能の翻訳キーを追加`

---

## 関連情報

- 関連チケット: TICKET-057（バックエンドAPI、先行必須）、TICKET-058（コメントAPI、先行必須）、TICKET-060（NavBarリンク追加）
- 参考:
  - `frontend/src/pages/QuestionListPage.tsx`
  - `frontend/src/pages/QuestionDetailPage.tsx`
  - `frontend/src/pages/QuestionCreatePage.tsx`
  - `frontend/src/components/CommentSection.tsx`
  - `frontend/src/components/AccordionSection.tsx`
- 備考:
  - `CommentSection` コンポーネントが `questionId` を Props に取る実装になっているため、`noteId` にも対応できるよう汎用化（または別コンポーネント化）が必要か確認すること
  - 問題ページとノートページで同一の `TagDropdown`・`TagChip` を使うため、タグ取得は既存の `fetchTags` ユーティリティを流用する
