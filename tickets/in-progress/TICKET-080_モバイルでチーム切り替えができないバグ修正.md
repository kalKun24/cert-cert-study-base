# TICKET-080 モバイルでチーム切り替えができないバグ修正

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-080 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-29 |
| 着手日 | 2026-06-29 |
| 完了日 | YYYY-MM-DD |
| ブランチ名 | `feature/TICKET-080` |
| PR番号 | #82 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/82 |

---

## 概要

スマートフォン（モバイル幅）でチームの切り替えができない不具合を修正する。モバイルナビゲーション（ハンバーガードロワー）にチーム切り替えUIを追加し、どの画面幅でもアクティブチームを変更できるようにする。

---

## 背景・目的

デスクトップではトップバーにチーム切り替えセレクト（`.topbar-team-area`）が表示されるが、モバイルでは利用できず、複数チームに所属するユーザーがスマホからチームを切り替えられない。

調査により以下が判明している（根本原因）:

1. **CSSでデスクトップ用セレクトを非表示**
   - `frontend/src/styles/global.css` の `@media (max-width: 768px)` 内、`.topbar-team-area { display: none; }`（おおよそ 2154-2156 行付近）
2. **モバイルドロワーに代替UIが存在しない**
   - `frontend/src/components/Layout.tsx` のモバイルドロワー（143-278 行付近）はナビリンク・ユーザー情報・ログアウトのみで、チーム切り替えが含まれていない
3. 状態管理自体は正常
   - `frontend/src/components/NavBar.tsx`（126-146 行付近のセレクト、45-50 行の `handleTeamChange`）と `frontend/src/context/TeamContext.tsx`（`setActiveTeam` / localStorage `activeTeamId`）は機能しており、UIへの導線のみが欠落している

目的は、モバイルでもチーム切り替え操作を可能にし、複数チーム所属ユーザーのUXを回復すること。

---

## 受け入れ条件

- [x] モバイル幅（≤768px）でチーム切り替えUIにアクセスできる（ハンバーガードロワー内など）
- [x] モバイルでチームを選択するとアクティブチームが切り替わり、localStorage（`activeTeamId`）に反映される
- [x] 所属チームが1つ以下の場合は切り替えUIを表示しない（デスクトップの挙動と一致） ※方針決定（レビュー指摘1・選択肢A）により、モバイル・デスクトップ（NavBar）とも `teams.length > 1` に統一
- [x] デスクトップのチーム切り替え挙動・表示にデグレが無い ※表示条件をモバイルと統一し、不一致を解消
- [x] 表示文言はすべてロケール/コンテンツJSONで管理（ハードコード禁止）
- [x] 既存のフロントエンドテストが通り、必要に応じてテストを追加する（Layout.test.tsx 4件・NavBar.test.tsx 4件追加、全66件PASS）

---

## サブチケット（コミット単位）

- [x] `fix(nav): モバイルドロワーにチーム切り替えUIを追加`（1f6fbd3）
- [x] `test(nav): モバイルチーム切り替えのテストを追加`（2b9e348）
- [x] `fix(nav): デスクトップのチーム切り替え表示条件をモバイルと統一`（10ca444）
- [x] `test(nav): NavBar のチーム切り替え表示境界テストを追加`（46aefa6）

---

## 関連情報

- 関連チケット: TICKET-037（ログイン後チーム選択フロー）, TICKET-012（チーム管理フロントエンド）
- 参考:
  - `frontend/src/components/NavBar.tsx`
  - `frontend/src/components/Layout.tsx`
  - `frontend/src/context/TeamContext.tsx`
  - `frontend/src/styles/global.css`
- 備考: バグ修正のため `fix` タイプで起票。
