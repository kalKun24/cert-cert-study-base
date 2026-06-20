package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// maxCommentBodyBytes はリクエストボディの最大サイズ制限です（100KB）。
const maxCommentBodyBytes = 100 * 1024

// uuidRegexp はUUID形式の正規表現です。
var uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// validateUUID はUUID形式の文字列を検証します。
// 不正な場合は false を返します。
func validateUUID(id string) bool {
	return uuidRegexp.MatchString(id)
}

// CommentHandler はコメント管理に関するHTTPハンドラです。
type CommentHandler struct {
	commentUC *usecase.CommentUseCase
}

// NewCommentHandler は CommentHandler を生成します。
func NewCommentHandler(commentUC *usecase.CommentUseCase) *CommentHandler {
	return &CommentHandler{commentUC: commentUC}
}

// HandleCreateComment は POST /api/v1/teams/{team_id}/questions/{id}/comments を処理します。
func (h *CommentHandler) HandleCreateComment(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}

	questionID := r.PathValue("id")
	if questionID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}
	if !validateUUID(questionID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCommentBodyBytes)
	var req CreateCommentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	comment, err := h.commentUC.CreateComment(r.Context(), usecase.CreateCommentInput{
		QuestionID: questionID,
		TeamID:     teamID,
		Body:       req.Body,
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCommentBodyEmpty):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		case errors.Is(err, domain.ErrCommentBodyTooLong):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied), errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		default:
			slog.Error("コメント投稿でエラーが発生しました", "team_id", teamID, "question_id", questionID, "caller_id", callerID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toCommentDTO(comment.Comment, comment.DisplayName)})
}

// HandleListComments は GET /api/v1/teams/{team_id}/questions/{id}/comments を処理します。
func (h *CommentHandler) HandleListComments(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}

	questionID := r.PathValue("id")
	if questionID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}
	if !validateUUID(questionID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	comments, err := h.commentUC.ListComments(r.Context(), usecase.ListCommentsInput{
		QuestionID: questionID,
		TeamID:     teamID,
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied), errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		default:
			slog.Error("コメント一覧取得でエラーが発生しました", "team_id", teamID, "question_id", questionID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	dtos := make([]CommentDTO, 0, len(comments))
	for _, c := range comments {
		dtos = append(dtos, toCommentDTO(c.Comment, c.DisplayName))
	}

	writeJSON(w, http.StatusOK, response{Data: dtos})
}

// HandleUpdateComment は PUT /api/v1/teams/{team_id}/questions/{id}/comments/{comment_id} を処理します。
func (h *CommentHandler) HandleUpdateComment(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}

	questionID := r.PathValue("id")
	if questionID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}
	if !validateUUID(questionID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDの形式が不正です"})
		return
	}

	commentID := r.PathValue("comment_id")
	if commentID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "コメントIDは必須です"})
		return
	}
	if !validateUUID(commentID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "コメントIDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCommentBodyBytes)
	var req UpdateCommentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	comment, err := h.commentUC.UpdateComment(r.Context(), usecase.UpdateCommentInput{
		QuestionID: questionID,
		TeamID:     teamID,
		CommentID:  commentID,
		Body:       req.Body,
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCommentBodyEmpty):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		case errors.Is(err, domain.ErrCommentBodyTooLong):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrCommentNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "コメントが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied), errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		default:
			slog.Error("コメント編集でエラーが発生しました", "team_id", teamID, "question_id", questionID, "comment_id", commentID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toCommentDTO(comment.Comment, comment.DisplayName)})
}

// HandleDeleteComment は DELETE /api/v1/teams/{team_id}/questions/{id}/comments/{comment_id} を処理します。
func (h *CommentHandler) HandleDeleteComment(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}

	questionID := r.PathValue("id")
	if questionID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}
	if !validateUUID(questionID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDの形式が不正です"})
		return
	}

	commentID := r.PathValue("comment_id")
	if commentID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "コメントIDは必須です"})
		return
	}
	if !validateUUID(commentID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "コメントIDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	if err := h.commentUC.DeleteComment(r.Context(), usecase.DeleteCommentInput{
		QuestionID: questionID,
		TeamID:     teamID,
		CommentID:  commentID,
		CallerID:   callerID,
		CallerRole: callerRole,
	}); err != nil {
		switch {
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrCommentNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "コメントが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied), errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		default:
			slog.Error("コメント削除でエラーが発生しました", "team_id", teamID, "question_id", questionID, "comment_id", commentID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "コメントを削除しました"}})
}
