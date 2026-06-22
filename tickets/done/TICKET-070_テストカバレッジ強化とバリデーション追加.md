# TICKET-070 テストカバレッジ強化とバリデーション追加

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-070 |
| ステータス | 🟢 完了 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | 2026-06-22 |
| ブランチ名 | `feature/test-coverage-070` |
| PR番号 | #61 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/61 |

---

## 概要

タグUUIDバグ（TICKET-067 で発見）と同種の「型・フォーマット不正が実行時まで検出されない」問題を、ハンドラー・ユースケース両レイヤーのテストで網羅的に防止する。合わせて不足していたバリデーションを実装に追加する。

---

## 背景・目的

TICKET-067 でタグフィールドに UUID ではなく名前文字列が混入する不具合が見つかった。根本原因はハンドラー層の入力バリデーション不足と、ユースケーステストが型の正当性を検証していなかったこと。同じ種類の盲点が複数箇所に存在するため、今回まとめて修正・テスト追加する。

---

## 受け入れ条件

- [x] `status` フィールドに不正値を渡すと 400 が返る（Question/Note 可視性更新）
- [x] ページネーション境界値（`page=0`・`page=-1`・`per_page=0`・`per_page=-5`・`per_page=101`）で 400 が返る
- [x] 可視性ルール全組み合わせ（draft/private/published × member/admin）がユースケース層でテストされている
- [x] admin が他人のリソース（Question/Note/Comment/NoteComment）を更新・削除できることがテストされている
- [x] 一般ユーザーが他人のリソースを更新しようとすると `ErrPermissionDenied` が返ることがテストされている
- [x] tag/user/invitation ハンドラーのパスパラメータに UUID チェックが追加されている
- [x] `go test ./...` が全件 PASS する

---

## サブチケット（コミット単位）

- [x] `fix(handler): Question・Note の可視性更新に status 列挙値バリデーションを追加`
- [x] `test(handler): UpdateVisibility の不正 status・空 status・有効 status テストを追加`
- [x] `test(handler): ページネーション境界値テスト（page/per_page）を追加`
- [x] `test(handler): Note ハンドラーのテストファイルを新規作成（タグ UUID・visibility・ページネーション）`
- [x] `test(usecase): 可視性ルール全組み合わせテストを question_test・note_test に追加`
- [x] `fix(usecase): Comment・NoteComment の更新を admin も許可するよう修正`
- [x] `test(usecase): 権限境界テスト（admin 可・一般ユーザー不可）を追加`
- [x] `fix(handler): tag/team/user/invitation ハンドラーにパスパラメータ UUID バリデーションを追加`

---

## 関連情報

- 関連チケット: TICKET-067（タグ UUID バグ修正）、TICKET-065（セキュリティ基盤強化）
- 備考: フロントエンドのタグ ID/名前混在バグ修正（TagDropdown・CreatePage・EditPage）も同セッションで実施済み
