package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
)

// TeamUseCase はチーム管理に関するユースケースを実装します。
type TeamUseCase struct {
	teamRepo domain.TeamRepository
	userRepo domain.UserRepository
}

// NewTeamUseCase は TeamUseCase を生成します（コンストラクタインジェクション）。
func NewTeamUseCase(teamRepo domain.TeamRepository, userRepo domain.UserRepository) *TeamUseCase {
	return &TeamUseCase{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// CreateTeamInput はチーム作成ユースケースの入力です。
type CreateTeamInput struct {
	Name        string
	Description string
	CallerID    string
	CallerRole  domain.Role
}

// CreateTeam は新しいチームを作成します（admin / teamowner のみ呼び出し可）。
// - チーム名が重複している場合は ErrTeamNameAlreadyExists を返します。
// - 呼び出し元が admin または teamowner でない場合は ErrPermissionDenied を返します。
func (uc *TeamUseCase) CreateTeam(ctx context.Context, input CreateTeamInput) (*domain.Team, error) {
	if input.CallerRole != domain.RoleAdmin && input.CallerRole != domain.RoleTeamOwner {
		return nil, domain.ErrPermissionDenied
	}

	if input.Name == "" {
		return nil, fmt.Errorf("チーム名は必須です")
	}

	if _, err := uc.teamRepo.FindByName(ctx, input.Name); err == nil {
		return nil, domain.ErrTeamNameAlreadyExists
	}

	now := time.Now().UTC()
	team := &domain.Team{
		ID:          uuid.NewString(),
		Name:        input.Name,
		Description: input.Description,
		OwnerID:     input.CallerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.teamRepo.Save(ctx, team); err != nil {
		return nil, fmt.Errorf("チームの保存に失敗しました: %w", err)
	}

	return team, nil
}

// ListTeamsInput はチーム一覧ユースケースの入力です。
type ListTeamsInput struct {
	CallerID   string
	CallerRole domain.Role
}

// ListTeams はチーム一覧を取得します。
// - admin は全チームを返します。
// - それ以外はオーナーまたはメンバーであるチームのみを返します。
func (uc *TeamUseCase) ListTeams(ctx context.Context, input ListTeamsInput) ([]*domain.Team, error) {
	if input.CallerRole == domain.RoleAdmin {
		teams, err := uc.teamRepo.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("チーム一覧取得に失敗しました: %w", err)
		}
		return teams, nil
	}

	teams, err := uc.teamRepo.ListByOwnerOrMember(ctx, input.CallerID)
	if err != nil {
		return nil, fmt.Errorf("チーム一覧取得に失敗しました: %w", err)
	}
	return teams, nil
}

// TeamDetailOutput はチーム詳細ユースケースの出力です。
type TeamDetailOutput struct {
	Team    *domain.Team
	Members []*domain.TeamMember
}

// GetTeam は指定IDのチーム詳細（メンバー一覧を含む）を取得します。
// - admin またはチームメンバー（オーナーを含む）のみアクセス可能です。
// - 権限がない場合は ErrPermissionDenied を返します。
func (uc *TeamUseCase) GetTeam(ctx context.Context, callerID string, callerRole domain.Role, teamID string) (*TeamDetailOutput, error) {
	team, err := uc.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("チーム取得に失敗しました: %w", err)
	}

	if callerRole != domain.RoleAdmin {
		isMember, err := uc.teamRepo.IsMember(ctx, teamID, callerID)
		if err != nil {
			return nil, fmt.Errorf("メンバー確認に失敗しました: %w", err)
		}
		if !isMember && team.OwnerID != callerID {
			return nil, domain.ErrPermissionDenied
		}
	}

	members, err := uc.teamRepo.ListMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("メンバー一覧取得に失敗しました: %w", err)
	}

	return &TeamDetailOutput{
		Team:    team,
		Members: members,
	}, nil
}

// UpdateTeamInput はチーム更新ユースケースの入力です。
// 各フィールドは nil ポインタの場合は更新しません。
type UpdateTeamInput struct {
	Name        *string
	Description *string
}

// UpdateTeam は指定IDのチーム情報を更新します。
// - owner_id 本人または admin のみ実行可能です。
// - チーム名が重複している場合は ErrTeamNameAlreadyExists を返します。
func (uc *TeamUseCase) UpdateTeam(ctx context.Context, callerID string, callerRole domain.Role, teamID string, input UpdateTeamInput) (*domain.Team, error) {
	team, err := uc.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("チーム取得に失敗しました: %w", err)
	}

	if callerRole != domain.RoleAdmin && team.OwnerID != callerID {
		return nil, domain.ErrPermissionDenied
	}

	if input.Name != nil {
		existing, err := uc.teamRepo.FindByName(ctx, *input.Name)
		if err == nil && existing.ID != teamID {
			return nil, domain.ErrTeamNameAlreadyExists
		}
		team.Name = *input.Name
	}

	if input.Description != nil {
		team.Description = *input.Description
	}

	team.UpdatedAt = time.Now().UTC()

	if err := uc.teamRepo.Save(ctx, team); err != nil {
		return nil, fmt.Errorf("チームの保存に失敗しました: %w", err)
	}

	return team, nil
}

// DeleteTeam は指定IDのチームを削除します。
// - owner_id 本人または admin のみ実行可能です。
func (uc *TeamUseCase) DeleteTeam(ctx context.Context, callerID string, callerRole domain.Role, teamID string) error {
	team, err := uc.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("チーム取得に失敗しました: %w", err)
	}

	if callerRole != domain.RoleAdmin && team.OwnerID != callerID {
		return domain.ErrPermissionDenied
	}

	if err := uc.teamRepo.Delete(ctx, teamID); err != nil {
		return fmt.Errorf("チームの削除に失敗しました: %w", err)
	}

	return nil
}

// AddMember は指定チームにユーザーを追加します。
// - owner_id 本人または admin のみ実行可能です。
// - 追加するユーザーが存在しない場合は ErrUserNotFound を返します。
// - 既にメンバーの場合は ErrMemberAlreadyExists を返します。
func (uc *TeamUseCase) AddMember(ctx context.Context, callerID string, callerRole domain.Role, teamID, userID string) (*domain.TeamMember, error) {
	team, err := uc.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("チーム取得に失敗しました: %w", err)
	}

	if callerRole != domain.RoleAdmin && team.OwnerID != callerID {
		return nil, domain.ErrPermissionDenied
	}

	if _, err := uc.userRepo.FindByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("ユーザー取得に失敗しました: %w", err)
	}

	member := &domain.TeamMember{
		TeamID:   teamID,
		UserID:   userID,
		JoinedAt: time.Now().UTC(),
	}

	if err := uc.teamRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("メンバー追加に失敗しました: %w", err)
	}

	return member, nil
}

// RemoveMember は指定チームからユーザーを除外します。
// - owner_id 本人または admin のみ実行可能です。
func (uc *TeamUseCase) RemoveMember(ctx context.Context, callerID string, callerRole domain.Role, teamID, userID string) error {
	team, err := uc.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("チーム取得に失敗しました: %w", err)
	}

	if callerRole != domain.RoleAdmin && team.OwnerID != callerID {
		return domain.ErrPermissionDenied
	}

	if err := uc.teamRepo.RemoveMember(ctx, teamID, userID); err != nil {
		return fmt.Errorf("メンバー除外に失敗しました: %w", err)
	}

	return nil
}
