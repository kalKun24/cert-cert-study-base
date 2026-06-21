// Package repository はドメインリポジトリインターフェースの具体実装を提供します。
package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/storage"
)

// notesObjectName はチームIDをもとにGCSオブジェクト名を返します。
// パス: teams/{teamID}/notes.json
func notesObjectName(teamID string) string {
	return fmt.Sprintf("teams/%s/notes.json", teamID)
}

// noteRecord はGCS上のJSONファイルに保存するノートレコードです。
// domain.Note と対応しており、JSON直列化のための構造体です。
type noteRecord struct {
	ID               string    `json:"id"`
	TeamID           string    `json:"team_id"`
	Title            string    `json:"title"`
	Body             string    `json:"body"`
	DiscussionPoints string    `json:"discussion_points"`
	Memo             string    `json:"memo"`
	Tags             []string  `json:"tags"`
	Status           string    `json:"status"`
	CreatedBy        string    `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// toNoteRecord はドメインエンティティをJSONレコードに変換します。
func toNoteRecord(n *domain.Note) noteRecord {
	tags := n.Tags
	if tags == nil {
		tags = []string{}
	}
	return noteRecord{
		ID:               n.ID,
		TeamID:           n.TeamID,
		Title:            n.Title,
		Body:             n.Body,
		DiscussionPoints: n.DiscussionPoints,
		Memo:             n.Memo,
		Tags:             tags,
		Status:           string(n.Status),
		CreatedBy:        n.CreatedBy,
		CreatedAt:        n.CreatedAt,
		UpdatedAt:        n.UpdatedAt,
	}
}

// toNote はJSONレコードをドメインエンティティに変換します。
// 後方互換性のため、status が空の場合はデフォルト値を設定します。
func toNote(r noteRecord) *domain.Note {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}

	// 後方互換: status フィールドが空の場合はデフォルト値を設定
	status := domain.NoteStatus(r.Status)
	if status == "" {
		status = domain.NoteStatusDraft
	}

	return &domain.Note{
		ID:               r.ID,
		TeamID:           r.TeamID,
		Title:            r.Title,
		Body:             r.Body,
		DiscussionPoints: r.DiscussionPoints,
		Memo:             r.Memo,
		Tags:             tags,
		Status:           status,
		CreatedBy:        r.CreatedBy,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

// GCSNoteRepository はGCS上のJSONファイルにノートデータを永続化するリポジトリです。
// domain.NoteRepository インターフェースを実装します。
// ノートデータはチームごとに GCS バケット内の teams/{teamID}/notes.json に保存します。
type GCSNoteRepository struct {
	mu      sync.RWMutex
	storage storage.StorageClient
	bucket  string
}

// NewGCSNoteRepository は GCSNoteRepository を生成します。
func NewGCSNoteRepository(sc storage.StorageClient, bucket string) *GCSNoteRepository {
	return &GCSNoteRepository{
		storage: sc,
		bucket:  bucket,
	}
}

// loadNotes はGCSから指定チームのノートデータを読み込みます。
// オブジェクトが存在しない場合は空のスライスを返します。
func (r *GCSNoteRepository) loadNotes(ctx context.Context, teamID string) ([]noteRecord, error) {
	objectName := notesObjectName(teamID)

	exists, err := r.storage.Exists(ctx, r.bucket, objectName)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}

	if !exists {
		return []noteRecord{}, nil
	}

	rc, err := r.storage.Read(ctx, r.bucket, objectName)
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

	var records []noteRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("ノートデータのJSONデコードに失敗しました: %w", err)
	}

	return records, nil
}

// saveNotes は指定チームのノートデータをGCSに書き込みます。
func (r *GCSNoteRepository) saveNotes(ctx context.Context, teamID string, records []noteRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("ノートデータのJSONエンコードに失敗しました: %w", err)
	}

	if err := r.storage.Write(ctx, r.bucket, notesObjectName(teamID), bytes.NewReader(data)); err != nil {
		return fmt.Errorf("GCS への書き込みに失敗しました: %w", err)
	}

	return nil
}

// FindByID はチームIDとIDでノートを検索します。
func (r *GCSNoteRepository) FindByID(ctx context.Context, teamID, id string) (*domain.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadNotes(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("ノートデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.ID == id {
			return toNote(rec), nil
		}
	}

	return nil, domain.ErrNoteNotFound
}

// ListByTeam は指定チームのノート一覧を返します。
// チーム別ファイルを直接読み込むため、アプリ側フィルタリングは不要です。
func (r *GCSNoteRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadNotes(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("ノートデータ読み込みに失敗しました: %w", err)
	}

	notes := make([]*domain.Note, 0, len(records))
	for _, rec := range records {
		notes = append(notes, toNote(rec))
	}

	return notes, nil
}

// SearchByTeam は指定チームのノートを検索・フィルタリングして返します。
// チーム別ファイルを直接読み込むため、teamID フィルタは不要です。
func (r *GCSNoteRepository) SearchByTeam(ctx context.Context, teamID string, filter domain.NoteSearchFilter) ([]*domain.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadNotes(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("ノートデータ読み込みに失敗しました: %w", err)
	}

	notes := make([]*domain.Note, 0)
	for _, rec := range records {
		n := toNote(rec)

		// タグANDフィルタリング: 指定されたタグIDをすべて持つノートのみ返す
		if len(filter.TagIDs) > 0 {
			tagSet := make(map[string]struct{}, len(rec.Tags))
			for _, t := range rec.Tags {
				tagSet[t] = struct{}{}
			}
			match := true
			for _, tid := range filter.TagIDs {
				if _, ok := tagSet[tid]; !ok {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		// キーワード検索: title / body / discussion_points / memo を対象とした部分一致
		if filter.Keyword != "" {
			kw := filter.Keyword
			if !strings.Contains(rec.Title, kw) &&
				!strings.Contains(rec.Body, kw) &&
				!strings.Contains(rec.DiscussionPoints, kw) &&
				!strings.Contains(rec.Memo, kw) {
				continue
			}
		}

		notes = append(notes, n)
	}

	return notes, nil
}

// FindByTagID は指定チームの指定タグIDを持つノートの一覧を返します。
func (r *GCSNoteRepository) FindByTagID(ctx context.Context, teamID, tagID string) ([]*domain.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadNotes(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("ノートデータ読み込みに失敗しました: %w", err)
	}

	var notes []*domain.Note
	for _, rec := range records {
		for _, tid := range rec.Tags {
			if tid == tagID {
				notes = append(notes, toNote(rec))
				break
			}
		}
	}

	return notes, nil
}

// Save はノートを新規作成または更新します。
// IDが一致するレコードが存在する場合は更新、存在しない場合は追加します。
func (r *GCSNoteRepository) Save(ctx context.Context, note *domain.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadNotes(ctx, note.TeamID)
	if err != nil {
		return fmt.Errorf("ノートデータ読み込みに失敗しました: %w", err)
	}

	rec := toNoteRecord(note)
	updated := false
	for i, existing := range records {
		if existing.ID == note.ID {
			records[i] = rec
			updated = true
			break
		}
	}

	if !updated {
		records = append(records, rec)
	}

	if err := r.saveNotes(ctx, note.TeamID, records); err != nil {
		return fmt.Errorf("ノートデータ保存に失敗しました: %w", err)
	}

	return nil
}

// Delete はチームIDとIDで指定したノートを削除します。
// ノートが存在しない場合は ErrNoteNotFound を返します。
func (r *GCSNoteRepository) Delete(ctx context.Context, teamID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadNotes(ctx, teamID)
	if err != nil {
		return fmt.Errorf("ノートデータ読み込みに失敗しました: %w", err)
	}

	newRecords := make([]noteRecord, 0, len(records))
	found := false
	for _, rec := range records {
		if rec.ID == id {
			found = true
			continue
		}
		newRecords = append(newRecords, rec)
	}

	if !found {
		return domain.ErrNoteNotFound
	}

	if err := r.saveNotes(ctx, teamID, newRecords); err != nil {
		return fmt.Errorf("ノートデータ保存に失敗しました: %w", err)
	}

	return nil
}

// GCSNoteRepository が domain.NoteRepository を実装していることをコンパイル時に保証します。
var _ domain.NoteRepository = (*GCSNoteRepository)(nil)
