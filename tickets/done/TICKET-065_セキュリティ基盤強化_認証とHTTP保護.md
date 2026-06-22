# TICKET-065 セキュリティ基盤強化（認証・HTTP保護）

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-065 |
| ステータス | ✅ 完了 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | 2026-06-22 |
| ブランチ名 | feature/TICKET-065 |
| PR番号 | #54 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/54 |

---

## 概要

JWT のトークン転用攻撃リスク・レート制限の欠落・CORS 未設定・Slowloris 脆弱性・localStorage へのトークン保管という複数のセキュリティ課題をまとめて対応する。

---

## 背景・目的

現状、以下のセキュリティ問題が複合している。

- **JWT Issuer/Audience なし** (`backend/internal/infrastructure/auth/jwt.go`): 将来のマイクロサービス化時に トークン転用攻撃が成立する。
- **ログインにレート制限なし** (`backend/internal/interface/handler/auth_handler.go`): ブルートフォース攻撃が無制限に実行でき、bcrypt の計算コストを利用した CPU 枯渇型 DoS にも悪用できる。
- **ReadHeaderTimeout 未設定** (`backend/cmd/server/main.go`): Slowloris 攻撃によるゴルーチン枯渇リスクがある（ReadTimeout は設定済み）。
- **CORS 未設定** (`backend/cmd/server/main.go`): Cloud Run 上でバックエンドが独立サービスとなった場合、任意オリジンからのクロスオリジンリクエストを許可する。
- **JWT を localStorage に保存** (`frontend/src/utils/auth.ts`): XSS 成功時にトークンが盗取される。

---

## 受け入れ条件

- [ ] `jwt.go` の GenerateToken() で `Issuer: "cert-study-base"` と `Audience: ["cert-study-base-api"]` を設定し、ParseToken() でも検証する
- [ ] `main.go` の http.Server に `ReadHeaderTimeout: 5 * time.Second` を追加する
- [ ] ログインエンドポイントにレート制限ミドルウェアを追加する（golang.org/x/time/rate を使い IP 単位で 10回/分 等）
- [ ] 本番環境用の CORS 設定を環境変数ベースで main.go に追加する
- [ ] `auth.ts` を HttpOnly Cookie 方式へ移行するか、リスク受け入れ判断をコメントとして記録する

---

## サブチケット（コミット単位）

- [x] `fix(auth): JWT GenerateToken / ParseToken に Issuer・Audience クレームを追加`
- [x] `fix(server): http.Server に ReadHeaderTimeout を追加`
- [x] `feat(middleware): ログインエンドポイント用レートリミットミドルウェアを追加`
- [x] `feat(middleware): CORS ミドルウェアを追加（本番オリジンを環境変数で設定）`
- [x] `fix(frontend): JWT トークンを HttpOnly Cookie に移行または localStorage 継続の理由をコメントに記録`

---

## 関連情報

- 関連チケット: TICKET-065 との組み合わせで対応すること（セキュリティヘッダーと一緒にレビューが効率的）
- 参考: golang.org/x/time/rate ドキュメント、RFC 7519 JWT クレーム仕様
- 備考: 次スプリント対応。検出エージェント: Security Engineer（High #2, #3, #4, #6、Medium #10）。品質チェック日: 2026-06-22。
