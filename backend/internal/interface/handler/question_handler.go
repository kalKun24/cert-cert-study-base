package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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
// クエリパラメータによるタグフィルタリング・キーワード検索・ページネーションに対応しています。
//
// クエリパラメータ:
//   - tag_ids: カンマ区切りのタグID一覧（複数指定時はAND絞り込み）
//   - keyword: タイトル・問題文・解説・メモを対象とした部分一致検索
//   - page: ページ番号（1始まり。省略時は1）
//   - per_page: 1ページあたりの件数（省略時は20、最大100）
func (h *QuestionHandler) HandleListQuestions(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)

	// tag_ids クエリパラメータのパース（カンマ区切り、空要素除去）
	var tagIDs []string
	if raw := r.URL.Query().Get("tag_ids"); raw != "" {
		for _, tid := range strings.Split(raw, ",") {
			tid = strings.TrimSpace(tid)
			if tid != "" {
				tagIDs = append(tagIDs, tid)
			}
		}
	}

	// keyword クエリパラメータのパース
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))

	// page クエリパラメータのパース・バリデーション
	page := 1
	if rawPage := r.URL.Query().Get("page"); rawPage != "" {
		parsed, err := strconv.Atoi(rawPage)
		if err != nil || parsed < 1 {
			writeJSON(w, http.StatusBadRequest, response{Error: "page は1以上の整数で指定してください"})
			return
		}
		page = parsed
	}

	// per_page クエリパラメータのパース・バリデーション
	perPage := 20
	if rawPerPage := r.URL.Query().Get("per_page"); rawPerPage != "" {
		parsed, err := strconv.Atoi(rawPerPage)
		if err != nil || parsed < 1 || parsed > 100 {
			writeJSON(w, http.StatusBadRequest, response{Error: "per_page は1〜100の整数で指定してください"})
			return
		}
		perPage = parsed
	}

	result, err := h.questionUC.SearchQuestions(r.Context(), usecase.SearchQuestionsInput{
		CallerID:   callerID,
		CallerRole: callerRole,
		TagIDs:     tagIDs,
		Keyword:    keyword,
		Page:       page,
		PerPage:    perPage,
	})
	if err != nil {
		slog.Error("問題一覧取得でエラーが発生しました", "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	items := make([]QuestionDTO, 0, len(result.Items))
	for _, q := range result.Items {
		items = append(items, toQuestionDTO(q))
	}

	writeJSON(w, http.StatusOK, response{Data: QuestionListResponseDTO{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}})
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
