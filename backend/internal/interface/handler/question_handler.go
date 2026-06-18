package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// QuestionHandler は問題管理に関するHTTPハンドラです。
type QuestionHandler struct {
	questionUC *usecase.QuestionUseCase
}

// NewQuestionHandler は QuestionHandler を生成します。
func NewQuestionHandler(questionUC *usecase.QuestionUseCase) *QuestionHandler {
	return &QuestionHandler{questionUC: questionUC}
}

// HandleCreateQuestion は POST /api/v1/questions を処理します（認証済みユーザー）。
func (h *QuestionHandler) HandleCreateQuestion(w http.ResponseWriter, r *http.Request) {
	callerID, _ := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	var req CreateQuestionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "タイトルは必須です"})
		return
	}

	input := usecase.CreateQuestionInput{
		CallerID:         callerID,
		Title:            req.Title,
		Body:             req.Body,
		Answer:           req.Answer,
		Explanation:      req.Explanation,
		Memo:             req.Memo,
		Tags:             req.Tags,
		PublishedTeamIDs: req.PublishedTeamIDs,
	}
	if req.Status != "" {
		input.Status = domain.QuestionStatus(req.Status)
	}
	if req.VisibilityScope != "" {
		input.VisibilityScope = domain.VisibilityScope(req.VisibilityScope)
	}

	question, err := h.questionUC.CreateQuestion(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidQuestionStatus):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidVisibilityScope):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("問題作成でエラーが発生しました", "caller_id", callerID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toQuestionDTO(question)})
}

// HandleListQuestions は GET /api/v1/questions を処理します（認証済みユーザー）。
// 可視性ルールに基づいてフィルタリングした問題一覧を返します。
func (h *QuestionHandler) HandleListQuestions(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)

	questions, err := h.questionUC.ListQuestions(r.Context(), usecase.ListQuestionsInput{
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		slog.Error("問題一覧取得でエラーが発生しました", "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	dtos := make([]QuestionDTO, 0, len(questions))
	for _, q := range questions {
		dtos = append(dtos, toQuestionDTO(q))
	}

	writeJSON(w, http.StatusOK, response{Data: dtos})
}

// HandleGetQuestion は GET /api/v1/questions/{id} を処理します（認証済みユーザー）。
// 可視性ルールに基づき、閲覧不可の場合は404を返します。
func (h *QuestionHandler) HandleGetQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)

	question, err := h.questionUC.GetQuestion(r.Context(), id, usecase.GetQuestionInput{
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		if errors.Is(err, domain.ErrQuestionNotFound) {
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
			return
		}
		slog.Error("問題取得でエラーが発生しました", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toQuestionDTO(question)})
}

// HandleUpdateQuestion は PUT /api/v1/questions/{id} を処理します（作成者本人または admin のみ）。
func (h *QuestionHandler) HandleUpdateQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	var req UpdateQuestionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	input := usecase.UpdateQuestionInput{
		CallerID:            callerID,
		CallerRole:          callerRole,
		Title:               req.Title,
		Body:                req.Body,
		Answer:              req.Answer,
		Explanation:         req.Explanation,
		Memo:                req.Memo,
		TagsSet:             req.Tags != nil,
		Tags:                req.Tags,
		PublishedTeamIDsSet: req.PublishedTeamIDs != nil,
		PublishedTeamIDs:    req.PublishedTeamIDs,
	}
	if req.Status != nil {
		s := domain.QuestionStatus(*req.Status)
		input.Status = &s
	}
	if req.VisibilityScope != nil {
		vs := domain.VisibilityScope(*req.VisibilityScope)
		input.VisibilityScope = &vs
	}

	question, err := h.questionUC.UpdateQuestion(r.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidQuestionStatus):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidVisibilityScope):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("問題更新でエラーが発生しました", "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toQuestionDTO(question)})
}

// HandleUpdateQuestionVisibility は PATCH /api/v1/questions/{id}/visibility を処理します（作成者本人または admin のみ）。
func (h *QuestionHandler) HandleUpdateQuestionVisibility(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	var req UpdateQuestionVisibilityRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Status == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "status は必須です"})
		return
	}

	input := usecase.UpdateQuestionVisibilityInput{
		CallerID:            callerID,
		CallerRole:          callerRole,
		Status:              domain.QuestionStatus(req.Status),
		PublishedTeamIDsSet: req.PublishedTeamIDs != nil,
		PublishedTeamIDs:    req.PublishedTeamIDs,
	}
	if req.VisibilityScope != nil {
		vs := domain.VisibilityScope(*req.VisibilityScope)
		input.VisibilityScope = &vs
	}

	question, err := h.questionUC.UpdateQuestionVisibility(r.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidQuestionStatus):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidVisibilityScope):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("問題公開設定変更でエラーが発生しました", "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toQuestionDTO(question)})
}

// HandleDeleteQuestion は DELETE /api/v1/questions/{id} を処理します（作成者本人または admin のみ）。
func (h *QuestionHandler) HandleDeleteQuestion(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	if err := h.questionUC.DeleteQuestion(r.Context(), id, callerID, callerRole); err != nil {
		switch {
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		default:
			slog.Error("問題削除でエラーが発生しました", "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "問題を削除しました"}})
}
