package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// maxNoteCommentBodyBytes はリクエストボディの最大サイズ制限です（100KB）。
const maxNoteCommentBodyBytes = 100 * 1024

// NoteCommentHandler はノートコメント管理に関するHTTPハンドラです。
type NoteCommentHandler struct {
	noteCommentUC *usecase.NoteCommentUseCase
}

// NewNoteCommentHandler は NoteCommentHandler を生成します。
func NewNoteCommentHandler(noteCommentUC *usecase.NoteCommentUseCase) *NoteCommentHandler {
	return &NoteCommentHandler{noteCommentUC: noteCommentUC}
}

// HandleCreateNoteComment は POST /api/v1/teams/{team_id}/notes/{note_id}/comments を処理します。
func (h *NoteCommentHandler) HandleCreateNoteComment(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	noteID := r.PathValue("note_id")
	if noteID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "ノートIDは必須です"})
		return
	}
	if !validateUUID(noteID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "ノートIDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxNoteCommentBodyBytes)
	var req CreateNoteCommentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	comment, err := h.noteCommentUC.CreateNoteComment(r.Context(), usecase.CreateNoteCommentInput{
		TeamID:     teamID,
		NoteID:     noteID,
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
		case errors.Is(err, domain.ErrNoteNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ノートが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied), errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		default:
			slog.Error("ノートコメント投稿でエラーが発生しました", "team_id", teamID, "note_id", noteID, "caller_id", callerID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toNoteCommentDTO(comment)})
}

// HandleListNoteComments は GET /api/v1/teams/{team_id}/notes/{note_id}/comments を処理します。
func (h *NoteCommentHandler) HandleListNoteComments(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	noteID := r.PathValue("note_id")
	if noteID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "ノートIDは必須です"})
		return
	}
	if !validateUUID(noteID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "ノートIDの形式が不正です"})
		return
	}

	callerID, callerRole := callerInfo(r)
	if callerID == "" {
		writeJSON(w, http.StatusUnauthorized, response{Error: "認証情報が取得できません"})
		return
	}

	comments, err := h.noteCommentUC.ListNoteComments(r.Context(), teamID, noteID, callerID, callerRole)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNoteNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ノートが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied), errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		default:
			slog.Error("ノートコメント一覧取得でエラーが発生しました", "team_id", teamID, "note_id", noteID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	dtos := make([]NoteCommentDTO, 0, len(comments))
	for _, c := range comments {
		dtos = append(dtos, toNoteCommentDTO(c))
	}

	writeJSON(w, http.StatusOK, response{Data: dtos})
}

// HandleUpdateNoteComment は PUT /api/v1/teams/{team_id}/notes/{note_id}/comments/{comment_id} を処理します。
func (h *NoteCommentHandler) HandleUpdateNoteComment(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	noteID := r.PathValue("note_id")
	if noteID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "ノートIDは必須です"})
		return
	}
	if !validateUUID(noteID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "ノートIDの形式が不正です"})
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

	r.Body = http.MaxBytesReader(w, r.Body, maxNoteCommentBodyBytes)
	var req UpdateNoteCommentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	comment, err := h.noteCommentUC.UpdateNoteComment(r.Context(), usecase.UpdateNoteCommentInput{
		TeamID:     teamID,
		NoteID:     noteID,
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
		case errors.Is(err, domain.ErrNoteNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ノートが見つかりません"})
		case errors.Is(err, domain.ErrNoteCommentNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "コメントが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied), errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		default:
			slog.Error("ノートコメント編集でエラーが発生しました", "team_id", teamID, "note_id", noteID, "comment_id", commentID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toNoteCommentDTO(comment)})
}

// HandleDeleteNoteComment は DELETE /api/v1/teams/{team_id}/notes/{note_id}/comments/{comment_id} を処理します。
func (h *NoteCommentHandler) HandleDeleteNoteComment(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("team_id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}
	if !validateUUID(teamID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDの形式が不正です"})
		return
	}

	noteID := r.PathValue("note_id")
	if noteID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "ノートIDは必須です"})
		return
	}
	if !validateUUID(noteID) {
		writeJSON(w, http.StatusBadRequest, response{Error: "ノートIDの形式が不正です"})
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

	if err := h.noteCommentUC.DeleteNoteComment(r.Context(), teamID, noteID, commentID, callerID, callerRole); err != nil {
		switch {
		case errors.Is(err, domain.ErrNoteNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ノートが見つかりません"})
		case errors.Is(err, domain.ErrNoteCommentNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "コメントが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied), errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		default:
			slog.Error("ノートコメント削除でエラーが発生しました", "team_id", teamID, "note_id", noteID, "comment_id", commentID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "コメントを削除しました"}})
}
