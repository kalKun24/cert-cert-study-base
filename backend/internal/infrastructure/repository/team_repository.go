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

const (
	teamsObjectName   = "teams.json"
	membersObjectName = "team_members.json"
)

// teamRecord はGCS上のJSONファイルに保存するチームレコードです。
type teamRecord struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// teamMemberRecord はGCS上のJSONファイルに保存するメンバーレコードです。
type teamMemberRecord struct {
	TeamID   string    `json:"team_id"`
	UserID   string    `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
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
		JoinedAt: r.JoinedAt,
	}
}

// GCSTeamRepository はGCS上のJSONファイルにチームデータを永続化するリポジトリです。
// domain.TeamRepository インターフェースを実装します。
// チームは teams.json、メンバーは team_members.json に保存します。
type GCSTeamRepository struct {
	mu      sync.RWMutex
	storage storage.StorageClient
	bucket  string
}

// NewGCSTeamRepository は GCSTeamRepository を生成します。
func NewGCSTeamRepository(sc storage.StorageClient, bucket string) *GCSTeamRepository {
	return &GCSTeamRepository{
		storage: sc,
		bucket:  bucket,
	}
}

// loadTeams はGCSからチームデータを読み込みます。
func (r *GCSTeamRepository) loadTeams(ctx context.Context) ([]teamRecord, error) {
	exists, err := r.storage.Exists(ctx, r.bucket, teamsObjectName)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}
	if !exists {
		return []teamRecord{}, nil
	}

	rc, err := r.storage.Read(ctx, r.bucket, teamsObjectName)
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

	var records []teamRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("チームデータのJSONデコードに失敗しました: %w", err)
	}
	return records, nil
}

// saveTeams はチームデータをGCSに書き込みます。
func (r *GCSTeamRepository) saveTeams(ctx context.Context, records []teamRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("チームデータのJSONエンコードに失敗しました: %w", err)
	}
	if err := r.storage.Write(ctx, r.bucket, teamsObjectName, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("GCS への書き込みに失敗しました: %w", err)
	}
	return nil
}

// loadMembers はGCSからメンバーデータを読み込みます。
func (r *GCSTeamRepository) loadMembers(ctx context.Context) ([]teamMemberRecord, error) {
	exists, err := r.storage.Exists(ctx, r.bucket, membersObjectName)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}
	if !exists {
		return []teamMemberRecord{}, nil
	}

	rc, err := r.storage.Read(ctx, r.bucket, membersObjectName)
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

	var records []teamMemberRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("メンバーデータのJSONデコードに失敗しました: %w", err)
	}
	return records, nil
}

// saveMembers はメンバーデータをGCSに書き込みます。
func (r *GCSTeamRepository) saveMembers(ctx context.Context, records []teamMemberRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("メンバーデータのJSONエンコードに失敗しました: %w", err)
	}
	if err := r.storage.Write(ctx, r.bucket, membersObjectName, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("GCS への書き込みに失敗しました: %w", err)
	}
	return nil
}

// FindByID はIDでチームを検索します。
func (r *GCSTeamRepository) FindByID(ctx context.Context, id string) (*domain.Team, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("チームデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.ID == id {
			return toTeam(rec), nil
		}
	}
	return nil, domain.ErrTeamNotFound
}

// FindByName はチーム名でチームを検索します。
func (r *GCSTeamRepository) FindByName(ctx context.Context, name string) (*domain.Team, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("チームデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.Name == name {
			return toTeam(rec), nil
		}
	}
	return nil, domain.ErrTeamNotFound
}

// List は全チームの一覧を返します。
func (r *GCSTeamRepository) List(ctx context.Context) ([]*domain.Team, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("チームデータ読み込みに失敗しました: %w", err)
	}

	teams := make([]*domain.Team, 0, len(records))
	for _, rec := range records {
		teams = append(teams, toTeam(rec))
	}
	return teams, nil
}

// ListByOwnerOrMember はユーザーがオーナーまたはメンバーであるチームを返します。
func (r *GCSTeamRepository) ListByOwnerOrMember(ctx context.Context, userID string) ([]*domain.Team, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	teamRecords, err := r.loadTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("チームデータ読み込みに失敗しました: %w", err)
	}

	memberRecords, err := r.loadMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("メンバーデータ読み込みに失敗しました: %w", err)
	}

	memberTeamIDs := make(map[string]struct{})
	for _, m := range memberRecords {
		if m.UserID == userID {
			memberTeamIDs[m.TeamID] = struct{}{}
		}
	}

	teams := make([]*domain.Team, 0)
	for _, rec := range teamRecords {
		_, isMember := memberTeamIDs[rec.ID]
		if rec.OwnerID == userID || isMember {
			teams = append(teams, toTeam(rec))
		}
	}
	return teams, nil
}

// Save はチームを新規作成または更新します。
func (r *GCSTeamRepository) Save(ctx context.Context, team *domain.Team) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadTeams(ctx)
	if err != nil {
		return fmt.Errorf("チームデータ読み込みに失敗しました: %w", err)
	}

	rec := toTeamRecord(team)
	updated := false
	for i, existing := range records {
		if existing.ID == team.ID {
			records[i] = rec
			updated = true
			break
		}
	}
	if !updated {
		records = append(records, rec)
	}

	if err := r.saveTeams(ctx, records); err != nil {
		return fmt.Errorf("チームデータ保存に失敗しました: %w", err)
	}
	return nil
}

// Delete はIDで指定したチームとそのメンバーを削除します。
func (r *GCSTeamRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadTeams(ctx)
	if err != nil {
		return fmt.Errorf("チームデータ読み込みに失敗しました: %w", err)
	}

	newRecords := make([]teamRecord, 0, len(records))
	found := false
	for _, rec := range records {
		if rec.ID == id {
			found = true
			continue
		}
		newRecords = append(newRecords, rec)
	}

	if !found {
		return domain.ErrTeamNotFound
	}

	if err := r.saveTeams(ctx, newRecords); err != nil {
		return fmt.Errorf("チームデータ保存に失敗しました: %w", err)
	}

	// チームに紐づくメンバーも削除
	memberRecords, err := r.loadMembers(ctx)
	if err != nil {
		return fmt.Errorf("メンバーデータ読み込みに失敗しました: %w", err)
	}

	newMemberRecords := make([]teamMemberRecord, 0, len(memberRecords))
	for _, m := range memberRecords {
		if m.TeamID != id {
			newMemberRecords = append(newMemberRecords, m)
		}
	}

	if err := r.saveMembers(ctx, newMemberRecords); err != nil {
		return fmt.Errorf("メンバーデータ保存に失敗しました: %w", err)
	}

	return nil
}

// AddMember はチームにメンバーを追加します。
func (r *GCSTeamRepository) AddMember(ctx context.Context, member *domain.TeamMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadMembers(ctx)
	if err != nil {
		return fmt.Errorf("メンバーデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.TeamID == member.TeamID && rec.UserID == member.UserID {
			return domain.ErrMemberAlreadyExists
		}
	}

	records = append(records, teamMemberRecord{
		TeamID:   member.TeamID,
		UserID:   member.UserID,
		JoinedAt: member.JoinedAt,
	})

	if err := r.saveMembers(ctx, records); err != nil {
		return fmt.Errorf("メンバーデータ保存に失敗しました: %w", err)
	}
	return nil
}

// RemoveMember はチームからメンバーを除外します。
func (r *GCSTeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadMembers(ctx)
	if err != nil {
		return fmt.Errorf("メンバーデータ読み込みに失敗しました: %w", err)
	}

	newRecords := make([]teamMemberRecord, 0, len(records))
	found := false
	for _, rec := range records {
		if rec.TeamID == teamID && rec.UserID == userID {
			found = true
			continue
		}
		newRecords = append(newRecords, rec)
	}

	if !found {
		return domain.ErrMemberNotFound
	}

	if err := r.saveMembers(ctx, newRecords); err != nil {
		return fmt.Errorf("メンバーデータ保存に失敗しました: %w", err)
	}
	return nil
}

// ListMembers はチームのメンバー一覧を返します。
func (r *GCSTeamRepository) ListMembers(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("メンバーデータ読み込みに失敗しました: %w", err)
	}

	members := make([]*domain.TeamMember, 0)
	for _, rec := range records {
		if rec.TeamID == teamID {
			members = append(members, toTeamMember(rec))
		}
	}
	return members, nil
}

// IsMember はユーザーがチームのメンバーかどうかを返します。
func (r *GCSTeamRepository) IsMember(ctx context.Context, teamID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadMembers(ctx)
	if err != nil {
		return false, fmt.Errorf("メンバーデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.TeamID == teamID && rec.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

// GCSTeamRepository が domain.TeamRepository を実装していることをコンパイル時に保証します。
var _ domain.TeamRepository = (*GCSTeamRepository)(nil)
