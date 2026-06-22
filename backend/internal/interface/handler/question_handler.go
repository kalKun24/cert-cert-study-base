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

// maxQuestionBodyBytes はリクエストボディの最大サイズ制限です（2MB）。
const maxQuestionBodyBytes = 2 << 20

// maxQuestionTagFilterCount は tag_ids フィルタの最大件数です。
// Firestore の array-contains クエリは1クエリ1タグのみサーバーサイド処理し、
// 残りはメモリフィルタとなるため、過大なリストによる DoS を防ぐために件数を制限します。
const maxQuestionTagFilterCount = 10

// QuestionHandler は問題管理に関するHTTPハンドラです。
type QuestionHandler struct {
	questionUC *usecase.QuestionUseCase
}

// NewQuestionHandler は QuestionHandler を生成します。
func NewQuestionHandler(questionUC *usecase.QuestionUseCase) *QuestionHandler {
	return &QuestionHandler{questionUC: questionUC}
}

// HandleCreateQuestion は POST /api/v1/teams/{team_id}/questions を処理します（チームメンバーのみ）。
func (h *QuestionHandler) HandleCreateQuestion(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	var req CreateQuestionRequestDTO
	r.Body = http.MaxBytesReader(w, r.Body, maxQuestionBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "タイトルは必須です"})
		return
	}

	input := usecase.CreateQuestionInput{
		CallerID:    callerID,
		CallerRole:  callerRole,
		Title:       req.Title,
		Body:        req.Body,
		Answer:      req.Answer,
		Explanation: req.Explanation,
		Memo:        req.Memo,
		Tags:        req.Tags,
	}
	if req.Status != "" {
		input.Status = domain.QuestionStatus(req.Status)
	}

	question, err := h.questionUC.CreateQuestion(r.Context(), teamID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrInvalidQuestionStatus):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("問題作成でエラーが発生しました", "team_id", teamID, "caller_id", callerID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toQuestionDTO(question)})
}

// HandleListQuestions は GET /api/v1/teams/{team_id}/questions を処理します（チームメンバーのみ）。
// クエリパラメータによるタグフィルタリング・キーワード検索・ページネーションに対応しています。
//
// クエリパラメータ:
//   - tag_ids: カンマ区切りのタグID一覧（複数指定時はAND絞り込み、最大10件、UUID形式必須）
//   - keyword: タイトル・問題文・解説・メモを対象とした部分一致検索
//   - page: ページ番号（1始まり。省略時は1）
//   - per_page: 1ページあたりの件数（省略時は20、最大100）
func (h *QuestionHandler) HandleListQuestions(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)

	// tag_ids クエリパラメータのパース・バリデーション（カンマ区切り、UUID形式チェック、件数上限）
	var tagIDs []string
	if raw := r.URL.Query().Get("tag_ids"); raw != "" {
		for _, tid := range strings.Split(raw, ",") {
			tid = strings.TrimSpace(tid)
			if tid == "" {
				continue
			}
			if !validateUUID(tid) {
				writeJSON(w, http.StatusBadRequest, response{Error: "tag_ids に不正な形式のIDが含まれています"})
				return
			}
			tagIDs = append(tagIDs, tid)
		}
		if len(tagIDs) > maxQuestionTagFilterCount {
			writeJSON(w, http.StatusBadRequest, response{Error: "tag_ids は最大10件まで指定できます"})
			return
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

	result, err := h.questionUC.SearchQuestions(r.Context(), teamID, usecase.SearchQuestionsInput{
		CallerID:   callerID,
		CallerRole: callerRole,
		TagIDs:     tagIDs,
		Keyword:    keyword,
		Page:       page,
		PerPage:    perPage,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		default:
			slog.Error("問題一覧取得でエラーが発生しました", "team_id", teamID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
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

// HandleGetQuestion は GET /api/v1/teams/{team_id}/questions/{id} を処理します（チームメンバーのみ）。
// チームスコープ外の問題や可視性ルール上アクセス不可の場合は404を返します。
func (h *QuestionHandler) HandleGetQuestion(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)

	question, err := h.questionUC.GetQuestion(r.Context(), id, teamID, usecase.GetQuestionInput{
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		default:
			slog.Error("問題取得でエラーが発生しました", "team_id", teamID, "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toQuestionDTO(question)})
}

// HandleUpdateQuestion は PUT /api/v1/teams/{team_id}/questions/{id} を処理します（作成者本人または admin のみ）。
func (h *QuestionHandler) HandleUpdateQuestion(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

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
	r.Body = http.MaxBytesReader(w, r.Body, maxQuestionBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	input := usecase.UpdateQuestionInput{
		CallerID:    callerID,
		CallerRole:  callerRole,
		Title:       req.Title,
		Body:        req.Body,
		Answer:      req.Answer,
		Explanation: req.Explanation,
		Memo:        req.Memo,
		TagsSet:     req.Tags != nil,
		Tags:        req.Tags,
	}
	if req.Status != nil {
		s := domain.QuestionStatus(*req.Status)
		input.Status = &s
	}

	question, err := h.questionUC.UpdateQuestion(r.Context(), id, teamID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidQuestionStatus):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("問題更新でエラーが発生しました", "team_id", teamID, "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toQuestionDTO(question)})
}

// HandleUpdateQuestionVisibility は PATCH /api/v1/teams/{team_id}/questions/{id}/visibility を処理します（作成者本人または admin のみ）。
func (h *QuestionHandler) HandleUpdateQuestionVisibility(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

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
		CallerID:   callerID,
		CallerRole: callerRole,
		Status:     domain.QuestionStatus(req.Status),
	}

	question, err := h.questionUC.UpdateQuestionVisibility(r.Context(), id, teamID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		case errors.Is(err, domain.ErrInvalidQuestionStatus):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("問題公開設定変更でエラーが発生しました", "team_id", teamID, "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toQuestionDTO(question)})
}

// HandleDeleteQuestion は DELETE /api/v1/teams/{team_id}/questions/{id} を処理します（作成者本人または admin のみ）。
func (h *QuestionHandler) HandleDeleteQuestion(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

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

	if err := h.questionUC.DeleteQuestion(r.Context(), id, teamID, callerID, callerRole); err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		default:
			slog.Error("問題削除でエラーが発生しました", "team_id", teamID, "id", id, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "問題を削除しました"}})
}
