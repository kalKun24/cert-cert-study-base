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

// HandleListTags は GET /api/v1/tags を処理します（認証済みユーザー）。
func (h *TagHandler) HandleListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.tagUC.ListTags(r.Context())
	if err != nil {
		slog.Error("タグ一覧取得でエラーが発生しました", "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	dtos := make([]TagDTO, 0, len(tags))
	for _, t := range tags {
		dtos = append(dtos, toTagDTO(t))
	}

	writeJSON(w, http.StatusOK, response{Data: dtos})
}

// HandleCreateTag は POST /api/v1/tags を処理します（admin のみ）。
func (h *TagHandler) HandleCreateTag(w http.ResponseWriter, r *http.Request) {
	var req CreateTagRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "タグ名は必須です"})
		return
	}

	tag, err := h.tagUC.CreateTag(r.Context(), usecase.CreateTagInput{Name: req.Name})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTagNameAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		default:
			slog.Error("タグ作成でエラーが発生しました", "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toTagDTO(tag)})
}

// HandleUpdateTag は PUT /api/v1/tags/{id} を処理します（admin のみ）。
func (h *TagHandler) HandleUpdateTag(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "タグIDは必須です"})
		return
	}

	var req UpdateTagRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	tag, err := h.tagUC.UpdateTag(r.Context(), id, usecase.UpdateTagInput{Name: req.Name})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTagNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "タグが見つかりません"})
		case errors.Is(err, domain.ErrTagNameAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		default:
			slog.Error("タグ更新でエラーが発生しました", "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toTagDTO(tag)})
}

// HandleDeleteTag は DELETE /api/v1/tags/{id} を処理します（admin のみ）。
func (h *TagHandler) HandleDeleteTag(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "タグIDは必須です"})
		return
	}

	if err := h.tagUC.DeleteTag(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, domain.ErrTagNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "タグが見つかりません"})
		case errors.Is(err, domain.ErrTagInUse):
			writeJSON(w, http.StatusConflict, response{Error: err.Error()})
		default:
			slog.Error("タグ削除でエラーが発生しました", "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "タグを削除しました"}})
}
