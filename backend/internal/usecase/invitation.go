package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
)

// InvitationUseCase は招待管理に関するユースケースを実装します。
type InvitationUseCase struct {
	invitationRepo domain.InvitationRepository
	teamRepo       domain.TeamRepository
	userRepo       domain.UserRepository
}

// NewInvitationUseCase は InvitationUseCase を生成します（コンストラクタインジェクション）。
func NewInvitationUseCase(
	invitationRepo domain.InvitationRepository,
	teamRepo domain.TeamRepository,
	userRepo domain.UserRepository,
) *InvitationUseCase {
	return &InvitationUseCase{
		invitationRepo: invitationRepo,
		teamRepo:       teamRepo,
		userRepo:       userRepo,
	}
}

// SendInvitationInput は招待送信ユースケースの入力です。
type SendInvitationInput struct {
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのグローバルロール
	CallerRole domain.Role
	// TeamID は招待先チームのID
	TeamID string
	// InviteeIdentifier は招待するユーザーの識別子（UUID / ユーザー名 / メールアドレス）
	InviteeIdentifier string
}

// resolveUser は inviteeIdentifier からユーザーを解決します。
// UUID 形式なら FindByID → FindByUsername → FindByEmail の順で試みます。
func (uc *InvitationUseCase) resolveUser(ctx context.Context, identifier string) (*domain.User, error) {
	// UUID 形式ならまず FindByID を試みる
	if _, err := uuid.Parse(identifier); err == nil {
		user, err := uc.userRepo.FindByID(ctx, identifier)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, domain.ErrUserNotFound) {
			return nil, fmt.Errorf("ユーザーID検索に失敗しました: %w", err)
		}
	}

	// ユーザー名で検索
	user, err := uc.userRepo.FindByUsername(ctx, identifier)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("ユーザー名検索に失敗しました: %w", err)
	}

	// メールアドレスで検索
	user, err = uc.userRepo.FindByEmail(ctx, identifier)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("メールアドレス検索に失敗しました: %w", err)
	}

	return nil, domain.ErrUserNotFound
}

// SendInvitation はチームへの招待を送信します。
// - チームの per-team owner または admin のみ実行可能です。
// - invitee_identifier からユーザーを解決し、解決できなければ ErrUserNotFound を返します。
// - 招待先ユーザーが既にメンバーの場合は ErrMemberAlreadyExists を返します。
// - 同一チームへの pending 招待が存在する場合は ErrInvitationAlreadyExists を返します。
func (uc *InvitationUseCase) SendInvitation(ctx context.Context, input SendInvitationInput) (*domain.Invitation, error) {
	if input.InviteeIdentifier == "" {
		return nil, fmt.Errorf("invitee_identifier は必須です")
	}

	// チームの存在確認
	if _, err := uc.teamRepo.FindByID(ctx, input.TeamID); err != nil {
		return nil, fmt.Errorf("チーム取得に失敗しました: %w", err)
	}

	// 権限確認: admin か、per-team owner か
	if input.CallerRole != domain.RoleAdmin {
		owners, err := uc.teamRepo.FindOwners(ctx, input.TeamID)
		if err != nil {
			return nil, fmt.Errorf("チームオーナー取得に失敗しました: %w", err)
		}
		isOwner := false
		for _, o := range owners {
			if o.UserID == input.CallerID {
				isOwner = true
				break
			}
		}
		if !isOwner {
			return nil, domain.ErrPermissionDenied
		}
	}

	// invitee_identifier からユーザーを解決
	invitee, err := uc.resolveUser(ctx, input.InviteeIdentifier)
	if err != nil {
		return nil, err
	}

	// 招待先ユーザーが既にメンバーかチェック
	isMember, err := uc.teamRepo.IsMember(ctx, input.TeamID, invitee.ID)
	if err != nil {
		return nil, fmt.Errorf("メンバー確認に失敗しました: %w", err)
	}
	if isMember {
		return nil, domain.ErrMemberAlreadyExists
	}

	// 同一チームへの pending 招待が既に存在するかチェック
	_, err = uc.invitationRepo.FindPendingByTeamAndInvitee(ctx, input.TeamID, invitee.ID)
	if err == nil {
		return nil, domain.ErrInvitationAlreadyExists
	}
	if !errors.Is(err, domain.ErrInvitationNotFound) {
		return nil, fmt.Errorf("招待確認に失敗しました: %w", err)
	}

	inv := &domain.Invitation{
		ID:                uuid.NewString(),
		TeamID:            input.TeamID,
		InvitedBy:         input.CallerID,
		InviteeIdentifier: input.InviteeIdentifier,
		InviteeUserID:     invitee.ID,
		Status:            domain.StatusPending,
		CreatedAt:         time.Now().UTC(),
	}

	if err := uc.invitationRepo.Save(ctx, inv); err != nil {
		return nil, fmt.Errorf("招待の保存に失敗しました: %w", err)
	}

	return inv, nil
}

// ListMyInvitationsInput は自分宛招待一覧ユースケースの入力です。
type ListMyInvitationsInput struct {
	// CallerID はログインユーザーのID
	CallerID string
}

// ListMyInvitations はログインユーザー宛の招待一覧を返します。
func (uc *InvitationUseCase) ListMyInvitations(ctx context.Context, input ListMyInvitationsInput) ([]*domain.Invitation, error) {
	invitations, err := uc.invitationRepo.ListByInvitee(ctx, input.CallerID)
	if err != nil {
		return nil, fmt.Errorf("招待一覧取得に失敗しました: %w", err)
	}
	return invitations, nil
}

// RespondInvitationInput は招待受諾/拒否ユースケースの入力です。
type RespondInvitationInput struct {
	// CallerID はログインユーザーのID
	CallerID string
	// InvitationID は対象招待のID
	InvitationID string
	// Status は応答ステータス（accepted または rejected）
	Status domain.InvitationStatus
}

// RespondInvitation は招待を受諾または拒否します。
// - 招待が存在しなければ ErrInvitationNotFound を返します。
// - CallerID が invitee_user_id と一致しなければ ErrPermissionDenied を返します。
// - status が pending でなければ ErrInvitationNotPending を返します。
// - accepted の場合: TeamMember に追加（既にメンバーなら ErrMemberAlreadyExists）→ status を更新。
// - rejected の場合: status を rejected に更新のみ。
func (uc *InvitationUseCase) RespondInvitation(ctx context.Context, input RespondInvitationInput) (*domain.Invitation, error) {
	inv, err := uc.invitationRepo.FindByID(ctx, input.InvitationID)
	if err != nil {
		return nil, fmt.Errorf("招待取得に失敗しました: %w", err)
	}

	// 招待先ユーザーのみが応答可能
	if inv.InviteeUserID != input.CallerID {
		return nil, domain.ErrPermissionDenied
	}

	// pending 状態のみ応答可能
	if inv.Status != domain.StatusPending {
		return nil, domain.ErrInvitationNotPending
	}

	if input.Status == domain.StatusAccepted {
		// チームにメンバーとして追加
		member := &domain.TeamMember{
			TeamID:   inv.TeamID,
			UserID:   input.CallerID,
			Role:     domain.MemberRoleMember,
			JoinedAt: time.Now().UTC(),
		}
		if err := uc.teamRepo.AddMember(ctx, member); err != nil {
			return nil, fmt.Errorf("メンバー追加に失敗しました: %w", err)
		}
	}

	// ステータスを更新
	inv.Status = input.Status
	if err := uc.invitationRepo.Save(ctx, inv); err != nil {
		return nil, fmt.Errorf("招待の更新に失敗しました: %w", err)
	}

	return inv, nil
}
