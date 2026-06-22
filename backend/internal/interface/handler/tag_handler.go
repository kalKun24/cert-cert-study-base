package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// TagHandler はタグ管理に関するHTTPハンドラです。
type TagHandler struct {
	tagUC *usecase.TagUseCase
}

// NewTagHandler は TagHandler を生成します。
func NewTagHandler(tagUC *usecase.TagUseCase) *TagHandler {
	return &TagHandler{tagUC: tagUC}
}

// HandleListTags は GET /api/v1/teams/{team_id}/tags を処理します（チームメンバー or admin）。
func (h *TagHandler) HandleListTags(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" || !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	tags, err := h.tagUC.ListTags(r.Context(), callerID, callerRole, teamID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのタグを参照する権限がありません"})
		default:
			slog.Error("タグ一覧取得でエラーが発生しました", "team_id", teamID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	dtos := make([]TagDTO, 0, len(tags))
	for _, t := range tags {
		dtos = append(dtos, toTagDTO(t))
	}

	writeJSON(w, http.StatusOK, response{Data: dtos})
}

// HandleCreateTag は POST /api/v1/teams/{team_id}/tags を処理します（チームメンバー or admin）。
func (h *TagHandler) HandleCreateTag(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" || !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	var req CreateTagRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "タグ名は必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	tag, err := h.tagUC.CreateTag(r.Context(), callerID, callerRole, teamID, usecase.CreateTagInput{Name: req.Name})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームにタグを作成する権限がありません"})
		case errors.Is(err, domain.ErrTagNameAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		default:
			slog.Error("タグ作成でエラーが発生しました", "team_id", teamID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toTagDTO(tag)})
}

// HandleUpdateTag は PUT /api/v1/teams/{team_id}/tags/{id} を処理します（admin のみ）。
func (h *TagHandler) HandleUpdateTag(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" || !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "タグIDは必須です"})
		return
	}
	if !validateUUID(id) {
		writeJSON(w, http.StatusBadRequest, response{Error: "タグIDの形式が不正です"})
		return
	}

	var req UpdateTagRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	_, callerRole := callerInfo(r)
	tag, err := h.tagUC.UpdateTag(r.Context(), callerRole, teamID, id, usecase.UpdateTagInput{Name: req.Name})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "タグを更新する権限がありません"})
		case errors.Is(err, domain.ErrTagNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "タグが見つかりません"})
		case errors.Is(err, domain.ErrTagTeamMismatch):
			writeJSON(w, http.StatusNotFound, response{Error: "タグが見つかりません"})
		case errors.Is(err, domain.ErrTagNameAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		default:
			slog.Error("タグ更新でエラーが発生しました", "team_id", teamID, "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toTagDTO(tag)})
}

// HandleDeleteTag は DELETE /api/v1/teams/{team_id}/tags/{id} を処理します（チームメンバー or admin）。
func (h *TagHandler) HandleDeleteTag(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" || !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "タグIDは必須です"})
		return
	}
	if !validateUUID(id) {
		writeJSON(w, http.StatusBadRequest, response{Error: "タグIDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if err := h.tagUC.DeleteTag(r.Context(), callerID, callerRole, teamID, id); err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのタグを削除する権限がありません"})
		case errors.Is(err, domain.ErrTagNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "タグが見つかりません"})
		case errors.Is(err, domain.ErrTagTeamMismatch):
			writeJSON(w, http.StatusNotFound, response{Error: "タグが見つかりません"})
		case errors.Is(err, domain.ErrTagInUse):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		default:
			slog.Error("タグ削除でエラーが発生しました", "team_id", teamID, "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "タグを削除しました"}})
}
