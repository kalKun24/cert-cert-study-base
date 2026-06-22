# TICKET-064 openapi.yaml と実装の乖離解消（API First 原則回復）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-064 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | - |
| ブランチ名 | feature/TICKET-064 |
| PR番号 | #49 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/49 |

---

## 概要

CLAUDE.md の API First 原則に違反している openapi.yaml と実装の乖離を3点修正する。NoteComment への display_name 追加、エラーレスポンスの omitempty 問題解消、UserDTO の last_login_at 欠落対応。

---

## 背景・目的

CLAUDE.md は「新しいAPIを追加・変更する場合は、必ず openapi.yaml を先に更新してから実装する」を定めているが、現状以下の乖離が存在する。

1. **NoteCommentDTO の display_name 未反映**: `backend/internal/interface/handler/dto.go:328` の NoteCommentDTO に `display_name` が実装済みだが、`api/openapi.yaml` の NoteComment スキーマ（L2686-2726）に `display_name` プロパティが存在しない。
2. **response struct の omitempty によるエラーレスポンス仕様違反**: `dto.go:14` の `Data any json:"data,omitempty"` により、エラー時のレスポンスが仕様の `{"data":null,"error":"..."}` ではなく `{"error":"..."}` となる。全エンドポイントに横断的な影響がある。
3. **UserDTO に last_login_at が欠落**: 仕様の User スキーマに `last_login_at` が定義されているが、`toUserDTO()` 関数で変換されておらず、ユーザー取得・作成・更新の全レスポンスでフィールドが欠落する。

---

## 受け入れ条件

- [ ] `api/openapi.yaml` の NoteComment スキーマに `display_name` プロパティを追加し required に含める
- [ ] `dto.go` の response struct から `Data` フィールドの `omitempty` タグを除去し、エラー時に `{"data":null,"error":"..."}` 形式で返ることを確認する
- [ ] `UserDTO` に `LastLoginAt *time.Time` を追加し `toUserDTO()` で変換する
- [ ] `InvitationDTO` の `invitee_identifier` 欠落について、セキュリティ上の除外判断を仕様側に反映（required から削除してコメント追記）するか、マスキング値を返すかを決定して実装する

---

## サブチケット（コミット単位）

- [x] `docs(api): NoteComment スキーマに display_name を追加`
- [x] `fix(handler): response struct の omitempty を除去してエラーレスポンス形式を仕様に合わせる`
- [ ] `fix(handler): UserDTO に last_login_at フィールドを追加`
- [ ] `docs(api): InvitationDTO の invitee_identifier 除外方針を仕様にコメント追記`

---

## 関連情報

- 関連チケット: -
- 参考: CLAUDE.md「API First」項目
- 備考: ブロッカー（omitempty 問題は全エンドポイントに影響、display_name 未反映は API First 原則違反）。検出エージェント: Code Reviewer（ブロッカー #5）、API Tester（HIGH #1、#3、MEDIUM #4）。品質チェック日: 2026-06-22。
