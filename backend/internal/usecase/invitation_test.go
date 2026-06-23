package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// --- 招待リポジトリモック ---

type mockInvitationRepository struct {
	invitations map[string]*domain.Invitation
	saveErr     error
}

func newMockInvitationRepository() *mockInvitationRepository {
	return &mockInvitationRepository{
		invitations: make(map[string]*domain.Invitation),
	}
}

func (m *mockInvitationRepository) Save(_ context.Context, inv *domain.Invitation) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.invitations[inv.ID] = inv
	return nil
}

func (m *mockInvitationRepository) FindByID(_ context.Context, id string) (*domain.Invitation, error) {
	if inv, ok := m.invitations[id]; ok {
		return inv, nil
	}
	return nil, domain.ErrInvitationNotFound
}

func (m *mockInvitationRepository) ListByInvitee(_ context.Context, inviteeUserID string) ([]*domain.Invitation, error) {
	result := make([]*domain.Invitation, 0)
	for _, inv := range m.invitations {
		if inv.InviteeUserID == inviteeUserID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (m *mockInvitationRepository) ListByTeam(_ context.Context, teamID string) ([]*domain.Invitation, error) {
	result := make([]*domain.Invitation, 0)
	for _, inv := range m.invitations {
		if inv.TeamID == teamID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (m *mockInvitationRepository) FindPendingByTeamAndInvitee(_ context.Context, teamID, inviteeUserID string) (*domain.Invitation, error) {
	for _, inv := range m.invitations {
		if inv.TeamID == teamID && inv.InviteeUserID == inviteeUserID && inv.Status == domain.StatusPending {
			return inv, nil
		}
	}
	return nil, domain.ErrInvitationNotFound
}

// --- テストヘルパー ---

func testInvitation(id, teamID, invitedBy, inviteeID string, status domain.InvitationStatus) *domain.Invitation {
	return &domain.Invitation{
		ID:                id,
		TeamID:            teamID,
		InvitedBy:         invitedBy,
		InviteeIdentifier: inviteeID,
		InviteeUserID:     inviteeID,
		Status:            status,
		CreatedAt:         time.Now().UTC(),
	}
}

// --- SendInvitation のテスト ---

func TestInvitationUseCase_SendInvitation_Success_ByOwner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	team := testTeam("team-1", "チームA", "owner-1")
	teamRepo.addTeam(team)
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	invitee := testUser("invitee-1", "bob", "bob@example.com", domain.RoleUser, true)
	userRepo.addUser(invitee)

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	// ユーザー名で検索
	inv, err := uc.SendInvitation(context.Background(), usecase.SendInvitationInput{
		CallerID:          "owner-1",
		CallerRole:        domain.RoleUser,
		TeamID:            "team-1",
		InviteeIdentifier: "bob",
	})

	if err != nil {
		t.Fatalf("招待送信に失敗しました: %v", err)
	}
	if inv.InviteeUserID != "invitee-1" {
		t.Errorf("InviteeUserID が期待値と異なります: got %s, want invitee-1", inv.InviteeUserID)
	}
	if inv.Status != domain.StatusPending {
		t.Errorf("Status が期待値と異なります: got %s, want pending", inv.Status)
	}
}

func TestInvitationUseCase_SendInvitation_Success_ByAdmin(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	invitee := testUser("invitee-1", "bob", "bob@example.com", domain.RoleUser, true)
	userRepo.addUser(invitee)

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	// メールアドレスで検索
	inv, err := uc.SendInvitation(context.Background(), usecase.SendInvitationInput{
		CallerID:          "admin-id",
		CallerRole:        domain.RoleAdmin,
		TeamID:            "team-1",
		InviteeIdentifier: "bob@example.com",
	})

	if err != nil {
		t.Fatalf("招待送信に失敗しました: %v", err)
	}
	if inv.TeamID != "team-1" {
		t.Errorf("TeamID が期待値と異なります: got %s, want team-1", inv.TeamID)
	}
}

func TestInvitationUseCase_SendInvitation_Success_ResolveByUsername(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	invitee := testUser("invitee-1", "bob", "bob@example.com", domain.RoleUser, true)
	userRepo.addUser(invitee)

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	// ユーザー名で解決
	inv, err := uc.SendInvitation(context.Background(), usecase.SendInvitationInput{
		CallerID:          "owner-1",
		CallerRole:        domain.RoleUser,
		TeamID:            "team-1",
		InviteeIdentifier: "bob",
	})

	if err != nil {
		t.Fatalf("招待送信に失敗しました: %v", err)
	}
	if inv.InviteeUserID != "invitee-1" {
		t.Errorf("InviteeUserID が期待値と異なります: got %s, want invitee-1", inv.InviteeUserID)
	}
}

func TestInvitationUseCase_SendInvitation_Success_ResolveByEmail(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	invitee := testUser("invitee-1", "bob", "bob@example.com", domain.RoleUser, true)
	userRepo.addUser(invitee)

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	// メールアドレスで解決
	inv, err := uc.SendInvitation(context.Background(), usecase.SendInvitationInput{
		CallerID:          "owner-1",
		CallerRole:        domain.RoleUser,
		TeamID:            "team-1",
		InviteeIdentifier: "bob@example.com",
	})

	if err != nil {
		t.Fatalf("招待送信に失敗しました: %v", err)
	}
	if inv.InviteeUserID != "invitee-1" {
		t.Errorf("InviteeUserID が期待値と異なります: got %s, want invitee-1", inv.InviteeUserID)
	}
}

func TestInvitationUseCase_SendInvitation_PermissionDenied_NotOwner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

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
	invitee := testUser("invitee-1", "bob", "bob@example.com", domain.RoleUser, true)
	userRepo.addUser(invitee)

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	_, err := uc.SendInvitation(context.Background(), usecase.SendInvitationInput{
		CallerID:          "member-1",
		CallerRole:        domain.RoleUser,
		TeamID:            "team-1",
		InviteeIdentifier: "invitee-1",
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestInvitationUseCase_SendInvitation_UserNotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	_, err := uc.SendInvitation(context.Background(), usecase.SendInvitationInput{
		CallerID:          "owner-1",
		CallerRole:        domain.RoleUser,
		TeamID:            "team-1",
		InviteeIdentifier: "nonexistent@example.com",
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestInvitationUseCase_SendInvitation_MemberAlreadyExists(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	// invitee-1 を既にメンバーとして追加
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "invitee-1",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})
	invitee := testUser("invitee-1", "bob", "bob@example.com", domain.RoleUser, true)
	userRepo.addUser(invitee)

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	// メールアドレスで検索
	_, err := uc.SendInvitation(context.Background(), usecase.SendInvitationInput{
		CallerID:          "owner-1",
		CallerRole:        domain.RoleUser,
		TeamID:            "team-1",
		InviteeIdentifier: "bob@example.com",
	})

	if !errors.Is(err, domain.ErrMemberAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrMemberAlreadyExists)
	}
}

func TestInvitationUseCase_SendInvitation_InvitationAlreadyExists(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	invitee := testUser("invitee-1", "bob", "bob@example.com", domain.RoleUser, true)
	userRepo.addUser(invitee)

	// pending 招待を事前に追加（InviteeUserID: invitee-1 で保存）
	existingInv := testInvitation(uuid.NewString(), "team-1", "owner-1", "invitee-1", domain.StatusPending)
	invRepo.invitations[existingInv.ID] = existingInv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	// メールアドレスで検索（ユーザー解決後に pending 招待の重複を検出する）
	_, err := uc.SendInvitation(context.Background(), usecase.SendInvitationInput{
		CallerID:          "owner-1",
		CallerRole:        domain.RoleUser,
		TeamID:            "team-1",
		InviteeIdentifier: "bob@example.com",
	})

	if !errors.Is(err, domain.ErrInvitationAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvitationAlreadyExists)
	}
}

// --- ListMyInvitations のテスト ---

func TestInvitationUseCase_ListMyInvitations_Success(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	// 自分宛の招待を2件、別ユーザー宛の招待を1件追加
	inv1 := testInvitation(uuid.NewString(), "team-1", "owner-1", "user-1", domain.StatusPending)
	inv2 := testInvitation(uuid.NewString(), "team-2", "owner-2", "user-1", domain.StatusAccepted)
	inv3 := testInvitation(uuid.NewString(), "team-1", "owner-1", "user-2", domain.StatusPending)
	invRepo.invitations[inv1.ID] = inv1
	invRepo.invitations[inv2.ID] = inv2
	invRepo.invitations[inv3.ID] = inv3

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	invitations, err := uc.ListMyInvitations(context.Background(), usecase.ListMyInvitationsInput{
		CallerID: "user-1",
	})

	if err != nil {
		t.Fatalf("招待一覧取得に失敗しました: %v", err)
	}
	if len(invitations) != 2 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 2", len(invitations))
	}
}

func TestInvitationUseCase_ListMyInvitations_Empty(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	invitations, err := uc.ListMyInvitations(context.Background(), usecase.ListMyInvitationsInput{
		CallerID: "user-1",
	})

	if err != nil {
		t.Fatalf("招待一覧取得に失敗しました: %v", err)
	}
	if len(invitations) != 0 {
		t.Errorf("招待が存在しないはずですが %d 件取得されました", len(invitations))
	}
}

// TestInvitationUseCase_ListMyInvitations_WithMeta_Success は
// チーム名・招待者表示名が正しく付与されることを検証します。
func TestInvitationUseCase_ListMyInvitations_WithMeta_Success(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	// チームと招待者ユーザーを登録する
	teamRepo.addTeam(testTeam("team-1", "CISSPチーム", "owner-1"))
	inviter := testUser("owner-1", "alice", "alice@example.com", domain.RoleUser, true)
	inviter.DisplayName = "Alice"
	userRepo.addUser(inviter)

	// 自分宛の招待を1件追加する
	inv := testInvitation(uuid.NewString(), "team-1", "owner-1", "user-1", domain.StatusPending)
	invRepo.invitations[inv.ID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	results, err := uc.ListMyInvitations(context.Background(), usecase.ListMyInvitationsInput{
		CallerID: "user-1",
	})

	if err != nil {
		t.Fatalf("招待一覧取得に失敗しました: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("取得件数が期待値と異なります: got %d, want 1", len(results))
	}
	got := results[0]
	if got.TeamName != "CISSPチーム" {
		t.Errorf("TeamName が期待値と異なります: got %q, want %q", got.TeamName, "CISSPチーム")
	}
	if got.InviterDisplayName != "Alice" {
		t.Errorf("InviterDisplayName が期待値と異なります: got %q, want %q", got.InviterDisplayName, "Alice")
	}
}

// TestInvitationUseCase_ListMyInvitations_TeamNotFound は
// 招待の TeamID に対応するチームが存在しない場合、TeamName が空文字にフォールバックされることを検証します。
func TestInvitationUseCase_ListMyInvitations_TeamNotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	// チームは登録しない（招待の TeamID "team-unknown" は teamRepo に存在しない）
	inviter := testUser("owner-1", "alice", "alice@example.com", domain.RoleUser, true)
	inviter.DisplayName = "Alice"
	userRepo.addUser(inviter)

	inv := testInvitation(uuid.NewString(), "team-unknown", "owner-1", "user-1", domain.StatusPending)
	invRepo.invitations[inv.ID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	results, err := uc.ListMyInvitations(context.Background(), usecase.ListMyInvitationsInput{
		CallerID: "user-1",
	})

	if err != nil {
		t.Fatalf("チームが存在しなくても招待一覧取得はエラーにならないはず: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("取得件数が期待値と異なります: got %d, want 1", len(results))
	}
	// チームが見つからない場合は TeamName が空文字になる
	if results[0].TeamName != "" {
		t.Errorf("TeamName はチーム未登録時に空文字を期待しますが got %q", results[0].TeamName)
	}
	// 招待者は登録済みのため InviterDisplayName は正しく設定される
	if results[0].InviterDisplayName != "Alice" {
		t.Errorf("InviterDisplayName が期待値と異なります: got %q, want %q", results[0].InviterDisplayName, "Alice")
	}
}

// TestInvitationUseCase_ListMyInvitations_InviterNotFound は
// 招待の InvitedBy に対応するユーザーが存在しない場合、InviterDisplayName が空文字にフォールバックされることを検証します。
func TestInvitationUseCase_ListMyInvitations_InviterNotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	// チームは登録するが、招待者ユーザーは登録しない
	teamRepo.addTeam(testTeam("team-1", "CISSPチーム", "owner-unknown"))

	inv := testInvitation(uuid.NewString(), "team-1", "owner-unknown", "user-1", domain.StatusPending)
	invRepo.invitations[inv.ID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	results, err := uc.ListMyInvitations(context.Background(), usecase.ListMyInvitationsInput{
		CallerID: "user-1",
	})

	if err != nil {
		t.Fatalf("招待者が存在しなくても招待一覧取得はエラーにならないはず: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("取得件数が期待値と異なります: got %d, want 1", len(results))
	}
	// チームは登録済みのため TeamName は正しく設定される
	if results[0].TeamName != "CISSPチーム" {
		t.Errorf("TeamName が期待値と異なります: got %q, want %q", results[0].TeamName, "CISSPチーム")
	}
	// 招待者が見つからない場合は InviterDisplayName が空文字になる
	if results[0].InviterDisplayName != "" {
		t.Errorf("InviterDisplayName は招待者未登録時に空文字を期待しますが got %q", results[0].InviterDisplayName)
	}
}

// TestInvitationUseCase_ListMyInvitations_ZeroInvitations は
// 自分宛の招待が0件の場合、空スライスが返されることを検証します。
func TestInvitationUseCase_ListMyInvitations_ZeroInvitations(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	// 別ユーザー宛の招待は存在するが、user-1 宛はない
	otherInv := testInvitation(uuid.NewString(), "team-1", "owner-1", "user-2", domain.StatusPending)
	invRepo.invitations[otherInv.ID] = otherInv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	results, err := uc.ListMyInvitations(context.Background(), usecase.ListMyInvitationsInput{
		CallerID: "user-1",
	})

	if err != nil {
		t.Fatalf("招待一覧取得に失敗しました: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("user-1 宛の招待は0件のはずですが %d 件取得されました", len(results))
	}
}

// --- RespondInvitation（accept）のテスト ---

func TestInvitationUseCase_RespondInvitation_Accept_Success(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))

	invID := uuid.NewString()
	inv := testInvitation(invID, "team-1", "owner-1", "invitee-1", domain.StatusPending)
	invRepo.invitations[invID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	result, err := uc.RespondInvitation(context.Background(), usecase.RespondInvitationInput{
		CallerID:     "invitee-1",
		InvitationID: invID,
		Status:       domain.StatusAccepted,
	})

	if err != nil {
		t.Fatalf("招待応答に失敗しました: %v", err)
	}
	if result.Status != domain.StatusAccepted {
		t.Errorf("Status が期待値と異なります: got %s, want accepted", result.Status)
	}

	// メンバーに追加されていることを確認
	isMember, _ := teamRepo.IsMember(context.Background(), "team-1", "invitee-1")
	if !isMember {
		t.Error("受諾後にメンバーとして追加されていません")
	}
}

func TestInvitationUseCase_RespondInvitation_Accept_NotMyInvitation(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))

	invID := uuid.NewString()
	inv := testInvitation(invID, "team-1", "owner-1", "invitee-1", domain.StatusPending)
	invRepo.invitations[invID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	// invitee-2 が invitee-1 宛の招待に応答しようとする
	_, err := uc.RespondInvitation(context.Background(), usecase.RespondInvitationInput{
		CallerID:     "invitee-2",
		InvitationID: invID,
		Status:       domain.StatusAccepted,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestInvitationUseCase_RespondInvitation_Accept_AlreadyProcessed(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))

	invID := uuid.NewString()
	// 既に accepted になっている招待
	inv := testInvitation(invID, "team-1", "owner-1", "invitee-1", domain.StatusAccepted)
	invRepo.invitations[invID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	_, err := uc.RespondInvitation(context.Background(), usecase.RespondInvitationInput{
		CallerID:     "invitee-1",
		InvitationID: invID,
		Status:       domain.StatusAccepted,
	})

	if !errors.Is(err, domain.ErrInvitationNotPending) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvitationNotPending)
	}
}

// TestInvitationUseCase_RespondInvitation_Accept_MemberAlreadyExists は
// 招待受諾時に対象ユーザーが既にメンバーだった場合、エラーを返さずに
// 招待ステータスを accepted に更新して正常終了することを検証します。
// （重複受諾防止のため、ErrMemberAlreadyExists は無視して正常終了とする設計）
func TestInvitationUseCase_RespondInvitation_Accept_MemberAlreadyExists(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	// invitee-1 を既にメンバーとして追加
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "invitee-1",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})

	invID := uuid.NewString()
	inv := testInvitation(invID, "team-1", "owner-1", "invitee-1", domain.StatusPending)
	invRepo.invitations[invID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	result, err := uc.RespondInvitation(context.Background(), usecase.RespondInvitationInput{
		CallerID:     "invitee-1",
		InvitationID: invID,
		Status:       domain.StatusAccepted,
	})

	// ErrMemberAlreadyExists は無視して正常終了（nil エラー）を期待する
	if err != nil {
		t.Errorf("エラーは期待しないが got: %v", err)
	}
	// 招待ステータスが accepted に更新されていることを確認
	if result != nil && result.Status != domain.StatusAccepted {
		t.Errorf("招待ステータスが期待値と異なります: got %v, want %v", result.Status, domain.StatusAccepted)
	}
}

// --- RespondInvitation（reject）のテスト ---

func TestInvitationUseCase_RespondInvitation_Reject_Success(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))

	invID := uuid.NewString()
	inv := testInvitation(invID, "team-1", "owner-1", "invitee-1", domain.StatusPending)
	invRepo.invitations[invID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	result, err := uc.RespondInvitation(context.Background(), usecase.RespondInvitationInput{
		CallerID:     "invitee-1",
		InvitationID: invID,
		Status:       domain.StatusRejected,
	})

	if err != nil {
		t.Fatalf("招待拒否に失敗しました: %v", err)
	}
	if result.Status != domain.StatusRejected {
		t.Errorf("Status が期待値と異なります: got %s, want rejected", result.Status)
	}

	// メンバーに追加されていないことを確認
	isMember, _ := teamRepo.IsMember(context.Background(), "team-1", "invitee-1")
	if isMember {
		t.Error("拒否後にメンバーとして追加されています（不正）")
	}
}

func TestInvitationUseCase_RespondInvitation_Reject_NotMyInvitation(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	invID := uuid.NewString()
	inv := testInvitation(invID, "team-1", "owner-1", "invitee-1", domain.StatusPending)
	invRepo.invitations[invID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	_, err := uc.RespondInvitation(context.Background(), usecase.RespondInvitationInput{
		CallerID:     "other-user",
		InvitationID: invID,
		Status:       domain.StatusRejected,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

func TestInvitationUseCase_RespondInvitation_Reject_AlreadyProcessed(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	invRepo := newMockInvitationRepository()

	invID := uuid.NewString()
	// 既に rejected になっている招待
	inv := testInvitation(invID, "team-1", "owner-1", "invitee-1", domain.StatusRejected)
	invRepo.invitations[invID] = inv

	uc := usecase.NewInvitationUseCase(invRepo, teamRepo, userRepo)

	_, err := uc.RespondInvitation(context.Background(), usecase.RespondInvitationInput{
		CallerID:     "invitee-1",
		InvitationID: invID,
		Status:       domain.StatusRejected,
	})

	if !errors.Is(err, domain.ErrInvitationNotPending) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvitationNotPending)
	}
}

// --- LeaveTeam のテスト ---

func TestTeamUseCase_LeaveTeam_Success(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()

	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	// owner-1 を per-team owner として追加
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	// member-1 を一般メンバーとして追加
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "member-1",
		Role:     domain.MemberRoleMember,
		JoinedAt: time.Now(),
	})

	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	if err := uc.LeaveTeam(context.Background(), "member-1", "team-1"); err != nil {
		t.Fatalf("チーム脱退に失敗しました: %v", err)
	}

	isMember, _ := teamRepo.IsMember(context.Background(), "team-1", "member-1")
	if isMember {
		t.Error("脱退後もメンバーとして登録されています")
	}
}

func TestTeamUseCase_LeaveTeam_TeamNotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	err := uc.LeaveTeam(context.Background(), "user-1", "nonexistent-team")

	if !errors.Is(err, domain.ErrTeamNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTeamNotFound)
	}
}

func TestTeamUseCase_LeaveTeam_MemberNotFound(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	// チームのメンバーでないユーザーが脱退しようとする
	err := uc.LeaveTeam(context.Background(), "nonmember-user", "team-1")

	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

func TestTeamUseCase_LeaveTeam_LastOwner(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	// owner-1 のみが per-team owner
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID:   "team-1",
		UserID:   "owner-1",
		Role:     domain.MemberRoleOwner,
		JoinedAt: time.Now(),
	})
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	// 唯一のオーナーが脱退しようとする
	err := uc.LeaveTeam(context.Background(), "owner-1", "team-1")

	if !errors.Is(err, domain.ErrLastTeamOwner) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrLastTeamOwner)
	}
}

func TestTeamUseCase_LeaveTeam_OwnerCanLeaveIfMultipleOwners(t *testing.T) {
	teamRepo := newMockTeamRepository()
	userRepo := newMockUserRepository()
	teamRepo.addTeam(testTeam("team-1", "チームA", "owner-1"))
	// owner-1 と owner-2 が両方オーナー
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
	uc := usecase.NewTeamUseCase(teamRepo, userRepo, nil, nil)

	// owner-1 が脱退（owner-2 が残るため可能）
	if err := uc.LeaveTeam(context.Background(), "owner-1", "team-1"); err != nil {
		t.Fatalf("チーム脱退に失敗しました: %v", err)
	}

	isMember, _ := teamRepo.IsMember(context.Background(), "team-1", "owner-1")
	if isMember {
		t.Error("脱退後もメンバーとして登録されています")
	}
}
