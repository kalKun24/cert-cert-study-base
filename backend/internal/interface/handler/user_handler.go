package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// UserHandler はユーザー管理に関するHTTPハンドラです。
type UserHandler struct {
	userUC *usecase.UserUseCase
}

// NewUserHandler は UserHandler を生成します。
func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

// HandleListUsers は GET /api/v1/users を処理します（admin のみ）。
func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userUC.ListUsers(r.Context())
	if err != nil {
		slog.Error("ユーザー一覧取得でエラーが発生しました", "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	dtos := make([]UserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, toUserDTO(u))
	}

	writeJSON(w, http.StatusOK, response{Data: dtos})
}

// HandleCreateUser は POST /api/v1/users を処理します（admin のみ）。
func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Username == "" || req.DisplayName == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "username, display_name, email, password, role は必須です"})
		return
	}

	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, response{Error: "パスワードは8文字以上で入力してください"})
		return
	}

	user, err := h.userUC.CreateUser(r.Context(), usecase.CreateUserInput{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Password:    req.Password,
		Role:        domain.Role(req.Role),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUsernameAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidRole):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("ユーザー作成でエラーが発生しました", "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toUserDTO(user)})
}

// HandleGetUser は GET /api/v1/users/{id} を処理します（admin のみ）。
func (h *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "ユーザーIDは必須です"})
		return
	}

	user, err := h.userUC.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, response{Error: "ユーザーが見つかりません"})
			return
		}
		slog.Error("ユーザー取得でエラーが発生しました", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toUserDTO(user)})
}

// HandleUpdateUser は PUT /api/v1/users/{id} を処理します（admin のみ）。
func (h *UserHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "ユーザーIDは必須です"})
		return
	}

	var req UpdateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	// パスワードが指定されている場合は長さを検証
	if req.Password != nil && len(*req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, response{Error: "パスワードは8文字以上で入力してください"})
		return
	}

	input := usecase.UpdateUserInput{
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Password:    req.Password,
	}
	if req.Role != nil {
		r := domain.Role(*req.Role)
		input.Role = &r
	}

	user, err := h.userUC.UpdateUser(r.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ユーザーが見つかりません"})
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidRole):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("ユーザー更新でエラーが発生しました", "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toUserDTO(user)})
}

// HandleDeleteUser は DELETE /api/v1/users/{id} を処理します（admin のみ）。
func (h *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "ユーザーIDは必須です"})
		return
	}

	if err := h.userUC.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, response{Error: "ユーザーが見つかりません"})
			return
		}
		slog.Error("ユーザー削除でエラーが発生しました", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "ユーザーを削除しました"}})
}

// HandleUpdateMyProfile は PATCH /api/v1/users/me/profile を処理します（全ロール可）。
// ログイン中のユーザー自身の display_name を変更します。
func (h *UserHandler) HandleUpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が不正です"})
		return
	}

	var req UpdateMyProfileRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.DisplayName == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "display_name は必須です"})
		return
	}

	user, err := h.userUC.UpdateProfile(r.Context(), usecase.UpdateProfileInput{
		UserID:      userID,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, response{Error: "ユーザーが見つかりません"})
			return
		}
		slog.Error("プロフィール更新でエラーが発生しました", "user_id", userID, "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toUserDTO(user)})
}

// HandleChangeMyPassword は PATCH /api/v1/users/me/password を処理します（全ロール可）。
// ログイン中のユーザー自身のパスワードを変更します。
// 現在のパスワードが誤っている場合は HTTP 422 を返します。
func (h *UserHandler) HandleChangeMyPassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が不正です"})
		return
	}

	var req ChangeMyPasswordRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "current_password と new_password は必須です"})
		return
	}

	if len(req.NewPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, response{Error: "新しいパスワードは8文字以上で入力してください"})
		return
	}
	// bcrypt は 72 バイトで入力を切り捨てるため、超過分は無効化される
	if len(req.NewPassword) > 72 {
		writeJSON(w, http.StatusBadRequest, response{Error: "新しいパスワードは72文字以下で入力してください"})
		return
	}

	err := h.userUC.ChangePassword(r.Context(), usecase.ChangePasswordInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCurrentPasswordIncorrect):
			writeJSON(w, http.StatusUnprocessableEntity, response{Error: "現在のパスワードが正しくありません"})
		case errors.Is(err, domain.ErrUserNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ユーザーが見つかりません"})
		default:
			slog.Error("パスワード変更でエラーが発生しました", "user_id", userID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{})
}

// HandleUpdateTeamOwnerStatus は PATCH /api/v1/admin/users/{id}/team-owner を処理します（admin のみ）。
// ユーザーのグローバルチームオーナー権限と作成可能チーム数の上限を更新します。
func (h *UserHandler) HandleUpdateTeamOwnerStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "ユーザーIDは必須です"})
		return
	}

	var req UpdateTeamOwnerStatusRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.MaxTeams < 0 {
		writeJSON(w, http.StatusBadRequest, response{Error: "max_teams は0以上の値を指定してください"})
		return
	}

	user, err := h.userUC.UpdateTeamOwnerStatus(r.Context(), usecase.UpdateTeamOwnerStatusInput{
		UserID:      id,
		IsTeamOwner: req.IsTeamOwner,
		MaxTeams:    req.MaxTeams,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, response{Error: "ユーザーが見つかりません"})
			return
		}
		slog.Error("チームオーナー権限更新でエラーが発生しました", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toUserDTO(user)})
}

// HandleUpdateUserStatus は PATCH /api/v1/users/{id}/status を処理します（admin のみ）。
func (h *UserHandler) HandleUpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "ユーザーIDは必須です"})
		return
	}

	var req UpdateUserStatusRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	user, err := h.userUC.UpdateUserStatus(r.Context(), id, req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, response{Error: "ユーザーが見つかりません"})
			return
		}
		slog.Error("ユーザーステータス更新でエラーが発生しました", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toUserDTO(user)})
}
