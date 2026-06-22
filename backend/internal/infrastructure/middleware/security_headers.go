package middleware

import "net/http"

// SecurityHeaders はセキュリティ関連の HTTP レスポンスヘッダーを全レスポンスに付与するミドルウェアです。
// 付与するヘッダー:
//   - X-Content-Type-Options: nosniff          - MIME タイプスニッフィングを防止する
//   - X-Frame-Options: DENY                    - クリックジャッキング攻撃を防止する
//   - Referrer-Policy: strict-origin-when-cross-origin - リファラ情報の漏洩を制限する
//   - Permissions-Policy                       - 不要なブラウザ機能へのアクセスを無効化する
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		next.ServeHTTP(w, r)
	})
}
