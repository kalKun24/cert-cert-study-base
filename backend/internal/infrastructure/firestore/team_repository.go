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

// teamRecord はFirestoreに保存するチームレコードです。
// コレクション: teams/{teamID}
type teamRecord struct {
	ID          string    `firestore:"id"`
	Name        string    `firestore:"name"`
	Description string    `firestore:"description"`
	OwnerID     string    `firestore:"owner_id"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
}

// teamMemberRecord はFirestoreに保存するメンバーレコードです。
// コレクション: teams/{teamID}/members/{userID}
type teamMemberRecord struct {
	TeamID   string    `firestore:"team_id"`
	UserID   string    `firestore:"user_id"`
	Role     string    `firestore:"role"`
	JoinedAt time.Time `firestore:"joined_at"`
}

func toTeamRecord(t *domain.Team) teamRecord {
	return teamRecord{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		OwnerID:     t.OwnerID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toTeam(r teamRecord) *domain.Team {
	return &domain.Team{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		OwnerID:     r.OwnerID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func toTeamMember(r teamMemberRecord) *domain.TeamMember {
	return &domain.TeamMember{
		TeamID:   r.TeamID,
		UserID:   r.UserID,
		Role:     domain.MemberRole(r.Role),
		JoinedAt: r.JoinedAt,
	}
}

// FirestoreTeamRepository はCloud Firestoreにチームデータを永続化するリポジトリです。
// domain.TeamRepository インターフェースを実装します。
type FirestoreTeamRepository struct {
	client *fs.Client
}

// NewFirestoreTeamRepository は FirestoreTeamRepository を生成します。
func NewFirestoreTeamRepository(client *fs.Client) *FirestoreTeamRepository {
	return &FirestoreTeamRepository{client: client}
}

func (r *FirestoreTeamRepository) teamsCol() *fs.CollectionRef {
	return r.client.Collection("teams")
}

func (r *FirestoreTeamRepository) membersCol(teamID string) *fs.CollectionRef {
	return r.client.Collection("teams").Doc(teamID).Collection("members")
}

// FindByID はIDでチームを検索します。
func (r *FirestoreTeamRepository) FindByID(ctx context.Context, id string) (*domain.Team, error) {
	doc, err := r.teamsCol().Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrTeamNotFound
		}
		return nil, fmt.Errorf("チームの取得に失敗しました: %w", err)
	}

	var rec teamRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("チームデータのデコードに失敗しました: %w", err)
	}
	return toTeam(rec), nil
}

// FindByName はチーム名でチームを検索します。
func (r *FirestoreTeamRepository) FindByName(ctx context.Context, name string) (*domain.Team, error) {
	iter := r.teamsCol().Where("name", "==", name).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, domain.ErrTeamNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("チーム検索に失敗しました: %w", err)
	}

	var rec teamRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("チームデータのデコードに失敗しました: %w", err)
	}
	return toTeam(rec), nil
}

// List は全チームの一覧を返します。
func (r *FirestoreTeamRepository) List(ctx context.Context) ([]*domain.Team, error) {
	iter := r.teamsCol().Documents(ctx)
	defer iter.Stop()

	var teams []*domain.Team
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("チーム一覧の取得に失敗しました: %w", err)
		}

		var rec teamRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("チームデータのデコードに失敗しました: %w", err)
		}
		teams = append(teams, toTeam(rec))
	}

	if teams == nil {
		teams = []*domain.Team{}
	}
	return teams, nil
}

// ListByOwnerOrMember はユーザーがオーナーまたはメンバーであるチームを返します。
func (r *FirestoreTeamRepository) ListByOwnerOrMember(ctx context.Context, userID string) ([]*domain.Team, error) {
	// メンバーとして所属するチームIDを収集
	// Firestore の collectionGroup クエリでも実現できますが、
	// サブコレクション設計との整合のため全チームを走査する方式を採用します。
	allTeams, err := r.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("チーム一覧の取得に失敗しました: %w", err)
	}

	var result []*domain.Team
	for _, team := range allTeams {
		if team.OwnerID == userID {
			result = append(result, team)
			continue
		}
		// メンバーサブコレクションを確認
		_, err := r.membersCol(team.ID).Doc(userID).Get(ctx)
		if err == nil {
			result = append(result, team)
		} else if status.Code(err) != codes.NotFound {
			return nil, fmt.Errorf("メンバー確認に失敗しました (teamID=%s): %w", team.ID, err)
		}
	}

	if result == nil {
		result = []*domain.Team{}
	}
	return result, nil
}

// Save はチームを新規作成または更新します（Upsert）。
func (r *FirestoreTeamRepository) Save(ctx context.Context, team *domain.Team) error {
	rec := toTeamRecord(team)
	_, err := r.teamsCol().Doc(team.ID).Set(ctx, rec)
	if err != nil {
		return fmt.Errorf("チームの保存に失敗しました: %w", err)
	}
	return nil
}

// Delete はIDで指定したチームと全サブコレクションをカスケード削除します。
// Firestoreはドキュメント削除時にサブコレクションを自動削除しないため、
// questions（+各コメント）、notes（+各コメント）、tags、members を明示的に削除します。
func (r *FirestoreTeamRepository) Delete(ctx context.Context, id string) error {
	teamRef := r.teamsCol().Doc(id)
	_, err := teamRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ErrTeamNotFound
		}
		return fmt.Errorf("チームの存在確認に失敗しました: %w", err)
	}

	// questions と各 question の comments を削除
	questionIter := teamRef.Collection("questions").Documents(ctx)
	defer questionIter.Stop()
	for {
		qDoc, err := questionIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("問題一覧の取得に失敗しました: %w", err)
		}
		if err := deleteSubCollection(ctx, r.client, qDoc.Ref.Collection("comments")); err != nil {
			return fmt.Errorf("問題コメントの削除に失敗しました: %w", err)
		}
		if _, err := qDoc.Ref.Delete(ctx); err != nil {
			return fmt.Errorf("問題の削除に失敗しました: %w", err)
		}
	}

	// notes と各 note の comments を削除
	noteIter := teamRef.Collection("notes").Documents(ctx)
	defer noteIter.Stop()
	for {
		nDoc, err := noteIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("ノート一覧の取得に失敗しました: %w", err)
		}
		if err := deleteSubCollection(ctx, r.client, nDoc.Ref.Collection("comments")); err != nil {
			return fmt.Errorf("ノートコメントの削除に失敗しました: %w", err)
		}
		if _, err := nDoc.Ref.Delete(ctx); err != nil {
			return fmt.Errorf("ノートの削除に失敗しました: %w", err)
		}
	}

	// tags を削除
	if err := deleteSubCollection(ctx, r.client, teamRef.Collection("tags")); err != nil {
		return fmt.Errorf("タグの削除に失敗しました: %w", err)
	}

	// members を削除
	if err := deleteSubCollection(ctx, r.client, r.membersCol(id)); err != nil {
		return fmt.Errorf("メンバーの削除に失敗しました: %w", err)
	}

	_, err = teamRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("チームの削除に失敗しました: %w", err)
	}
	return nil
}

// deleteSubCollection はサブコレクション内の全ドキュメントを BulkWriter で削除します。
// BulkWriter は操作をバッチ化して並列送信するため、1件ずつ直列削除より高速です。
// Firestoreのカスケード削除が必要な全リポジトリで共用します。
func deleteSubCollection(ctx context.Context, client *fs.Client, col *fs.CollectionRef) error {
	bw := client.BulkWriter(ctx)

	iter := col.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			bw.End()
			return fmt.Errorf("サブコレクション走査に失敗しました: %w", err)
		}
		if _, err := bw.Delete(doc.Ref); err != nil {
			bw.End()
			return fmt.Errorf("削除ジョブのエンキューに失敗しました: %w", err)
		}
	}
	bw.End()
	return nil
}

// AddMember はチームにメンバーを追加します。
func (r *FirestoreTeamRepository) AddMember(ctx context.Context, member *domain.TeamMember) error {
	memberRef := r.membersCol(member.TeamID).Doc(member.UserID)

	_, err := memberRef.Get(ctx)
	if err == nil {
		return domain.ErrMemberAlreadyExists
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("メンバーの存在確認に失敗しました: %w", err)
	}

	rec := teamMemberRecord{
		TeamID:   member.TeamID,
		UserID:   member.UserID,
		Role:     string(member.Role),
		JoinedAt: member.JoinedAt,
	}
	_, err = memberRef.Set(ctx, rec)
	if err != nil {
		return fmt.Errorf("メンバーの追加に失敗しました: %w", err)
	}
	return nil
}

// RemoveMember はチームからメンバーを除外します。
func (r *FirestoreTeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	memberRef := r.membersCol(teamID).Doc(userID)

	_, err := memberRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ErrMemberNotFound
		}
		return fmt.Errorf("メンバーの存在確認に失敗しました: %w", err)
	}

	_, err = memberRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("メンバーの削除に失敗しました: %w", err)
	}
	return nil
}

// ListMembers はチームのメンバー一覧を返します。
func (r *FirestoreTeamRepository) ListMembers(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	iter := r.membersCol(teamID).Documents(ctx)
	defer iter.Stop()

	var members []*domain.TeamMember
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("メンバー一覧の取得に失敗しました: %w", err)
		}

		var rec teamMemberRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("メンバーデータのデコードに失敗しました: %w", err)
		}
		members = append(members, toTeamMember(rec))
	}

	if members == nil {
		members = []*domain.TeamMember{}
	}
	return members, nil
}

// IsMember はユーザーがチームのメンバーかどうかを返します。
func (r *FirestoreTeamRepository) IsMember(ctx context.Context, teamID, userID string) (bool, error) {
	_, err := r.membersCol(teamID).Doc(userID).Get(ctx)
	if err == nil {
		return true, nil
	}
	if status.Code(err) == codes.NotFound {
		return false, nil
	}
	return false, fmt.Errorf("メンバー確認に失敗しました: %w", err)
}

// FindOwners はチームのオーナーロールを持つメンバー一覧を返します。
func (r *FirestoreTeamRepository) FindOwners(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	iter := r.membersCol(teamID).Where("role", "==", string(domain.MemberRoleOwner)).Documents(ctx)
	defer iter.Stop()

	var owners []*domain.TeamMember
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("オーナー一覧の取得に失敗しました: %w", err)
		}

		var rec teamMemberRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("メンバーデータのデコードに失敗しました: %w", err)
		}
		owners = append(owners, toTeamMember(rec))
	}

	if owners == nil {
		owners = []*domain.TeamMember{}
	}
	return owners, nil
}

// UpdateMemberRole はチームメンバーのロールを変更します。
func (r *FirestoreTeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID string, role domain.MemberRole) error {
	memberRef := r.membersCol(teamID).Doc(userID)

	_, err := memberRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ErrMemberNotFound
		}
		return fmt.Errorf("メンバーの存在確認に失敗しました: %w", err)
	}

	_, err = memberRef.Update(ctx, []fs.Update{
		{Path: "role", Value: string(role)},
	})
	if err != nil {
		return fmt.Errorf("メンバーロールの更新に失敗しました: %w", err)
	}
	return nil
}

// FirestoreTeamRepository が domain.TeamRepository を実装していることをコンパイル時に保証します。
var _ domain.TeamRepository = (*FirestoreTeamRepository)(nil)
