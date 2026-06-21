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
	teamRepo     domain.TeamRepository
	userRepo     domain.UserRepository
	questionRepo domain.QuestionRepository
	commentRepo  domain.CommentRepository
}

// NewTeamUseCase は TeamUseCase を生成します（コンストラクタインジェクション）。
// questionRepo / commentRepo はメンバー統計機能（ListMemberStats）で使用します。
func NewTeamUseCase(
	teamRepo domain.TeamRepository,
	userRepo domain.UserRepository,
	questionRepo domain.QuestionRepository,
	commentRepo domain.CommentRepository,
) *TeamUseCase {
	return &TeamUseCase{
		teamRepo:     teamRepo,
		userRepo:     userRepo,
		questionRepo: questionRepo,
		commentRepo:  commentRepo,
	}
}

// CreateTeamInput はチーム作成ユースケースの入力です。
type CreateTeamInput struct {
	Name        string
	Description string
	CallerID    string
	CallerRole  domain.Role
}

// CreateTeam は新しいチームを作成します（admin / IsTeamOwner=true のユーザーのみ呼び出し可）。
// - チーム名が重複している場合は ErrTeamNameAlreadyExists を返します。
// - 呼び出し元が admin でなく IsTeamOwner=false の場合は ErrPermissionDenied を返します。
// - MaxTeams が設定されており上限に達している場合は ErrPermissionDenied を返します。
// - チーム作成に成功すると、作成者が per-team owner として自動的に追加されます。
func (uc *TeamUseCase) CreateTeam(ctx context.Context, input CreateTeamInput) (*domain.Team, error) {
	// admin は常に許可。それ以外は IsTeamOwner フラグで判定
	if input.CallerRole != domain.RoleAdmin {
		caller, err := uc.userRepo.FindByID(ctx, input.CallerID)
		if err != nil {
			return nil, fmt.Errorf("呼び出し元ユーザー取得に失敗しました: %w", err)
		}
		if !caller.IsTeamOwner {
			return nil, domain.ErrPermissionDenied
		}

		// MaxTeams 上限チェック（0 は制限なし）
		// MaxTeams は「per-team owner ロールを持つチームの数」を制限する。
		// Team.OwnerID（作成者フィールド）は後方互換のために残しているが、
		// カウントの基準は TeamMember.Role == "owner" のみとする。
		if caller.MaxTeams > 0 {
			ownedTeams, err := uc.teamRepo.ListByOwnerOrMember(ctx, input.CallerID)
			if err != nil {
				return nil, fmt.Errorf("チーム一覧取得に失敗しました: %w", err)
			}
			ownerCount := 0
			for _, t := range ownedTeams {
				owners, err := uc.teamRepo.FindOwners(ctx, t.ID)
				if err != nil {
					return nil, fmt.Errorf("チームオーナー取得に失敗しました: %w", err)
				}
				for _, o := range owners {
					if o.UserID == input.CallerID {
						ownerCount++
						break
					}
				}
			}
			if ownerCount >= caller.MaxTeams {
				return nil, domain.ErrPermissionDenied
			}
		}
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

	// 作成者を per-team owner として自動追加
	creatorMember := &domain.TeamMember{
		TeamID:   team.ID,
		UserID:   input.CallerID,
		Role:     domain.MemberRoleOwner,
		JoinedAt: now,
	}
	if err := uc.teamRepo.AddMember(ctx, creatorMember); err != nil {
		return nil, fmt.Errorf("作成者のメンバー追加に失敗しました: %w", err)
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
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now().UTC(),
	}

	if err := uc.teamRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("メンバー追加に失敗しました: %w", err)
	}

	return member, nil
}

// ChangeMemberRoleInput はチームメンバーロール変更ユースケースの入力です。
type ChangeMemberRoleInput struct {
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのグローバルロール
	CallerRole domain.Role
	// TeamID は対象チームのID
	TeamID string
	// TargetUserID はロールを変更するメンバーのユーザーID
	TargetUserID string
	// Role は変更後のロール
	Role domain.MemberRole
}

// ChangeMemberRole はチームメンバーのロールを変更します（チームの per-team owner または admin のみ）。
// - ロールが無効な値の場合はエラーを返します。
// - 対象ユーザーがチームのメンバーでない場合は ErrMemberNotFound を返します。
// - 呼び出し元が admin でなく per-team owner でもない場合は ErrPermissionDenied を返します。
// - チームの最後のオーナーを member に降格する場合は ErrLastTeamOwner を返します。
func (uc *TeamUseCase) ChangeMemberRole(ctx context.Context, input ChangeMemberRoleInput) (*domain.TeamMember, error) {
	if !input.Role.IsValid() {
		return nil, fmt.Errorf("無効なロールです: %s", input.Role)
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

	// 対象ユーザーがメンバーか確認
	isMember, err := uc.teamRepo.IsMember(ctx, input.TeamID, input.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("メンバー確認に失敗しました: %w", err)
	}
	if !isMember {
		return nil, domain.ErrMemberNotFound
	}

	// member への降格時：最後のオーナーになるケースを防止
	if input.Role == domain.MemberRoleMember {
		owners, err := uc.teamRepo.FindOwners(ctx, input.TeamID)
		if err != nil {
			return nil, fmt.Errorf("チームオーナー取得に失敗しました: %w", err)
		}
		if len(owners) <= 1 {
			// 対象が owner かどうかを確認
			for _, o := range owners {
				if o.UserID == input.TargetUserID {
					return nil, domain.ErrLastTeamOwner
				}
			}
		}
	}

	// ロールを更新
	if err := uc.teamRepo.UpdateMemberRole(ctx, input.TeamID, input.TargetUserID, input.Role); err != nil {
		return nil, fmt.Errorf("ロール更新に失敗しました: %w", err)
	}

	// 更新後のメンバー情報を返す
	members, err := uc.teamRepo.ListMembers(ctx, input.TeamID)
	if err != nil {
		return nil, fmt.Errorf("メンバー一覧取得に失敗しました: %w", err)
	}
	for _, m := range members {
		if m.UserID == input.TargetUserID {
			return m, nil
		}
	}

	return nil, domain.ErrMemberNotFound
}

// LeaveTeam はログインユーザーが自分自身をチームから脱退させます。
// - チームが存在しなければ ErrTeamNotFound を返します。
// - 自分がメンバーでなければ ErrMemberNotFound を返します。
// - 自分が唯一の per-team owner なら ErrLastTeamOwner を返します。
func (uc *TeamUseCase) LeaveTeam(ctx context.Context, callerID string, teamID string) error {
	// チームの存在確認
	if _, err := uc.teamRepo.FindByID(ctx, teamID); err != nil {
		return fmt.Errorf("チーム取得に失敗しました: %w", err)
	}

	// 自分がメンバーかどうか確認
	isMember, err := uc.teamRepo.IsMember(ctx, teamID, callerID)
	if err != nil {
		return fmt.Errorf("メンバー確認に失敗しました: %w", err)
	}
	if !isMember {
		return domain.ErrMemberNotFound
	}

	// 唯一の per-team owner による脱退は禁止
	owners, err := uc.teamRepo.FindOwners(ctx, teamID)
	if err != nil {
		return fmt.Errorf("チームオーナー取得に失敗しました: %w", err)
	}
	if len(owners) <= 1 {
		for _, o := range owners {
			if o.UserID == callerID {
				return domain.ErrLastTeamOwner
			}
		}
	}

	if err := uc.teamRepo.RemoveMember(ctx, teamID, callerID); err != nil {
		return fmt.Errorf("チームからの脱退に失敗しました: %w", err)
	}

	return nil
}

// RemoveMember は指定チームからユーザーを除外します。
// - per-team owner または admin のみ実行可能です。
// - 除外対象が唯一の per-team owner の場合は ErrLastTeamOwner を返します。
func (uc *TeamUseCase) RemoveMember(ctx context.Context, callerID string, callerRole domain.Role, teamID, userID string) error {
	team, err := uc.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return fmt.Errorf("チーム取得に失敗しました: %w", err)
	}

	// 権限チェック: admin か per-team owner か
	if callerRole != domain.RoleAdmin {
		owners, err := uc.teamRepo.FindOwners(ctx, teamID)
		if err != nil {
			return fmt.Errorf("チームオーナー取得に失敗しました: %w", err)
		}
		isOwner := false
		for _, o := range owners {
			if o.UserID == callerID {
				isOwner = true
				break
			}
		}
		// 後方互換: Team.OwnerID でも許可（将来削除予定）
		if !isOwner && team.OwnerID != callerID {
			return domain.ErrPermissionDenied
		}
	}

	// 除外対象が唯一の per-team owner の場合は拒否
	owners, err := uc.teamRepo.FindOwners(ctx, teamID)
	if err != nil {
		return fmt.Errorf("チームオーナー取得に失敗しました: %w", err)
	}
	if len(owners) <= 1 {
		for _, o := range owners {
			if o.UserID == userID {
				return domain.ErrLastTeamOwner
			}
		}
	}

	if err := uc.teamRepo.RemoveMember(ctx, teamID, userID); err != nil {
		return fmt.Errorf("メンバー除外に失敗しました: %w", err)
	}

	return nil
}

// MemberStats はチームメンバー1人分の統計情報です。
type MemberStats struct {
	// UserID はメンバーのユーザーID
	UserID string
	// DisplayName は表示名
	DisplayName string
	// Role はチーム内ロール（owner / member）
	Role domain.MemberRole
	// IsTeamOwner はグローバルチームオーナー権限を持つか
	IsTeamOwner bool
	// QuestionCount はこのチーム内で作成した問題数
	QuestionCount int
	// CommentCount はこのチーム内の問題に投稿したコメント数
	CommentCount int
	// LastLoginAt は最終ログイン日時（一度もログインしていない場合は nil）
	LastLoginAt *time.Time
}

// ListMemberStats はチームメンバーごとの統計情報（問題数・コメント数・最終ログイン日時）を返します。
// - admin またはチームメンバーのみアクセス可能です。
// - questionRepo / commentRepo が設定されていない場合は問題数・コメント数は 0 を返します。
func (uc *TeamUseCase) ListMemberStats(ctx context.Context, callerID string, callerRole domain.Role, teamID string) ([]*MemberStats, error) {
	// チームの存在確認
	if _, err := uc.teamRepo.FindByID(ctx, teamID); err != nil {
		return nil, fmt.Errorf("チーム取得に失敗しました: %w", err)
	}

	// 権限確認: admin またはチームメンバーのみアクセス可能
	if callerRole != domain.RoleAdmin {
		isMember, err := uc.teamRepo.IsMember(ctx, teamID, callerID)
		if err != nil {
			return nil, fmt.Errorf("メンバー確認に失敗しました: %w", err)
		}
		if !isMember {
			return nil, domain.ErrPermissionDenied
		}
	}

	// メンバー一覧を取得
	members, err := uc.teamRepo.ListMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("メンバー一覧取得に失敗しました: %w", err)
	}

	// 問題一覧を一度取得し、問題数・コメント数の両方の集計に再利用する
	// （questionRepo が未設定の場合は集計をスキップし 0 のままとする）
	questionCountByUser := make(map[string]int)
	commentCountByUser := make(map[string]int)
	if uc.questionRepo != nil {
		questions, err := uc.questionRepo.ListByTeam(ctx, teamID)
		if err != nil {
			return nil, fmt.Errorf("問題一覧取得に失敗しました: %w", err)
		}
		for _, q := range questions {
			questionCountByUser[q.CreatedBy]++
		}

		// コメント数の集計（commentRepo が設定されている場合のみ）
		// 既に取得した questions スライスを再利用することで二重取得を防ぐ
		if uc.commentRepo != nil {
			for _, q := range questions {
				comments, err := uc.commentRepo.ListByQuestionID(ctx, q.TeamID, q.ID)
				if err != nil {
					return nil, fmt.Errorf("コメント一覧取得に失敗しました（question_id=%s）: %w", q.ID, err)
				}
				for _, c := range comments {
					commentCountByUser[c.CreatedBy]++
				}
			}
		}
	}

	// 全ユーザーを一括取得し、ユーザーIDをキーにしたマップを作成する。
	// これにより、メンバーごとの個別 GCS アクセス（N+1）を O(1) のマップ参照に置き換える。
	allUsers, err := uc.userRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("ユーザー一覧取得に失敗しました: %w", err)
	}
	userByID := make(map[string]*domain.User, len(allUsers))
	for _, u := range allUsers {
		userByID[u.ID] = u
	}

	// メンバーごとにユーザー情報をマップから取得して統計を組み立て
	stats := make([]*MemberStats, 0, len(members))
	for _, m := range members {
		user, ok := userByID[m.UserID]
		if !ok {
			return nil, fmt.Errorf("データ整合性エラー: チームメンバーに対応するユーザーが存在しません（user_id=%s）", m.UserID)
		}

		stats = append(stats, &MemberStats{
			UserID:        m.UserID,
			DisplayName:   user.DisplayName,
			Role:          m.Role,
			IsTeamOwner:   user.IsTeamOwner,
			QuestionCount: questionCountByUser[m.UserID],
			CommentCount:  commentCountByUser[m.UserID],
			LastLoginAt:   user.LastLoginAt,
		})
	}

	return stats, nil
}
