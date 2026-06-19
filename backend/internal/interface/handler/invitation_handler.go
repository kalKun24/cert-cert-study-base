package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// InvitationHandler は招待管理に関するHTTPハンドラです。
type InvitationHandler struct {
	invitationUC *usecase.InvitationUseCase
}

// NewInvitationHandler は InvitationHandler を生成します。
func NewInvitationHandler(invitationUC *usecase.InvitationUseCase) *InvitationHandler {
	return &InvitationHandler{invitationUC: invitationUC}
}

// HandleSendInvitation は POST /api/v1/teams/{id}/invitations を処理します。
// チームの per-team owner または admin のみが実行可能です。
func (h *InvitationHandler) HandleSendInvitation(w http.ResponseWriter, r *http.Request) {
	callerID, callerRole := callerInfo(r)
	teamID := r.PathValue("id")
	if teamID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "チームIDは必須です"})
		return
	}

	var req SendInvitationRequestDTO
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.InviteeIdentifier == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "invitee_identifier は必須です"})
		return
	}

	inv, err := h.invitationUC.SendInvitation(r.Context(), usecase.SendInvitationInput{
		CallerID:          callerID,
		CallerRole:        callerRole,
		TeamID:            teamID,
		InviteeIdentifier: req.InviteeIdentifier,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "チームが見つかりません"})
		case errors.Is(err, domain.ErrUserNotFound):
			// ユーザー列挙攻撃を防ぐため 422 で返し、存在有無を推測させない
			writeJSON(w, http.StatusUnprocessableEntity, response{Error: "指定した識別子のユーザーを招待できませんでした"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		case errors.Is(err, domain.ErrMemberAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: "このユーザーはすでにチームのメンバーです"})
		case errors.Is(err, domain.ErrInvitationAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: "同一チームへの招待がすでに存在します"})
		default:
			slog.Error("招待送信でエラーが発生しました", "team_id", teamID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toInvitationDTO(inv)})
}

// HandleListMyInvitations は GET /api/v1/invitations/me を処理します。
func (h *InvitationHandler) HandleListMyInvitations(w http.ResponseWriter, r *http.Request) {
	callerID, _ := callerInfo(r)

	invitations, err := h.invitationUC.ListMyInvitations(r.Context(), usecase.ListMyInvitationsInput{
		CallerID: callerID,
	})
	if err != nil {
		slog.Error("招待一覧取得でエラーが発生しました", "caller_id", callerID, "error", err)
		writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		return
	}

	dtos := make([]InvitationDTO, 0, len(invitations))
	for _, inv := range invitations {
		dtos = append(dtos, toInvitationDTO(inv))
	}

	writeJSON(w, http.StatusOK, response{Data: dtos})
}

// HandleRespondInvitation は PATCH /api/v1/invitations/{id} を処理します。
// 招待されたユーザーのみが実行可能です。
func (h *InvitationHandler) HandleRespondInvitation(w http.ResponseWriter, r *http.Request) {
	callerID, _ := callerInfo(r)
	invitationID := r.PathValue("id")
	if invitationID == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "招待IDは必須です"})
		return
	}

	var req RespondInvitationRequestDTO
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Status == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "status は必須です"})
		return
	}

	status := domain.InvitationStatus(req.Status)
	if status != domain.StatusAccepted && status != domain.StatusRejected {
		writeJSON(w, http.StatusBadRequest, response{Error: "status は accepted または rejected を指定してください"})
		return
	}

	inv, err := h.invitationUC.RespondInvitation(r.Context(), usecase.RespondInvitationInput{
		CallerID:     callerID,
		InvitationID: invitationID,
		Status:       status,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvitationNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "招待が見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		case errors.Is(err, domain.ErrInvitationNotPending):
			writeJSON(w, http.StatusUnprocessableEntity, response{Error: "この招待はすでに処理済みです"})
		case errors.Is(err, domain.ErrMemberAlreadyExists):
			writeJSON(w, http.StatusConflict, response{Error: "このユーザーはすでにチームのメンバーです"})
		default:
			slog.Error("招待応答でエラーが発生しました", "invitation_id", invitationID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toInvitationDTO(inv)})
}
