# TICKET-066 セキュリティ HTTP ヘッダーとバリデーション強化

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-066 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | 2026-06-22 |
| ブランチ名 | feature/TICKET-066 |
| PR番号 | #57 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/57 |

---

## 概要

セキュリティ HTTP ヘッダーの追加、リクエストボディサイズ制限の統一、ユーザー名・メールアドレスのフォーマットバリデーション追加、comment 系ハンドラの teamID UUID バリデーション補完を行う。

---

## 背景・目的

- **セキュリティヘッダー完全欠落** (`backend/cmd/server/main.go`・`frontend/Dockerfile`): CSP・X-Content-Type-Options・X-Frame-Options・HSTS・Referrer-Policy・Permissions-Policy が全て未設定。XSS 成功時の被害拡大・クリックジャッキング・MIME スニッフィングに対して無防備。
- **リクエストボディサイズ制限の不統一**: `question_handler.go`・`note_handler.go`・`comment_handler.go` では `http.MaxBytesReader` を使用しているが、`user_handler.go` と `team_handler.go` では未実装。
- **ユーザー名・メールアドレスのバリデーション欠落** (`backend/internal/usecase/auth.go`・`user_handler.go`): tag には `tagNameMaxLength` による長さ制限があるが、ユーザー関連フィールドには同等のバリデーションが存在しない。
- **comment 系ハンドラで teamID の UUID バリデーション欠落** (`backend/internal/interface/handler/comment_handler.go`): questionID・commentID には実施しているが teamID が未確認。

---

## 受け入れ条件

- [ ] セキュリティヘッダーミドルウェアを作成し `X-Content-Type-Options: nosniff`・`X-Frame-Options: DENY`・`Referrer-Policy: strict-origin-when-cross-origin`・`Permissions-Policy` を全レスポンスに設定する
- [ ] `frontend/Dockerfile` の nginx 設定に同等のセキュリティヘッダーを追加する
- [ ] `user_handler.go` と `team_handler.go` の json.Decode 前に `http.MaxBytesReader` を適用する
- [ ] ユーザー作成・更新時に username のフォーマット（英数字・ハイフン・アンダースコア、3〜50文字）と email のフォーマットを検証するバリデーションをユースケース層に追加する
- [ ] `comment_handler.go` の HandleCreateComment / HandleListComments / HandleUpdateComment / HandleDeleteComment に `teamID` の UUID バリデーションを追加する

---

## サブチケット（コミット単位）

- [ ] `feat(middleware): セキュリティHTTPヘッダーミドルウェアを追加`
- [ ] `fix(docker): フロントエンド nginx 設定にセキュリティヘッダーを追加`
- [ ] `fix(handler): user / team ハンドラに MaxBytesReader を追加`
- [ ] `fix(usecase): ユーザー名・メールアドレスのフォーマットバリデーションを追加`
- [ ] `fix(handler): comment 系ハンドラに teamID の UUID バリデーションを追加`

---

## 関連情報

- 関連チケット: TICKET-065（認証・HTTP保護と合わせてセキュリティ強化）
- 参考: OWASP Secure Headers Project https://owasp.org/www-project-secure-headers/
- 備考: 次スプリント対応。検出エージェント: Security Engineer（High #5、Medium #7, #8）、API Tester（Medium #7）。品質チェック日: 2026-06-22。
