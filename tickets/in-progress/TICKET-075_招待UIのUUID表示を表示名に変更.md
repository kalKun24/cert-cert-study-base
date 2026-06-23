# TICKET-075 招待UIのUUID表示を表示名に変更

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-075 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-23 |
| 着手日 | 2026-06-23 |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-075` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

チームへの招待一覧画面（`/invitations`）でチームIDと招待者がUUIDのまま表示されている問題を修正する。
バックエンドのAPIレスポンスにチーム名（`team_name`）と招待者の表示名（`inviter_display_name`）を追加し、フロントエンドでそれらを表示するよう変更する。

---

## 背景・目的

現在、招待一覧画面では `team_id` と `invited_by` フィールドの値（UUIDの文字列）がそのまま表示されており、ユーザーにとって意味のある情報になっていない。チーム名と招待者の表示名を表示することで、どのチームへ誰から招待されたのかが一目でわかるようにする。

---

## 受け入れ条件

- [ ] `GET /api/v1/invitations/me` のレスポンスに `team_name`（チームの表示名）が含まれる
- [ ] `GET /api/v1/invitations/me` のレスポンスに `inviter_display_name`（招待者の DisplayName）が含まれる
- [ ] `api/openapi.yaml` の `Invitation` スキーマに上記2フィールドが追加されている
- [ ] フロントエンドの招待一覧画面でUUIDの代わりにチーム名・招待者の表示名が表示される
- [ ] チーム名・招待者名の取得に失敗した場合もAPIはエラーにならず、フォールバック値（空文字またはUUID）を返す

---

## サブチケット（コミット単位）

- [x] `docs(api): Invitationスキーマにteam_nameとinviter_display_nameを追加`
- [x] `feat(invitation): ListMyInvitationsの出力にチーム名・招待者表示名を追加`
- [x] `feat(invitation): 招待一覧UIをUUID表示から表示名表示に変更`

---

## 関連情報

- 関連チケット: なし
- 参考: `backend/internal/interface/handler/dto.go` の `InvitationDTO`、`frontend/src/pages/InvitationListPage.tsx`
- 備考:
  - `GET /api/v1/invitations/{id}` の `RespondInvitation` レスポンスは受諾/拒否後の処理用途なので今回のスコープ外とする
  - ユースケース層でリポジトリを呼び出して情報を補完する。ハンドラ層（Interface）でリポジトリを直接呼ぶことはしない
