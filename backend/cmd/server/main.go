// Package main はバックエンドサーバーのエントリーポイントです。
// HTTPサーバーを起動し、ヘルスチェックおよびAPI v1エンドポイントを提供します。
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/auth"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/repository"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/seed"
	gcsStorage "github.com/kalKun24/cert-study-base/backend/internal/infrastructure/storage"
	"github.com/kalKun24/cert-study-base/backend/internal/interface/handler"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// response は統一レスポンスフォーマットです。
type response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func main() {
	// 構造化ログ（JSON形式）をセットアップ
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// ポート番号として有効な範囲（1〜65535）を検証
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		slog.Error("PORT 環境変数が無効です", "port", port)
		os.Exit(1)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET 環境変数が設定されていません")
		os.Exit(1)
	}

	gcsBucket := os.Getenv("GCS_BUCKET")
	if gcsBucket == "" {
		slog.Error("GCS_BUCKET 環境変数が設定されていません")
		os.Exit(1)
	}

	ctx := context.Background()

	// GCS_EMULATOR_HOST が設定されている場合はエミュレータへ、未設定時は実 GCS へ接続
	gcsClient, err := gcsStorage.NewClientFromEnv(ctx)
	if err != nil {
		slog.Error("GCSクライアントの初期化に失敗しました", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := gcsClient.Close(); err != nil {
			slog.Warn("GCSクライアントのクローズに失敗しました", "error", err)
		}
	}()

	// エミュレータ使用時はバケットが存在しないため起動時に作成する
	if os.Getenv("GCS_EMULATOR_HOST") != "" {
		if err := gcsStorage.EnsureBucketExists(ctx, gcsClient, gcsBucket); err != nil {
			slog.Error("エミュレータバケットの作成に失敗しました", "error", err)
			os.Exit(1)
		}
		slog.Info("GCS エミュレータを使用します", "host", os.Getenv("GCS_EMULATOR_HOST"), "bucket", gcsBucket)
	}

	// StorageClient アダプター（GCS → 独自インターフェース）
	sc := gcsStorage.NewGCSStorageClient(gcsClient)

	// 依存関係を構築（コンポジションルート）
	userRepo := repository.NewGCSUserRepository(sc, gcsBucket)
	bcryptHasher := auth.NewBcryptHasher()
	jwtManager := auth.NewJWTManager(jwtSecret)

	// 初回起動時に admin ユーザーが存在しない場合のみ seed を実行
	if err := seed.SeedAdminIfNeeded(ctx, userRepo, bcryptHasher); err != nil {
		slog.Error("admin seed に失敗しました", "error", err)
		os.Exit(1)
	}

	teamRepo := repository.NewGCSTeamRepository(sc, gcsBucket)
	invitationRepo := repository.NewGCSInvitationRepository(sc, gcsBucket)

	authUC := usecase.NewAuthUseCase(userRepo, bcryptHasher, jwtManager)
	userUC := usecase.NewUserUseCase(userRepo, bcryptHasher)
	invitationUC := usecase.NewInvitationUseCase(invitationRepo, teamRepo, userRepo)

	questionRepo := repository.NewGCSQuestionRepository(sc, gcsBucket)
	questionUC := usecase.NewQuestionUseCase(questionRepo, teamRepo)

	commentRepo := repository.NewGCSCommentRepository(sc, gcsBucket)
	commentUC := usecase.NewCommentUseCase(commentRepo, questionRepo, userRepo, teamRepo)

	// TeamUseCase はメンバー統計機能のために questionRepo / commentRepo も注入する
	teamUC := usecase.NewTeamUseCase(teamRepo, userRepo, questionRepo, commentRepo)

	tagRepo := repository.NewGCSTagRepository(sc, gcsBucket, questionRepo)
	tagUC := usecase.NewTagUseCase(tagRepo, teamRepo)

	noteRepo := repository.NewGCSNoteRepository(sc, gcsBucket)
	noteUC := usecase.NewNoteUseCase(noteRepo, teamRepo)

	authHandler := handler.NewAuthHandler(authUC)
	userHandler := handler.NewUserHandler(userUC)
	questionHandler := handler.NewQuestionHandler(questionUC)
	commentHandler := handler.NewCommentHandler(commentUC)
	teamHandler := handler.NewTeamHandler(teamUC)
	tagHandler := handler.NewTagHandler(tagUC)
	noteHandler := handler.NewNoteHandler(noteUC)
	invitationHandler := handler.NewInvitationHandler(invitationUC)

	// ルーティング設定
	mux := http.NewServeMux()

	// ヘルスチェック（認証不要）
	mux.HandleFunc("GET /health", handleHealth)

	// 認証エンドポイント（認証不要）
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.HandleLogin)

	// authMiddleware はDBからユーザーの最新 is_active を確認する認証ミドルウェアです。
	// トークン発行後に管理者がユーザーを停止した場合も即座に反映されます。
	authMiddleware := jwtManager.AuthMiddlewareWithRepo(userRepo)

	// ログアウト（JWTの検証のみ、ロール不問）
	mux.Handle("POST /api/v1/auth/logout",
		authMiddleware(http.HandlerFunc(authHandler.HandleLogout)),
	)

	// admin ロールのみ許可するミドルウェアチェーンを組み立てるヘルパー
	withAdmin := func(h http.HandlerFunc) http.Handler {
		return authMiddleware(
			auth.RequireRole(domain.RoleAdmin)(h),
		)
	}

	// ユーザー管理（admin のみ）
	mux.Handle("GET /api/v1/users", withAdmin(userHandler.HandleListUsers))
	mux.Handle("POST /api/v1/users", withAdmin(userHandler.HandleCreateUser))
	mux.Handle("GET /api/v1/users/{id}", withAdmin(userHandler.HandleGetUser))
	mux.Handle("PUT /api/v1/users/{id}", withAdmin(userHandler.HandleUpdateUser))
	mux.Handle("DELETE /api/v1/users/{id}", withAdmin(userHandler.HandleDeleteUser))
	mux.Handle("PATCH /api/v1/users/{id}/status", withAdmin(userHandler.HandleUpdateUserStatus))
	// グローバルチームオーナー権限付与・剥奪（admin のみ）
	mux.Handle("PATCH /api/v1/admin/users/{id}/team-owner", withAdmin(userHandler.HandleUpdateTeamOwnerStatus))

	// 認証済みユーザーのみ許可するミドルウェアチェーンを組み立てるヘルパー
	withAuth := func(h http.HandlerFunc) http.Handler {
		return authMiddleware(h)
	}

	// 自分のプロフィール編集・パスワード変更（全ロール可）
	// Go 1.22+ の net/http では /users/me/... は /users/{id}/... より具体的なパターンが優先されるため
	// 登録順序に依存せず "me" がパスパラメータに誤マッチすることはない
	mux.Handle("PATCH /api/v1/users/me/profile", withAuth(userHandler.HandleUpdateMyProfile))
	mux.Handle("PATCH /api/v1/users/me/password", withAuth(userHandler.HandleChangeMyPassword))

	// チーム管理（認証済み全ユーザー。各エンドポイントで認可チェックをユースケース層が実施）
	mux.Handle("POST /api/v1/teams", withAuth(teamHandler.HandleCreateTeam))
	mux.Handle("GET /api/v1/teams", withAuth(teamHandler.HandleListTeams))
	mux.Handle("GET /api/v1/teams/{id}", withAuth(teamHandler.HandleGetTeam))
	mux.Handle("PUT /api/v1/teams/{id}", withAuth(teamHandler.HandleUpdateTeam))
	mux.Handle("DELETE /api/v1/teams/{id}", withAuth(teamHandler.HandleDeleteTeam))
	mux.Handle("GET /api/v1/teams/{id}/members", withAuth(teamHandler.HandleListTeamMembers))
	mux.Handle("POST /api/v1/teams/{id}/members", withAuth(teamHandler.HandleAddTeamMember))
	// /members/me は /members/{user_id} より具体的なパターンが優先される（Go 1.22+ net/http）
	mux.Handle("DELETE /api/v1/teams/{id}/members/me", withAuth(teamHandler.HandleLeaveTeam))
	mux.Handle("DELETE /api/v1/teams/{id}/members/{user_id}", withAuth(teamHandler.HandleRemoveTeamMember))
	// チーム内メンバーロール変更（認証済みユーザー、認可はユースケース層で実施）
	mux.Handle("PATCH /api/v1/teams/{id}/members/{user_id}/role", withAuth(teamHandler.HandleChangeMemberRole))

	// 招待管理（認証済みユーザー。認可はユースケース層で実施）
	mux.Handle("POST /api/v1/teams/{id}/invitations", withAuth(invitationHandler.HandleSendInvitation))
	mux.Handle("GET /api/v1/invitations/me", withAuth(invitationHandler.HandleListMyInvitations))
	mux.Handle("PATCH /api/v1/invitations/{id}", withAuth(invitationHandler.HandleRespondInvitation))

	// タグ管理（チームスコープ: 一覧・作成・削除はチームメンバー or admin、更新は admin のみ）
	mux.Handle("GET /api/v1/teams/{team_id}/tags", withAuth(tagHandler.HandleListTags))
	mux.Handle("POST /api/v1/teams/{team_id}/tags", withAuth(tagHandler.HandleCreateTag))
	mux.Handle("PUT /api/v1/teams/{team_id}/tags/{id}", withAuth(tagHandler.HandleUpdateTag))
	mux.Handle("DELETE /api/v1/teams/{team_id}/tags/{id}", withAuth(tagHandler.HandleDeleteTag))

	// 問題CRUD（チームスコープ: チームメンバーのみアクセス可能。admin も例外なしでメンバーチェック）
	mux.Handle("POST /api/v1/teams/{team_id}/questions", withAuth(questionHandler.HandleCreateQuestion))
	mux.Handle("GET /api/v1/teams/{team_id}/questions", withAuth(questionHandler.HandleListQuestions))
	mux.Handle("GET /api/v1/teams/{team_id}/questions/{id}", withAuth(questionHandler.HandleGetQuestion))
	mux.Handle("PUT /api/v1/teams/{team_id}/questions/{id}", withAuth(questionHandler.HandleUpdateQuestion))
	mux.Handle("DELETE /api/v1/teams/{team_id}/questions/{id}", withAuth(questionHandler.HandleDeleteQuestion))
	// 公開設定変更（作成者本人または admin のみ。認可はユースケース層で実施）
	mux.Handle("PATCH /api/v1/teams/{team_id}/questions/{id}/visibility", withAuth(questionHandler.HandleUpdateQuestionVisibility))

	// コメントCRUD（チームスコープ: チームメンバーかつ問題の閲覧権限を持つユーザー）
	mux.Handle("POST /api/v1/teams/{team_id}/questions/{id}/comments", withAuth(commentHandler.HandleCreateComment))
	mux.Handle("GET /api/v1/teams/{team_id}/questions/{id}/comments", withAuth(commentHandler.HandleListComments))
	mux.Handle("PUT /api/v1/teams/{team_id}/questions/{id}/comments/{comment_id}", withAuth(commentHandler.HandleUpdateComment))
	mux.Handle("DELETE /api/v1/teams/{team_id}/questions/{id}/comments/{comment_id}", withAuth(commentHandler.HandleDeleteComment))

	// ノートCRUD（チームスコープ: チームメンバーのみアクセス可能。編集・削除はチームオーナー・admin・作成者のみ）
	mux.Handle("POST /api/v1/teams/{team_id}/notes", withAuth(noteHandler.HandleCreateNote))
	mux.Handle("GET /api/v1/teams/{team_id}/notes", withAuth(noteHandler.HandleListNotes))
	mux.Handle("GET /api/v1/teams/{team_id}/notes/{note_id}", withAuth(noteHandler.HandleGetNote))
	mux.Handle("PUT /api/v1/teams/{team_id}/notes/{note_id}", withAuth(noteHandler.HandleUpdateNote))
	mux.Handle("DELETE /api/v1/teams/{team_id}/notes/{note_id}", withAuth(noteHandler.HandleDeleteNote))
	// ノート公開設定変更（チームオーナー・admin・作成者のみ。認可はユースケース層で実施）
	mux.Handle("PATCH /api/v1/teams/{team_id}/notes/{note_id}/visibility", withAuth(noteHandler.HandleUpdateNoteVisibility))

	srv := &http.Server{
		Addr:         net.JoinHostPort("", port),
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// グレースフルシャットダウンのためにシグナルを待機
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		slog.Info("シャットダウンシグナルを受信しました。サーバーを停止します。")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("サーバーのシャットダウンに失敗しました", "error", err)
		}
		close(idleConnsClosed)
	}()

	slog.Info("サーバーを起動します", "port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("サーバーの起動に失敗しました", "error", err)
		os.Exit(1)
	}

	<-idleConnsClosed
	slog.Info("サーバーを正常に停止しました")
}

// handleHealth はヘルスチェックエンドポイントのハンドラです。
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, response{
		Data: map[string]string{
			"status": "ok",
		},
	})
}

// loggingMiddleware はリクエストの構造化ログを出力するミドルウェアです。
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		slog.Info("HTTPリクエスト",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"latency_ms", time.Since(start).Milliseconds(),
		)
	})
}

// responseWriter はHTTPステータスコードを記録するResponseWriterのラッパーです。
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// writeJSON はJSONレスポンスを書き込むヘルパー関数です。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("JSONエンコードに失敗しました", "error", err)
	}
}
