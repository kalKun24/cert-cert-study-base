package firestore

import (
	"context"
	"fmt"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// invitationRecord はFirestoreに保存する招待レコードです。
// コレクション: invitations/{invitationID}
type invitationRecord struct {
	ID                string    `firestore:"id"`
	TeamID            string    `firestore:"team_id"`
	InvitedBy         string    `firestore:"invited_by"`
	InviteeIdentifier string    `firestore:"invitee_identifier"`
	InviteeUserID     string    `firestore:"invitee_user_id"`
	Status            string    `firestore:"status"`
	CreatedAt         time.Time `firestore:"created_at"`
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

// FirestoreInvitationRepository はCloud Firestoreに招待データを永続化するリポジトリです。
// domain.InvitationRepository インターフェースを実装します。
type FirestoreInvitationRepository struct {
	client *fs.Client
}

// NewFirestoreInvitationRepository は FirestoreInvitationRepository を生成します。
func NewFirestoreInvitationRepository(client *fs.Client) *FirestoreInvitationRepository {
	return &FirestoreInvitationRepository{client: client}
}

func (r *FirestoreInvitationRepository) invitationsCol() *fs.CollectionRef {
	return r.client.Collection("invitations")
}

// Save は招待を新規作成または更新します（Upsert）。
func (r *FirestoreInvitationRepository) Save(ctx context.Context, inv *domain.Invitation) error {
	rec := toInvitationRecord(inv)
	_, err := r.invitationsCol().Doc(inv.ID).Set(ctx, rec)
	if err != nil {
		return fmt.Errorf("招待の保存に失敗しました: %w", err)
	}
	return nil
}

// FindByID はIDで招待を検索します。
func (r *FirestoreInvitationRepository) FindByID(ctx context.Context, id string) (*domain.Invitation, error) {
	doc, err := r.invitationsCol().Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrInvitationNotFound
		}
		return nil, fmt.Errorf("招待の取得に失敗しました: %w", err)
	}

	var rec invitationRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("招待データのデコードに失敗しました: %w", err)
	}
	return toInvitation(rec), nil
}

// ListByInvitee は招待先ユーザーIDで招待一覧を返します。
func (r *FirestoreInvitationRepository) ListByInvitee(ctx context.Context, inviteeUserID string) ([]*domain.Invitation, error) {
	iter := r.invitationsCol().Where("invitee_user_id", "==", inviteeUserID).Documents(ctx)
	defer iter.Stop()

	return r.collectInvitations(ctx, iter)
}

// ListByTeam はチームIDで招待一覧を返します。
func (r *FirestoreInvitationRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.Invitation, error) {
	iter := r.invitationsCol().Where("team_id", "==", teamID).Documents(ctx)
	defer iter.Stop()

	return r.collectInvitations(ctx, iter)
}

// FindPendingByTeamAndInvitee は特定チーム・招待先ユーザーの pending 招待を返します。
func (r *FirestoreInvitationRepository) FindPendingByTeamAndInvitee(ctx context.Context, teamID, inviteeUserID string) (*domain.Invitation, error) {
	iter := r.invitationsCol().
		Where("team_id", "==", teamID).
		Where("invitee_user_id", "==", inviteeUserID).
		Where("status", "==", string(domain.StatusPending)).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, domain.ErrInvitationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("招待の検索に失敗しました: %w", err)
	}

	var rec invitationRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("招待データのデコードに失敗しました: %w", err)
	}
	return toInvitation(rec), nil
}

func (r *FirestoreInvitationRepository) collectInvitations(_ context.Context, iter *fs.DocumentIterator) ([]*domain.Invitation, error) {
	var invitations []*domain.Invitation
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("招待一覧の取得に失敗しました: %w", err)
		}

		var rec invitationRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("招待データのデコードに失敗しました: %w", err)
		}
		invitations = append(invitations, toInvitation(rec))
	}

	if invitations == nil {
		invitations = []*domain.Invitation{}
	}
	return invitations, nil
}

// FirestoreInvitationRepository が domain.InvitationRepository を実装していることをコンパイル時に保証します。
var _ domain.InvitationRepository = (*FirestoreInvitationRepository)(nil)
