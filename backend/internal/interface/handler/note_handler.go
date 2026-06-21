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

// maxNoteBodyBytes はリクエストボディの最大サイズ制限です（2MB）。
const maxNoteBodyBytes = 2 << 20

// NoteHandler はノート管理に関するHTTPハンドラです。
type NoteHandler struct {
	noteUC *usecase.NoteUseCase
}

// NewNoteHandler は NoteHandler を生成します。
func NewNoteHandler(noteUC *usecase.NoteUseCase) *NoteHandler {
	return &NoteHandler{noteUC: noteUC}
}

// HandleCreateNote は POST /api/v1/teams/{team_id}/notes を処理します（チームメンバーのみ）。
func (h *NoteHandler) HandleCreateNote(w http.ResponseWriter, r *http.Request) {
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

	var req CreateNoteRequestDTO
	r.Body = http.MaxBytesReader(w, r.Body, maxNoteBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "タイトルは必須です"})
		return
	}

	input := usecase.CreateNoteInput{
		CallerID:         callerID,
		CallerRole:       callerRole,
		Title:            req.Title,
		Body:             req.Body,
		DiscussionPoints: req.DiscussionPoints,
		Memo:             req.Memo,
		Tags:             req.Tags,
	}
	if req.Status != "" {
		input.Status = domain.NoteStatus(req.Status)
	}

	note, err := h.noteUC.CreateNote(r.Context(), teamID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrInvalidNoteStatus):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("ノート作成でエラーが発生しました", "team_id", teamID, "caller_id", callerID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: toNoteDTO(note)})
}

// HandleListNotes は GET /api/v1/teams/{team_id}/notes を処理します（チームメンバーのみ）。
// クエリパラメータによるタグフィルタリング・キーワード検索・ページネーションに対応しています。
//
// クエリパラメータ:
//   - tag_ids: カンマ区切りのタグID一覧（複数指定時はAND絞り込み）
//   - keyword: タイトル・本文・議論点・メモを対象とした部分一致検索
//   - page: ページ番号（1始まり。省略時は1）
//   - per_page: 1ページあたりの件数（省略時は20、最大100）
func (h *NoteHandler) HandleListNotes(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.noteUC.SearchNotes(r.Context(), teamID, usecase.SearchNotesInput{
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
			slog.Error("ノート一覧取得でエラーが発生しました", "team_id", teamID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	items := make([]NoteDTO, 0, len(result.Items))
	for _, n := range result.Items {
		items = append(items, toNoteDTO(n))
	}

	writeJSON(w, http.StatusOK, response{Data: NoteListResponseDTO{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}})
}

// HandleGetNote は GET /api/v1/teams/{team_id}/notes/{note_id} を処理します（チームメンバーのみ）。
// チームスコープ外のノートや可視性ルール上アクセス不可の場合は404を返します。
func (h *NoteHandler) HandleGetNote(w http.ResponseWriter, r *http.Request) {
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

	note, err := h.noteUC.GetNote(r.Context(), noteID, teamID, usecase.GetNoteInput{
		CallerID:   callerID,
		CallerRole: callerRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrNoteNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ノートが見つかりません"})
		default:
			slog.Error("ノート取得でエラーが発生しました", "team_id", teamID, "note_id", noteID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toNoteDTO(note)})
}

// HandleUpdateNote は PUT /api/v1/teams/{team_id}/notes/{note_id} を処理します（チームオーナー・admin・作成者のみ）。
func (h *NoteHandler) HandleUpdateNote(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateNoteRequestDTO
	r.Body = http.MaxBytesReader(w, r.Body, maxNoteBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	input := usecase.UpdateNoteInput{
		CallerID:         callerID,
		CallerRole:       callerRole,
		Title:            req.Title,
		Body:             req.Body,
		DiscussionPoints: req.DiscussionPoints,
		Memo:             req.Memo,
		TagsSet:          req.Tags != nil,
		Tags:             req.Tags,
	}
	if req.Status != nil {
		s := domain.NoteStatus(*req.Status)
		input.Status = &s
	}

	note, err := h.noteUC.UpdateNote(r.Context(), noteID, teamID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrNoteNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ノートが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		case errors.Is(err, domain.ErrInvalidNoteStatus):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("ノート更新でエラーが発生しました", "team_id", teamID, "note_id", noteID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toNoteDTO(note)})
}

// HandleUpdateNoteVisibility は PATCH /api/v1/teams/{team_id}/notes/{note_id}/visibility を処理します（チームオーナー・admin・作成者のみ）。
func (h *NoteHandler) HandleUpdateNoteVisibility(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateNoteVisibilityRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Status == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "status は必須です"})
		return
	}

	input := usecase.UpdateNoteVisibilityInput{
		CallerID:   callerID,
		CallerRole: callerRole,
		Status:     domain.NoteStatus(req.Status),
	}

	note, err := h.noteUC.UpdateNoteVisibility(r.Context(), noteID, teamID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrNoteNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ノートが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		case errors.Is(err, domain.ErrInvalidNoteStatus):
			writeJSON(w, http.StatusBadRequest, response{Error: err.Error()})
		default:
			slog.Error("ノート公開設定変更でエラーが発生しました", "team_id", teamID, "note_id", noteID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: toNoteDTO(note)})
}

// HandleDeleteNote は DELETE /api/v1/teams/{team_id}/notes/{note_id} を処理します（チームオーナー・admin・作成者のみ）。
func (h *NoteHandler) HandleDeleteNote(w http.ResponseWriter, r *http.Request) {
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

	if err := h.noteUC.DeleteNote(r.Context(), noteID, teamID, callerID, callerRole); err != nil {
		switch {
		case errors.Is(err, domain.ErrMemberNotFound):
			writeJSON(w, http.StatusForbidden, response{Error: "このチームのメンバーではありません"})
		case errors.Is(err, domain.ErrNoteNotFound):
			writeJSON(w, http.StatusNotFound, response{Error: "ノートが見つかりません"})
		case errors.Is(err, domain.ErrPermissionDenied):
			writeJSON(w, http.StatusForbidden, response{Error: "この操作を行う権限がありません"})
		default:
			slog.Error("ノート削除でエラーが発生しました", "team_id", teamID, "note_id", noteID, "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"message": "ノートを削除しました"}})
}
