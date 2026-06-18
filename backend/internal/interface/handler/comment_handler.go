package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// CommentHandler はコメント管理に関するHTTPハンドラです。
type CommentHandler struct {
	commentUC *usecase.CommentUseCase
}

// NewCommentHandler は CommentHandler を生成します。
func NewCommentHandler(commentUC *usecase.CommentUseCase) *CommentHandler {
	return &CommentHandler{commentUC: commentUC}
}

// HandleCreateComment は POST /api/v1/questions/{id}/comments を処理します。
func (h *CommentHandler) HandleCreateComment(w http.ResponseWriter, r *http.Request) {
	questionID := r.PathValue("id")
	if questionID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	var req CreateCommentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	comment, err := h.commentUC.CreateComment(r.Context(), usecase.CreateCommentInput{
		QuestionID: questionID,
		Body:       req.Body,
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCommentBodyEmpty):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		default:
			slog.Error("コメント投稿でエラーが発生しました", "question_id", questionID, "caller_id", callerID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toCommentDTO(comment)})
}

// HandleListComments は GET /api/v1/questions/{id}/comments を処理します。
func (h *CommentHandler) HandleListComments(w http.ResponseWriter, r *http.Request) {
	questionID := r.PathValue("id")
	if questionID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)

	comments, err := h.commentUC.ListComments(r.Context(), usecase.ListCommentsInput{
		QuestionID: questionID,
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		default:
			slog.Error("コメント一覧取得でエラーが発生しました", "question_id", questionID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	dtos := make([]CommentDTO, 0, len(comments))
	for _, c := range comments {
		dtos = append(dtos, toCommentDTO(c))
	}

	writeJSON(w, http.StatusOK, response{Data: dtos})
}

// HandleUpdateComment は PUT /api/v1/questions/{id}/comments/{comment_id} を処理します。
func (h *CommentHandler) HandleUpdateComment(w http.ResponseWriter, r *http.Request) {
	questionID := r.PathValue("id")
	if questionID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}

	commentID := r.PathValue("comment_id")
	if commentID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "コメントIDは必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	var req UpdateCommentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	comment, err := h.commentUC.UpdateComment(r.Context(), usecase.UpdateCommentInput{
		QuestionID: questionID,
		CommentID:  commentID,
		Body:       req.Body,
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCommentBodyEmpty):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrCommentNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "コメントが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		default:
			slog.Error("コメント編集でエラーが発生しました", "question_id", questionID, "comment_id", commentID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toCommentDTO(comment)})
}

// HandleDeleteComment は DELETE /api/v1/questions/{id}/comments/{comment_id} を処理します。
func (h *CommentHandler) HandleDeleteComment(w http.ResponseWriter, r *http.Request) {
	questionID := r.PathValue("id")
	if questionID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "問題IDは必須です"})
		return
	}

	commentID := r.PathValue("comment_id")
	if commentID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "コメントIDは必須です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	if err := h.commentUC.DeleteComment(r.Context(), usecase.DeleteCommentInput{
		QuestionID: questionID,
		CommentID:  commentID,
		CallerID:   callerID,
		CallerRole: callerRole,
	}); err != nil {
		switch {
		case errors.Is(err, domain.ErrQuestionNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "問題が見つかりません"})
		case errors.Is(err, domain.ErrCommentNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "コメントが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: err.Error()})
		default:
			slog.Error("コメント削除でエラーが発生しました", "question_id", questionID, "comment_id", commentID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "コメントを削除しました"}})
}
