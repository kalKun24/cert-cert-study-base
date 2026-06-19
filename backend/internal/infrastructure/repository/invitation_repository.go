package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/storage"
)

const invitationsObjectName = "invitations.json"

// invitationRecord はGCS上のJSONファイルに保存する招待レコードです。
type invitationRecord struct {
	ID                string    `json:"id"`
	TeamID            string    `json:"team_id"`
	InvitedBy         string    `json:"invited_by"`
	InviteeIdentifier string    `json:"invitee_identifier"`
	InviteeUserID     string    `json:"invitee_user_id"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

func toInvitationRecord(inv *domain.Invitation) invitationRecord {
	return invitationRecord{
		ID:                inv.ID,
		TeamID:            inv.TeamID,
		InvitedBy:         inv.InvitedBy,
		InviteeIdentifier: inv.InviteeIdentifier,
		InviteeUserID:     inv.InviteeUserID,
		Status:            string(inv.Status),
		CreatedAt:         inv.CreatedAt,
	}
}

func toInvitation(r invitationRecord) *domain.Invitation {
	return &domain.Invitation{
		ID:                r.ID,
		TeamID:            r.TeamID,
		InvitedBy:         r.InvitedBy,
		InviteeIdentifier: r.InviteeIdentifier,
		InviteeUserID:     r.InviteeUserID,
		Status:            domain.InvitationStatus(r.Status),
		CreatedAt:         r.CreatedAt,
	}
}

// GCSInvitationRepository はGCS上のJSONファイルに招待データを永続化するリポジトリです。
// domain.InvitationRepository インターフェースを実装します。
type GCSInvitationRepository struct {
	mu      sync.RWMutex
	storage storage.StorageClient
	bucket  string
}

// NewGCSInvitationRepository は GCSInvitationRepository を生成します。
func NewGCSInvitationRepository(sc storage.StorageClient, bucket string) *GCSInvitationRepository {
	return &GCSInvitationRepository{
		storage: sc,
		bucket:  bucket,
	}
}

// loadInvitations はGCSから招待データを読み込みます。
func (r *GCSInvitationRepository) loadInvitations(ctx context.Context) ([]invitationRecord, error) {
	exists, err := r.storage.Exists(ctx, r.bucket, invitationsObjectName)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}
	if !exists {
		return []invitationRecord{}, nil
	}

	rc, err := r.storage.Read(ctx, r.bucket, invitationsObjectName)
	if err != nil {
		return nil, fmt.Errorf("GCS からの読み込みに失敗しました: %w", err)
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			slog.Warn("GCS ReadCloser のクローズに失敗しました", "error", cerr)
		}
	}()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("GCS データの読み取りに失敗しました: %w", err)
	}

	var records []invitationRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("招待データのJSONデコードに失敗しました: %w", err)
	}
	return records, nil
}

// saveInvitations は招待データをGCSに書き込みます。
func (r *GCSInvitationRepository) saveInvitations(ctx context.Context, records []invitationRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("招待データのJSONエンコードに失敗しました: %w", err)
	}
	if err := r.storage.Write(ctx, r.bucket, invitationsObjectName, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("GCS への書き込みに失敗しました: %w", err)
	}
	return nil
}

// Save は招待を新規作成または更新します。
func (r *GCSInvitationRepository) Save(ctx context.Context, inv *domain.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadInvitations(ctx)
	if err != nil {
		return fmt.Errorf("招待データ読み込みに失敗しました: %w", err)
	}

	rec := toInvitationRecord(inv)
	updated := false
	for i, existing := range records {
		if existing.ID == inv.ID {
			records[i] = rec
			updated = true
			break
		}
	}
	if !updated {
		records = append(records, rec)
	}

	if err := r.saveInvitations(ctx, records); err != nil {
		return fmt.Errorf("招待データ保存に失敗しました: %w", err)
	}
	return nil
}

// FindByID はIDで招待を検索します。
func (r *GCSInvitationRepository) FindByID(ctx context.Context, id string) (*domain.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadInvitations(ctx)
	if err != nil {
		return nil, fmt.Errorf("招待データ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.ID == id {
			return toInvitation(rec), nil
		}
	}
	return nil, domain.ErrInvitationNotFound
}

// ListByInvitee は招待先ユーザーIDで招待一覧を返します。
func (r *GCSInvitationRepository) ListByInvitee(ctx context.Context, inviteeUserID string) ([]*domain.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadInvitations(ctx)
	if err != nil {
		return nil, fmt.Errorf("招待データ読み込みに失敗しました: %w", err)
	}

	invitations := make([]*domain.Invitation, 0)
	for _, rec := range records {
		if rec.InviteeUserID == inviteeUserID {
			invitations = append(invitations, toInvitation(rec))
		}
	}
	return invitations, nil
}

// ListByTeam はチームIDで招待一覧を返します。
func (r *GCSInvitationRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadInvitations(ctx)
	if err != nil {
		return nil, fmt.Errorf("招待データ読み込みに失敗しました: %w", err)
	}

	invitations := make([]*domain.Invitation, 0)
	for _, rec := range records {
		if rec.TeamID == teamID {
			invitations = append(invitations, toInvitation(rec))
		}
	}
	return invitations, nil
}

// FindPendingByTeamAndInvitee は特定チーム・招待先ユーザーの pending 招待を返します。
func (r *GCSInvitationRepository) FindPendingByTeamAndInvitee(ctx context.Context, teamID, inviteeUserID string) (*domain.Invitation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadInvitations(ctx)
	if err != nil {
		return nil, fmt.Errorf("招待データ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.TeamID == teamID && rec.InviteeUserID == inviteeUserID && rec.Status == string(domain.StatusPending) {
			return toInvitation(rec), nil
		}
	}
	return nil, domain.ErrInvitationNotFound
}

// GCSInvitationRepository が domain.InvitationRepository を実装していることをコンパイル時に保証します。
var _ domain.InvitationRepository = (*GCSInvitationRepository)(nil)
