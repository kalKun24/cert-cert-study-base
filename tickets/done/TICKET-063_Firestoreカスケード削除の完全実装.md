# TICKET-063 Firestoreカスケード削除の完全実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-063 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | 2026-06-22 |
| ブランチ名 | `feature/TICKET-063` |
| PR番号 | #50 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/50 |

---

## 概要

Firestore マイグレーションで導入されたチーム・問題・ノートの Delete() が、サブコレクション（comments / notes / questions / tags）を削除しないため、削除操作のたびに孤立データが無制限に蓄積する問題を修正する。

---

## 背景・目的

Firestore はドキュメントを削除してもサブコレクションを自動削除しない。GCS 時代はチームキーのファイルが自然にスコープされていたため同問題は顕在化しなかったが、Firestore 移行後は以下3箇所でカスケード削除が未実装になっている。

- `backend/internal/infrastructure/firestore/team_repository.go:194` — members のみ削除。questions/{id}/comments、notes/{id}/comments、tags/ が残存する。
- `backend/internal/infrastructure/firestore/question_repository.go:190` — 問題ドキュメントのみ削除。comments サブコレクションが残存する。
- `backend/internal/infrastructure/firestore/note_repository.go:186` — ノートドキュメントのみ削除。comments サブコレクションが残存する。

加えて `backend/internal/infrastructure/firestore/tag_repository.go:159` の Delete() 使用中チェックは questionRepo のみを確認しており、ノートのみが参照するタグを誤って削除できる状態になっている。

---

## 受け入れ条件

- [x] `team_repository.go` の Delete() が questions/{id}/comments、notes/{id}/comments、tags/ を再帰的に削除する
- [x] `question_repository.go` の Delete() が questions/{id}/comments サブコレクションを削除する
- [x] `note_repository.go` の Delete() が notes/{id}/comments サブコレクションを削除する
- [x] `tag_repository.go` の Delete() の使用中チェックがノートへの参照も確認する（noteRepo.FindByTagID 相当）
- [x] 削除後に孤立ドキュメントが残らないことをユニットテスト（Firestore エミュレータ or モック）で確認する

---

## サブチケット（コミット単位）

- [x] `fix(firestore): question / note の Delete() にコメントサブコレクション削除を追加`
- [x] `fix(firestore): team の Delete() にカスケード削除（questions/notes/tags とその入れ子）を追加`
- [x] `fix(firestore): tag の Delete() 使用中チェックにノート参照確認を追加`
- [x] `test(firestore): カスケード削除の孤立ドキュメント残存チェックをテストに追加`

---

## 関連情報

- 関連チケット: -
- 参考: Firestore ドキュメント「サブコレクションの削除」https://firebase.google.com/docs/firestore/manage-data/delete-data
- 備考: ブロッカー。このバグは feature/firestore-migration ブランチが本番マージされると即座にデータ汚染が始まる。他のチケットより優先して対応すること。検出エージェント: Code Reviewer（ブロッカー #1〜#4）。品質チェック日: 2026-06-22。
