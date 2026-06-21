# TICKET-055 複数チーム招待承認時に残り招待が操作不能になるバグ修正

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-055 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-21 |
| 着手日 | 2026-06-21 |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-055` |
| PR番号 | #41 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/41 |

---

## 概要

複数チームへの招待を受けているユーザーが1つの招待を承認すると、残りのチームへの招待をUI上から承認・拒否できなくなるバグを修正する。

---

## 背景・目的

`InvitationListPage` の `handleRespond` で招待を承認した直後に `navigate('/', { replace: true })` でホーム画面に遷移してしまう。
承認によりチームに参加した結果、`TeamContext` の `teams` が更新され、`/` ルートの `TeamSelectionGate` が `teams.length > 0` と判定して `HomePage` を表示するため、残りの pending 招待がある場合でも招待一覧ページへ戻れなくなる。

バックエンド（`RespondInvitation` ユースケース）には問題なく、フロントエンドの画面遷移ロジックのみが原因。

### 修正方針

承認後に残りの pending 招待がある場合は招待一覧画面に留まり、すべての招待を処理し終えてからホームに遷移するよう `handleRespond` のロジックを修正する。

具体的には:
- 承認成功後、招待一覧のステートから該当招待を除去する
- 残りの pending 招待が0件になった時点で `refreshTeams()` してホームへ遷移する
- 残りがある場合は成功メッセージを表示して一覧に留まる

---

## 受け入れ条件

- [x] 招待一覧ページ（`/invitations`）へのナビゲーション導線をナビバーに常設し、pending 招待がある限りいつでもアクセスできる
- [x] 承認・拒否後はホーム画面へ遷移する
- [x] 承認・拒否後もナビバーの招待リンクから再度招待一覧へアクセスできる
- [x] pending 招待が0件のときはナビバーの招待リンクを非表示にする（件数バッジ付きで表示）

---

## サブチケット（コミット単位）

- [x] `fix(invitation): 複数招待がある場合に承認後も招待一覧に留まるよう修正`

---

## 関連情報

- 関連ファイル: `frontend/src/pages/InvitationListPage.tsx`
- 関連ファイル: `frontend/src/App.tsx`（`TeamSelectionGate`）
- 関連ファイル: `frontend/src/context/TeamContext.tsx`
- バグ報告の仮説: フロントエンドの状態（チーム選択）変更による招待一覧ページへの到達不能 → 正解
- バックエンド: 問題なし（`RespondInvitation` ユースケースは複数承認に対応済み）
- 参考: TICKET-051（TeamDetailPage からメンバー一覧への導線追加）
