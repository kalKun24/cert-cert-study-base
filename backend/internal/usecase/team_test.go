package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// --- チームリポジトリモック ---

type mockTeamRepository struct {
	teams   map[string]*domain.Team
	byName  map[string]*domain.Team
	members []domain.TeamMember
	saveErr error
}

func newMockTeamRepository() *mockTeamRepository {
	return &mockTeamRepository{
		teams:   make(map[string]*domain.Team),
		byName:  make(map[string]*domain.Team),
		members: []domain.TeamMember{},
	}
}

func (m *mockTeamRepository) addTeam(t *domain.Team) {
	m.teams[t.ID] = t
	m.byName[t.Name] = t
}

func (m *mockTeamRepository) FindByID(_ context.Context, id string) (*domain.Team, error) {
	if t, ok := m.teams[id]; ok {
		return t, nil
	}
	return nil, domain.ErrTeamNotFound
}

func (m *mockTeamRepository) FindByName(_ context.Context, name string) (*domain.Team, error) {
	if t, ok := m.byName[name]; ok {
		return t, nil
	}
	return nil, domain.ErrTeamNotFound
}

func (m *mockTeamRepository) List(_ context.Context) ([]*domain.Team, error) {
	teams := make([]*domain.Team, 0, len(m.teams))
	for _, t := range m.teams {
		teams = append(teams, t)
	}
	return teams, nil
}

func (m *mockTeamRepository) ListByOwnerOrMember(_ context.Context, userID string) ([]*domain.Team, error) {
	memberTeamIDs := make(map[string]struct{})
	for _, mem := range m.members {
		if mem.UserID == userID {
			memberTeamIDs[mem.TeamID] = struct{}{}
		}
	}

	teams := make([]*domain.Team, 0)
	for _, t := range m.teams {
		_, isMember := memberTeamIDs[t.ID]
		if t.OwnerID == userID || isMember {
			teams = append(teams, t)
		}
	}
	return teams, nil
}

func (m *mockTeamRepository) Save(_ context.Context, team *domain.Team) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.teams[team.ID] = team
	m.byName[team.Name] = team
	return nil
}

func (m *mockTeamRepository) Delete(_ context.Context, id string) error {
	t, ok := m.teams[id]
	if !ok {
		return domain.ErrTeamNotFound
	}
	delete(m.byName, t.Name)
	delete(m.teams, id)

	newMembers := make([]domain.TeamMember, 0, len(m.members))
	for _, mem := range m.members {
		if mem.TeamID != id {
			newMembers = append(newMembers, mem)
		}
	}
	m.members = newMembers
	return nil
}

func (m *mockTeamRepository) AddMember(_ context.Context, member *domain.TeamMember) error {
	for _, existing := range m.members {
		if existing.TeamID == member.TeamID && existing.UserID == member.UserID {
			return domain.ErrMemberAlreadyExists
		}
	}
	m.members = append(m.members, *member)
	return nil
}

func (m *mockTeamRepository) RemoveMember(_ context.Context, teamID, userID string) error {
	newMembers := make([]domain.TeamMember, 0, len(m.members))
	found := false
	for _, mem := range m.members {
		if mem.TeamID == teamID && mem.UserID == userID {
			found = true
			continue
		}
		newMembers = append(newMembers, mem)
	}
	if !found {
		return domain.ErrMemberNotFound
	}
	m.members = newMembers
	return nil
}

func (m *mockTeamRepository) ListMembers(_ context.Context, teamID string) ([]*domain.TeamMember, error) {
	result := make([]*domain.TeamMember, 0)
	for _, mem := range m.members {
		if mem.TeamID == teamID {
			copy := mem
			result = append(result, &copy)
		}
	}
	return result, nil
}

func (m *mockTeamRepository) IsMember(_ context.Context, teamID, userID string) (bool, error) {
	for _, mem := range m.members {
		if mem.TeamID == teamID && mem.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

// FindOwners はチームのオーナーロールを持つメンバー一覧を返します。
func (m *mockTeamRepository) FindOwners(_ context.Context, teamID string) ([]*domain.TeamMember, error) {
	owners := make([]*domain.TeamMember, 0)
	for _, mem := range m.members {
		if mem.TeamID == teamID && mem.Role == domain.MemberRoleOwner {
			copyMem := mem
			owners = append(owners, &copyMem)
		}
	}
	return owners, nil
}

// UpdateMemberRole はチームメンバーのロールを変更します。
func (m *mockTeamRepository) UpdateMemberRole(_ context.Context, teamID, userID string, role domain.MemberRole) error {
	for i, mem := range m.members {
		if mem.TeamID == teamID && mem.UserID == userID {
			m.members[i].Role = role
			return nil
		}
	}
	return domain.ErrMemberNotFound
}

// --- テストヘルパー ---

func testTeam(id, name, ownerID string) *domain.Team {
	return &domain.Team{
		ID:          id,
		Name:        name,
		Description: "テストチーム",
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// --- TeamUseCase のテスト ---

func TestTeamUseCase_CreateTeam_Success_Admin(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	team, err := uc.CreateTeam(context.Background(), usecase.CreateTeamInput{
		Name:       "CISSP勉強会",
		CallerID:   "admin-id",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Fatalf("チーム作成に失敗しました: %v", err)
	}
	if team.Name != "CISSP勉強会" {
		t.Errorf("チーム名が期待値と異なります: got %s, want CISSP勉強会", team.Name)
	}
	if team.OwnerID != "admin-id" {
		t.Errorf("OwnerIDが期待値と異なります: got %s, want admin-id", team.OwnerID)
	}
}

func TestTeamUseCase_CreateTeam_Success_IsTeamOwner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	// IsTeamOwner=true のユーザーを追加
	ownerUser := testUser("owner-id", "teamowner", "owner@example.com", domain.RoleUser, true)
	ownerUser.IsTeamOwner = true
	ownerUser.MaxTeams = 0 // 制限なし
	userRepo.addUser(ownerUser)
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	team, err := uc.CreateTeam(context.Background(), usecase.CreateTeamInput{
		Name:       "SC勉強会",
		CallerID:   "owner-id",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("チーム作成に失敗しました: %v", err)
	}
	if team.OwnerID != "owner-id" {
		t.Errorf("OwnerIDが期待値と異なります: got %s, want owner-id", team.OwnerID)
	}
	// 作成者が per-team owner として追加されていること
	owners, err := teamRepo.FindOwners(context.Background(), team.ID)
	if err != nil {
		t.Fatalf("オーナー取得に失敗しました: %v", err)
	}
	if len(owners) != 1 || owners[0].UserID != "owner-id" {
		t.Errorf("作成者が per-team owner として追加されていません")
	}
}

func TestTeamUseCase_CreateTeam_PermissionDenied_IsTeamOwnerFalse(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	// IsTeamOwner=false のユーザー
	regularUser := testUser("user-id", "regularuser", "user@example.com", domain.RoleUser, true)
	regularUser.IsTeamOwner = false
	userRepo.addUser(regularUser)
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.CreateTeam(context.Background(), usecase.CreateTeamInput{
		Name:       "チーム名",
		CallerID:   "user-id",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestTeamUseCase_CreateTeam_MaxTeams_Exceeded(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	// MaxTeams=1 のユーザー
	ownerUser := testUser("owner-id", "teamowner", "owner@example.com", domain.RoleUser, true)
	ownerUser.IsTeamOwner = true
	ownerUser.MaxTeams = 1
	userRepo.addUser(ownerUser)
	// 既に1チームのオーナー（ownerID で設定）
	existingTeam := testTeam("existing-team", "既存チーム", "owner-id")
	teamRepo.addTeam(existingTeam)
	// per-team owner として追加済み
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "existing-team",
		UserID:   "owner-id",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.CreateTeam(context.Background(), usecase.CreateTeamInput{
		Name:       "新しいチーム",
		CallerID:   "owner-id",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestTeamUseCase_CreateTeam_Admin_CreatorAddedAsOwner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	team, err := uc.CreateTeam(context.Background(), usecase.CreateTeamInput{
		Name:       "CISSP勉強会",
		CallerID:   "admin-id",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Fatalf("チーム作成に失敗しました: %v", err)
	}
	// admin は作成者として per-team owner になること
	owners, err := teamRepo.FindOwners(context.Background(), team.ID)
	if err != nil {
		t.Fatalf("オーナー取得に失敗しました: %v", err)
	}
	if len(owners) != 1 || owners[0].UserID != "admin-id" {
		t.Errorf("admin が per-team owner として追加されていません")
	}
}

func TestTeamUseCase_CreateTeam_DuplicateName(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "CISSP勉強会", "owner-1"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.CreateTeam(context.Background(), usecase.CreateTeamInput{
		Name:       "CISSP勉強会",
		CallerID:   "owner-2",
		CallerRole: domain.RoleAdmin,
	})

	if !errors.Is(err, domain.ErrTeamNameAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTeamNameAlreadyExists)
	}
}

func TestTeamUseCase_ListTeams_Admin_GetsAll(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	teamRepo.addTeam(testTeam("team-2", "チームB", "owner-2"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	teams, err := uc.ListTeams(context.Background(), usecase.ListTeamsInput{
		CallerID:   "admin-id",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Fatalf("チーム一覧取得に失敗しました: %v", err)
	}
	if len(teams) != 2 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 2", len(teams))
	}
}

func TestTeamUseCase_ListTeams_User_GetsOwnOnly(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "user-1"))
	teamRepo.addTeam(testTeam("team-2", "チームB", "owner-2"))
	teamRepo.addTeam(testTeam("team-3", "チームC", "owner-3"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: "team-3", UserID: "user-1", JoinedAt: time.Now()})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	teams, err := uc.ListTeams(context.Background(), usecase.ListTeamsInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("チーム一覧取得に失敗しました: %v", err)
	}
	// team-1（オーナー）とteam-3（メンバー）の2件
	if len(teams) != 2 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 2", len(teams))
	}
}

func TestTeamUseCase_GetTeam_Success_Owner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	out, err := uc.GetTeam(context.Background(), "owner-1", domain.RoleTeamOwner, "team-1")

	if err != nil {
		t.Fatalf("チーム取得に失敗しました: %v", err)
	}
	if out.Team.ID != "team-1" {
		t.Errorf("チームIDが期待値と異なります: got %s, want team-1", out.Team.ID)
	}
}

func TestTeamUseCase_GetTeam_Success_Member(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: "team-1", UserID: "user-2", JoinedAt: time.Now()})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	out, err := uc.GetTeam(context.Background(), "user-2", domain.RoleUser, "team-1")

	if err != nil {
		t.Fatalf("チーム取得に失敗しました: %v", err)
	}
	if out.Team.ID != "team-1" {
		t.Errorf("チームIDが期待値と異なります: got %s, want team-1", out.Team.ID)
	}
}

func TestTeamUseCase_GetTeam_PermissionDenied_NonMember(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.GetTeam(context.Background(), "user-2", domain.RoleUser, "team-1")

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestTeamUseCase_GetTeam_NotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.GetTeam(context.Background(), "admin-id", domain.RoleAdmin, "nonexistent-id")

	if !errors.Is(err, domain.ErrTeamNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTeamNotFound)
	}
}

func TestTeamUseCase_UpdateTeam_Success_Owner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	newName := "チームA 改"
	team, err := uc.UpdateTeam(context.Background(), "owner-1", domain.RoleTeamOwner, "team-1", usecase.UpdateTeamInput{
		Name: &newName,
	})

	if err != nil {
		t.Fatalf("チーム更新に失敗しました: %v", err)
	}
	if team.Name != "チームA 改" {
		t.Errorf("チーム名が期待値と異なります: got %s, want チームA 改", team.Name)
	}
}

func TestTeamUseCase_UpdateTeam_PermissionDenied_NotOwner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	newName := "変更試み"
	_, err := uc.UpdateTeam(context.Background(), "other-user", domain.RoleUser, "team-1", usecase.UpdateTeamInput{
		Name: &newName,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestTeamUseCase_UpdateTeam_DuplicateName(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	teamRepo.addTeam(testTeam("team-2", "チームB", "owner-1"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	existingName := "チームB"
	_, err := uc.UpdateTeam(context.Background(), "owner-1", domain.RoleAdmin, "team-1", usecase.UpdateTeamInput{
		Name: &existingName,
	})

	if !errors.Is(err, domain.ErrTeamNameAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTeamNameAlreadyExists)
	}
}

func TestTeamUseCase_DeleteTeam_Success_Admin(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	if err := uc.DeleteTeam(context.Background(), "admin-id", domain.RoleAdmin, "team-1"); err != nil {
		t.Fatalf("チーム削除に失敗しました: %v", err)
	}

	_, err := teamRepo.FindByID(context.Background(), "team-1")
	if !errors.Is(err, domain.ErrTeamNotFound) {
		t.Error("削除後もチームが存在しています")
	}
}

func TestTeamUseCase_DeleteTeam_PermissionDenied_NotOwner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	err := uc.DeleteTeam(context.Background(), "other-user", domain.RoleUser, "team-1")

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestTeamUseCase_AddMember_Success(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	userRepo := newMockUserRepository()
	userRepo.addUser(testUser("user-2", "bob", "bob@example.com", domain.RoleUser, true))
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	member, err := uc.AddMember(context.Background(), "owner-1", domain.RoleTeamOwner, "team-1", "user-2")

	if err != nil {
		t.Fatalf("メンバー追加に失敗しました: %v", err)
	}
	if member.UserID != "user-2" {
		t.Errorf("UserIDが期待値と異なります: got %s, want user-2", member.UserID)
	}
}

func TestTeamUseCase_AddMember_UserNotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.AddMember(context.Background(), "owner-1", domain.RoleTeamOwner, "team-1", "nonexistent-user")

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestTeamUseCase_AddMember_AlreadyExists(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: "team-1", UserID: "user-2", JoinedAt: time.Now()})
	userRepo := newMockUserRepository()
	userRepo.addUser(testUser("user-2", "bob", "bob@example.com", domain.RoleUser, true))
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.AddMember(context.Background(), "owner-1", domain.RoleTeamOwner, "team-1", "user-2")

	if !errors.Is(err, domain.ErrMemberAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrMemberAlreadyExists)
	}
}

func TestTeamUseCase_AddMember_PermissionDenied_NotOwner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	userRepo := newMockUserRepository()
	userRepo.addUser(testUser("user-2", "bob", "bob@example.com", domain.RoleUser, true))
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.AddMember(context.Background(), "other-user", domain.RoleUser, "team-1", "user-2")

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestTeamUseCase_RemoveMember_Success(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	// owner-1 を per-team owner として追加
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	// user-2 を一般メンバーとして追加
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "user-2",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	if err := uc.RemoveMember(context.Background(), "owner-1", domain.RoleTeamOwner, "team-1", "user-2"); err != nil {
		t.Fatalf("メンバー除外に失敗しました: %v", err)
	}

	isMember, _ := teamRepo.IsMember(context.Background(), "team-1", "user-2")
	if isMember {
		t.Error("除外後もメンバーが存在しています")
	}
}

func TestTeamUseCase_RemoveMember_NotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	// owner-1 を per-team owner として追加
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	err := uc.RemoveMember(context.Background(), "owner-1", domain.RoleTeamOwner, "team-1", "nonexistent-user")

	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestTeamUseCase_RemoveMember_ErrLastTeamOwner は唯一の per-team owner を除外しようとするとエラーになることを確認します。
func TestTeamUseCase_RemoveMember_ErrLastTeamOwner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	// owner-1 のみが per-team owner
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	// admin が唯一のオーナーを除外しようとする
	err := uc.RemoveMember(context.Background(), "admin-id", domain.RoleAdmin, "team-1", "owner-1")

	if !errors.Is(err, domain.ErrLastTeamOwner) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrLastTeamOwner)
	}
}

// TestTeamUseCase_RemoveMember_PermissionDenied は per-team owner でも admin でもないユーザーが除外しようとするとエラーになることを確認します。
func TestTeamUseCase_RemoveMember_PermissionDenied(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "member-1",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	// 一般メンバーが別のメンバーを除外しようとする
	err := uc.RemoveMember(context.Background(), "member-1", domain.RoleUser, "team-1", "owner-1")

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// --- ChangeMemberRole のテスト ---

func TestTeamUseCase_ChangeMemberRole_Success_OwnerToMember_ByAdmin(t *testing.T) {
	// admin はオーナーを降格できる（複数オーナーがいる場合）
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-2",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	member, err := uc.ChangeMemberRole(context.Background(), usecase.ChangeMemberRoleInput{
		CallerID:     "admin-id",
		CallerRole:   domain.RoleAdmin,
		TeamID:       "team-1",
		TargetUserID: "owner-2",
		Role:         domain.MemberRoleMember,
	})

	if err != nil {
		t.Fatalf("ロール変更に失敗しました: %v", err)
	}
	if member.Role != domain.MemberRoleMember {
		t.Errorf("ロールが期待値と異なります: got %s, want %s", member.Role, domain.MemberRoleMember)
	}
}

func TestTeamUseCase_ChangeMemberRole_Success_MemberToOwner_ByTeamOwner(t *testing.T) {
	// per-team owner が一般メンバーを owner に昇格
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "member-1",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	member, err := uc.ChangeMemberRole(context.Background(), usecase.ChangeMemberRoleInput{
		CallerID:     "owner-1",
		CallerRole:   domain.RoleUser,
		TeamID:       "team-1",
		TargetUserID: "member-1",
		Role:         domain.MemberRoleOwner,
	})

	if err != nil {
		t.Fatalf("ロール変更に失敗しました: %v", err)
	}
	if member.Role != domain.MemberRoleOwner {
		t.Errorf("ロールが期待値と異なります: got %s, want %s", member.Role, domain.MemberRoleOwner)
	}
}

func TestTeamUseCase_ChangeMemberRole_PermissionDenied_NotTeamOwner(t *testing.T) {
	// per-team owner でも admin でもないユーザーが変更しようとする
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "member-1",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.ChangeMemberRole(context.Background(), usecase.ChangeMemberRoleInput{
		CallerID:     "member-1", // 一般メンバーは変更できない
		CallerRole:   domain.RoleUser,
		TeamID:       "team-1",
		TargetUserID: "member-1",
		Role:         domain.MemberRoleOwner,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestTeamUseCase_ChangeMemberRole_ErrLastTeamOwner(t *testing.T) {
	// チームの最後のオーナーを降格しようとするとエラー
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.ChangeMemberRole(context.Background(), usecase.ChangeMemberRoleInput{
		CallerID:     "admin-id",
		CallerRole:   domain.RoleAdmin,
		TeamID:       "team-1",
		TargetUserID: "owner-1", // 唯一のオーナーを降格
		Role:         domain.MemberRoleMember,
	})

	if !errors.Is(err, domain.ErrLastTeamOwner) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrLastTeamOwner)
	}
}

func TestTeamUseCase_ChangeMemberRole_MemberNotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	_, err := uc.ChangeMemberRole(context.Background(), usecase.ChangeMemberRoleInput{
		CallerID:     "admin-id",
		CallerRole:   domain.RoleAdmin,
		TeamID:       "team-1",
		TargetUserID: "nonexistent-user",
		Role:         domain.MemberRoleOwner,
	})

	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// --- ListMemberStats のテスト ---

func TestTeamUseCase_ListMemberStats_Success_Member(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "user-2",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})

	userRepo := newMockUserRepository()
	userRepo.addUser(testUser("owner-1", "owner", "owner@example.com", domain.RoleUser, true))
	userRepo.addUser(testUser("user-2", "user2", "user2@example.com", domain.RoleUser, true))

	questionRepo := newMockQuestionRepository()
	questionRepo.questions["q-1"] = &domain.Question{ID: "q-1", TeamID: "team-1", CreatedBy: "owner-1"}
	questionRepo.questions["q-2"] = &domain.Question{ID: "q-2", TeamID: "team-1", CreatedBy: "owner-1"}
	questionRepo.questions["q-3"] = &domain.Question{ID: "q-3", TeamID: "team-1", CreatedBy: "user-2"}

	commentRepo := newMockCommentRepository()
	commentRepo.comments[commentKey("q-1", "c-1")] = &domain.Comment{ID: "c-1", QuestionID: "q-1", CreatedBy: "user-2"}
	commentRepo.comments[commentKey("q-2", "c-2")] = &domain.Comment{ID: "c-2", QuestionID: "q-2", CreatedBy: "owner-1"}
	commentRepo.comments[commentKey("q-2", "c-3")] = &domain.Comment{ID: "c-3", QuestionID: "q-2", CreatedBy: "user-2"}

	uc := usecase.NewTeamUseCase(teamRepo, userRepo, questionRepo, commentRepo)

	stats, err := uc.ListMemberStats(context.Background(), "user-2", domain.RoleUser, "team-1")
	if err != nil {
		t.Fatalf("ListMemberStats に失敗しました: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("メンバー数が期待値と異なります: got %d, want 2", len(stats))
	}

	// ユーザーIDでインデックスを作成して検証
	statsByUser := make(map[string]*usecase.MemberStats)
	for _, s := range stats {
		statsByUser[s.UserID] = s
	}

	ownerStats := statsByUser["owner-1"]
	if ownerStats == nil {
		t.Fatal("owner-1 の統計が存在しません")
	}
	if ownerStats.QuestionCount != 2 {
		t.Errorf("owner-1 の問題数が期待値と異なります: got %d, want 2", ownerStats.QuestionCount)
	}
	if ownerStats.CommentCount != 1 {
		t.Errorf("owner-1 のコメント数が期待値と異なります: got %d, want 1", ownerStats.CommentCount)
	}
	if ownerStats.Role != domain.MemberRoleOwner {
		t.Errorf("owner-1 のロールが期待値と異なります: got %s, want owner", ownerStats.Role)
	}

	user2Stats := statsByUser["user-2"]
	if user2Stats == nil {
		t.Fatal("user-2 の統計が存在しません")
	}
	if user2Stats.QuestionCount != 1 {
		t.Errorf("user-2 の問題数が期待値と異なります: got %d, want 1", user2Stats.QuestionCount)
	}
	if user2Stats.CommentCount != 2 {
		t.Errorf("user-2 のコメント数が期待値と異なります: got %d, want 2", user2Stats.CommentCount)
	}
}

func TestTeamUseCase_ListMemberStats_PermissionDenied_NonMember(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})

	userRepo := newMockUserRepository()
	questionRepo := newMockQuestionRepository()
	commentRepo := newMockCommentRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, questionRepo, commentRepo)

	// 非メンバーがアクセスしようとする
	_, err := uc.ListMemberStats(context.Background(), "stranger-id", domain.RoleUser, "team-1")
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestTeamUseCase_ListMemberStats_Success_Admin(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})

	userRepo := newMockUserRepository()
	userRepo.addUser(testUser("owner-1", "owner", "owner@example.com", domain.RoleUser, true))
	questionRepo := newMockQuestionRepository()
	commentRepo := newMockCommentRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, questionRepo, commentRepo)

	// admin は非メンバーでもアクセス可能
	stats, err := uc.ListMemberStats(context.Background(), "admin-id", domain.RoleAdmin, "team-1")
	if err != nil {
		t.Fatalf("ListMemberStats に失敗しました: %v", err)
	}
	if len(stats) != 1 {
		t.Errorf("メンバー数が期待値と異なります: got %d, want 1", len(stats))
	}
}

func TestTeamUseCase_ListMemberStats_TeamNotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	questionRepo := newMockQuestionRepository()
	commentRepo := newMockCommentRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, questionRepo, commentRepo)

	_, err := uc.ListMemberStats(context.Background(), "admin-id", domain.RoleAdmin, "nonexistent-team")
	if !errors.Is(err, domain.ErrTeamNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTeamNotFound)
	}
}

func TestTeamUseCase_ListMemberStats_LastLoginAt(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})

	loginTime := time.Now().Add(-1 * time.Hour)
	owner := testUser("owner-1", "owner", "owner@example.com", domain.RoleUser, true)
	owner.LastLoginAt = &loginTime

	userRepo := newMockUserRepository()
	userRepo.addUser(owner)
	questionRepo := &mockQuestionRepository{}
	commentRepo := newMockCommentRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, questionRepo, commentRepo)

	stats, err := uc.ListMemberStats(context.Background(), "admin-id", domain.RoleAdmin, "team-1")
	if err != nil {
		t.Fatalf("ListMemberStats に失敗しました: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("メンバー数が期待値と異なります: got %d, want 1", len(stats))
	}
	if stats[0].LastLoginAt == nil {
		t.Fatal("LastLoginAt が nil です")
	}
	if !stats[0].LastLoginAt.Equal(loginTime) {
		t.Errorf("LastLoginAt が期待値と異なります: got %v, want %v", *stats[0].LastLoginAt, loginTime)
	}
}

// countingUserRepository は UserRepository の呼び出し回数を記録するモックです。
// ListMemberStats が一括取得（List）のみ使用し、個別取得（FindByID）を呼ばないことを検証します。
type countingUserRepository struct {
	mockUserRepository
	listCallCount     int
	findByIDCallCount int
}

func newCountingUserRepository() *countingUserRepository {
	return &countingUserRepository{
		mockUserRepository: mockUserRepository{
			users:      make(map[string]*domain.User),
			byUsername: make(map[string]*domain.User),
			byEmail:    make(map[string]*domain.User),
		},
	}
}

func (m *countingUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	m.listCallCount++
	return m.mockUserRepository.List(ctx)
}

func (m *countingUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	m.findByIDCallCount++
	return m.mockUserRepository.FindByID(ctx, id)
}

// TestTeamUseCase_ListMemberStats_UsesListNotFindByID は、
// ListMemberStats が FindByID をループ呼び出し（N+1）せず、
// List を1回だけ呼び出す一括取得パターンを使用していることを検証します。
func TestTeamUseCase_ListMemberStats_UsesListNotFindByID(t *testing.T) {
	teamRepo := newMockTeamRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	// メンバーを複数追加して N+1 が発生しやすい状況を作る
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "user-2",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "user-3",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})

	userRepo := newCountingUserRepository()
	userRepo.addUser(testUser("owner-1", "owner", "owner@example.com", domain.RoleUser, true))
	userRepo.addUser(testUser("user-2", "user2", "user2@example.com", domain.RoleUser, true))
	userRepo.addUser(testUser("user-3", "user3", "user3@example.com", domain.RoleUser, true))

	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	stats, err := uc.ListMemberStats(context.Background(), "admin-id", domain.RoleAdmin, "team-1")
	if err != nil {
		t.Fatalf("ListMemberStats に失敗しました: %v", err)
	}
	if len(stats) != 3 {
		t.Fatalf("メンバー数が期待値と異なります: got %d, want 3", len(stats))
	}

	// List が1回だけ呼ばれていること（GCS アクセスが O(1) であること）を検証
	if userRepo.listCallCount != 1 {
		t.Errorf("List の呼び出し回数が期待値と異なります: got %d, want 1", userRepo.listCallCount)
	}

	// FindByID がまったく呼ばれていないこと（N+1 が解消されていること）を検証
	if userRepo.findByIDCallCount != 0 {
		t.Errorf("FindByID が呼ばれています（N+1 が発生しています）: got %d calls, want 0", userRepo.findByIDCallCount)
	}
}
