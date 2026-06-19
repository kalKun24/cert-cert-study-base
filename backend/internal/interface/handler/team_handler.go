package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// TeamHandler はチーム管理に関するHTTPハンドラです。
type TeamHandler struct {
	teamUC *usecase.TeamUseCase
}

// NewTeamHandler は TeamHandler を生成します。
func NewTeamHandler(teamUC *usecase.TeamUseCase) *TeamHandler {
	return &TeamHandler{teamUC: teamUC}
}

// callerInfo はコンテキストから呼び出し元のIDとロールを取得します。
func callerInfo(r *http.Request) (id string, role domain.Role) {
	id, _ = r.Context().Value(ContextKeyUserID).(string)
	roleStr, _ := r.Context().Value(ContextKeyUserRole).(string)
	return id, domain.Role(roleStr)
}

// HandleCreateTeam は POST /api/v1/teams を処理します（admin / teamowner のみ）。
func (h *TeamHandler) HandleCreateTeam(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)

	var req CreateTeamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "name は必須です"})
		return
	}

	team, err := h.teamUC.CreateTeam(r.Context(), usecase.CreateTeamInput{
		Name:        req.Name,
		Description: req.Description,
		CallerID:    callerID,
		CallerRole:  callerRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		case errors.Is(err, domain.ErrTeamNameAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		default:
			slog.Error("チーム作成でエラーが発生しました", "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toTeamDTO(team)})
}

// HandleListTeams は GET /api/v1/teams を処理します。
func (h *TeamHandler) HandleListTeams(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)

	teams, err := h.teamUC.ListTeams(r.Context(), usecase.ListTeamsInput{
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		slog.Error("チーム一覧取得でエラーが発生しました", "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	dtos := make([]TeamDTO, 0, len(teams))
	for _, t := range teams {
		dtos = append(dtos, toTeamDTO(t))
	}

	writeJSON(w, http.StatusOK, response{Data: dtos})
}

// HandleGetTeam は GET /api/v1/teams/{id} を処理します。
func (h *TeamHandler) HandleGetTeam(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)
	teamID := r.PathValue("id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}

	out, err := h.teamUC.GetTeam(r.Context(), callerID, callerRole, teamID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "チームが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		default:
			slog.Error("チーム取得でエラーが発生しました", "id", teamID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	memberDTOs := make([]TeamMemberDTO, 0, len(out.Members))
	for _, m := range out.Members {
		memberDTOs = append(memberDTOs, toTeamMemberDTO(m))
	}

	writeJSON(w, http.StatusOK, response{Data: TeamDetailDTO{
		TeamDTO: toTeamDTO(out.Team),
		Members: memberDTOs,
	}})
}

// HandleUpdateTeam は PUT /api/v1/teams/{id} を処理します。
func (h *TeamHandler) HandleUpdateTeam(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)
	teamID := r.PathValue("id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}

	var req UpdateTeamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Name != nil && *req.Name == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "name を空文字にすることはできません"})
		return
	}

	team, err := h.teamUC.UpdateTeam(r.Context(), callerID, callerRole, teamID, usecase.UpdateTeamInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "チームが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		case errors.Is(err, domain.ErrTeamNameAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		default:
			slog.Error("チーム更新でエラーが発生しました", "id", teamID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toTeamDTO(team)})
}

// HandleDeleteTeam は DELETE /api/v1/teams/{id} を処理します。
func (h *TeamHandler) HandleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)
	teamID := r.PathValue("id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}

	if err := h.teamUC.DeleteTeam(r.Context(), callerID, callerRole, teamID); err != nil {
		switch {
		case errors.Is(err, domain.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "チームが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		default:
			slog.Error("チーム削除でエラーが発生しました", "id", teamID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "チームを削除しました"}})
}

// HandleAddTeamMember は POST /api/v1/teams/{id}/members を処理します。
func (h *TeamHandler) HandleAddTeamMember(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)
	teamID := r.PathValue("id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}

	var req AddTeamMemberRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "user_id は必須です"})
		return
	}

	member, err := h.teamUC.AddMember(r.Context(), callerID, callerRole, teamID, req.UserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "チームが見つかりません"})
		case errors.Is(err, domain.ErrUserNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ユーザーが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		case errors.Is(err, domain.ErrMemberAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		default:
			slog.Error("メンバー追加でエラーが発生しました", "team_id", teamID, "user_id", req.UserID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toTeamMemberDTO(member)})
}

// HandleChangeMemberRole は PATCH /api/v1/teams/{id}/members/{user_id}/role を処理します。
// チームの per-team owner または admin のみが実行可能です。
func (h *TeamHandler) HandleChangeMemberRole(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)
	teamID := r.PathValue("id")
	userID := r.PathValue("user_id")
	if teamID == "" || userID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDとユーザーIDは必須です"})
		return
	}

	var req ChangeMemberRoleRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Role == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "role は必須です"})
		return
	}

	role := domain.MemberRole(req.Role)
	if !role.IsValid() {
		writeJSON(w, http.StatusBadRequest, response{Error: "role は owner または member を指定してください"})
		return
	}

	member, err := h.teamUC.ChangeMemberRole(r.Context(), usecase.ChangeMemberRoleInput{
		CallerID:     callerID,
		CallerRole:   callerRole,
		TeamID:       teamID,
		TargetUserID: userID,
		Role:         role,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		case errors.Is(err, domain.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "チームが見つかりません"})
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "メンバーが見つかりません"})
		case errors.Is(err, domain.ErrLastTeamOwner):
			writeJSON(w, http.StatusUnprocessableEntity, response{Error: err.Error()})
		default:
			slog.Error("メンバーロール変更でエラーが発生しました", "team_id", teamID, "user_id", userID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toTeamMemberDTO(member)})
}

// HandleRemoveTeamMember は DELETE /api/v1/teams/{id}/members/{user_id} を処理します。
func (h *TeamHandler) HandleRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)
	teamID := r.PathValue("id")
	userID := r.PathValue("user_id")
	if teamID == "" || userID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDとユーザーIDは必須です"})
		return
	}

	if err := h.teamUC.RemoveMember(r.Context(), callerID, callerRole, teamID, userID); err != nil {
		switch {
		case errors.Is(err, domain.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "チームが見つかりません"})
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "メンバーが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		default:
			slog.Error("メンバー除外でエラーが発生しました", "team_id", teamID, "user_id", userID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "メンバーを除外しました"}})
}
