# TICKET-068 Firestore クエリのパフォーマンス改善（N+1・フルスキャン解消）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-068 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | - |
| ブランチ名 | feature/TICKET-068 |
| PR番号 | #53 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/53 |

---

## 概要

`ListByOwnerOrMember` の N+1 クエリパターンと、SearchByTeam のコレクション全件メモリロードを Firestore ネイティブなクエリに置き換え、将来の本番規模での性能劣化を防ぐ。

---

## 背景・目的

- **N+1 クエリ** (`backend/internal/infrastructure/firestore/team_repository.go:151`): `ListByOwnerOrMember` が全チームを取得した後、各チームに対して `members` サブコレクションを個別にフェッチしている。チーム数 N のループ内で N 回の Firestore RPC が発生し、Firestore の読み取り課金が線形増加する。
- **フルスキャン** (`backend/internal/infrastructure/firestore/note_repository.go:121`・`question_repository.go` 同等箇所): `SearchByTeam` がチームの全ドキュメントを取得してからメモリ内でフィルタリングしている。Firestore の `array-contains` 演算子によるサーバーサイドのタグ絞り込みが利用可能。

現フェーズのデータ量では機能するが、本番運用で問題数・チーム数が増えると性能・コストの両面で影響が出る技術的負債として対応が必要。

---

## 受け入れ条件

- [x] `ListByOwnerOrMember` を Firestore の `collectionGroup` クエリ（`Where("user_id", "==", userID)`）で書き換えるか、チームドキュメントにメンバーIDリストを非正規化する設計を選択・決定し実装する
- [x] `SearchByTeam` の TagIDs フィルタに `array-contains` または `array-contains-any` クエリを使用してサーバーサイドで絞り込む
- [x] キーワード検索のフルスキャン継続が避けられない場合はスケール限界（件数・応答時間）を設計ドキュメントまたはコードコメントに記録する
- [x] 修正後のクエリが既存のユニットテストをパスすることを確認する

---

## サブチケット（コミット単位）

- [x] `refactor(firestore): ListByOwnerOrMember を collectionGroup クエリに書き換え`
- [x] `refactor(firestore): SearchByTeam のタグフィルタに array-contains を使用`
- [x] `docs(firestore): キーワード検索フルスキャンのスケール限界をコードコメントに記録`

---

## 関連情報

- 関連チケット: TICKET-063（カスケード削除）と同じ Firestore 層の改善
- 参考: Firestore collectionGroup インデックス https://firebase.google.com/docs/firestore/query-data/queries#collection-group-query
- 備考: 将来対応。現フェーズでは機能上問題なし。検出エージェント: Code Reviewer（提案 #9, #10）。品質チェック日: 2026-06-22。
