// Package repository はドメインリポジトリインターフェースの具体実装を提供します。
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

// noteCommentObjectName はGCS上の個別ノートコメントオブジェクト名を返します。
// パス: teams/{teamID}/notes/{noteID}/comments/{commentID}.json
func noteCommentObjectName(teamID, noteID, commentID string) string {
	return fmt.Sprintf("teams/%s/notes/%s/comments/%s.json", teamID, noteID, commentID)
}

// noteCommentPrefix は指定したチームIDとノートIDのコメント一覧取得用プレフィックスです。
func noteCommentPrefix(teamID, noteID string) string {
	return fmt.Sprintf("teams/%s/notes/%s/comments/", teamID, noteID)
}

// noteCommentRecord はGCS上のJSONファイルに保存するノートコメントレコードです。
// domain.NoteComment と対応しており、JSON直列化のための構造体です。
type noteCommentRecord struct {
	ID        string    `json:"id"`
	NoteID    string    `json:"note_id"`
	Body      string    `json:"body"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// toNoteCommentRecord はドメインエンティティをJSONレコードに変換します。
func toNoteCommentRecord(c *domain.NoteComment) noteCommentRecord {
	return noteCommentRecord{
		ID:        c.ID,
		NoteID:    c.NoteID,
		Body:      c.Body,
		CreatedBy: c.CreatedBy,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// toNoteComment はJSONレコードをドメインエンティティに変換します。
func toNoteComment(r noteCommentRecord) *domain.NoteComment {
	return &domain.NoteComment{
		ID:        r.ID,
		NoteID:    r.NoteID,
		Body:      r.Body,
		CreatedBy: r.CreatedBy,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// GCSNoteCommentRepository はGCS上のJSONファイルにノートコメントデータを永続化するリポジトリです。
// domain.NoteCommentRepository インターフェースを実装します。
// コメントは1件1ファイル形式で保存します。
// パス: teams/{teamID}/notes/{noteID}/comments/{commentID}.json
type GCSNoteCommentRepository struct {
	mu      sync.RWMutex
	storage storage.StorageClient
	bucket  string
}

// NewGCSNoteCommentRepository は GCSNoteCommentRepository を生成します。
func NewGCSNoteCommentRepository(sc storage.StorageClient, bucket string) *GCSNoteCommentRepository {
	return &GCSNoteCommentRepository{
		storage: sc,
		bucket:  bucket,
	}
}

// loadNoteComment はGCSから指定したノートコメントを読み込みます。
func (r *GCSNoteCommentRepository) loadNoteComment(ctx context.Context, teamID, noteID, commentID string) (noteCommentRecord, error) {
	objectName := noteCommentObjectName(teamID, noteID, commentID)

	exists, err := r.storage.Exists(ctx, r.bucket, objectName)
	if err != nil {
		return noteCommentRecord{}, fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}
	if !exists {
		return noteCommentRecord{}, domain.ErrNoteCommentNotFound
	}

	rc, err := r.storage.Read(ctx, r.bucket, objectName)
	if err != nil {
		return noteCommentRecord{}, fmt.Errorf("GCS からの読み込みに失敗しました: %w", err)
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			slog.Warn("GCS ReadCloser のクローズに失敗しました", "error", cerr)
		}
	}()

	data, err := io.ReadAll(rc)
	if err != nil {
		return noteCommentRecord{}, fmt.Errorf("GCS データの読み取りに失敗しました: %w", err)
	}

	var rec noteCommentRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return noteCommentRecord{}, fmt.Errorf("ノートコメントデータのJSONデコードに失敗しました: %w", err)
	}

	return rec, nil
}

// saveNoteComment はノートコメントをGCSに書き込みます。
func (r *GCSNoteCommentRepository) saveNoteComment(ctx context.Context, teamID string, rec noteCommentRecord) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("ノートコメントデータのJSONエンコードに失敗しました: %w", err)
	}

	objectName := noteCommentObjectName(teamID, rec.NoteID, rec.ID)
	if err := r.storage.Write(ctx, r.bucket, objectName, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("GCS への書き込みに失敗しました: %w", err)
	}

	return nil
}

// FindByID はIDでノートコメントを検索します。
func (r *GCSNoteCommentRepository) FindByID(ctx context.Context, teamID, noteID, commentID string) (*domain.NoteComment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rec, err := r.loadNoteComment(ctx, teamID, noteID, commentID)
	if err != nil {
		return nil, err
	}

	return toNoteComment(rec), nil
}

// ListByNoteID は指定したチームIDとノートIDのコメント一覧を返します。
// ソートはユースケース層で行うため、ここでは順序を保証しません。
func (r *GCSNoteCommentRepository) ListByNoteID(ctx context.Context, teamID, noteID string) ([]*domain.NoteComment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := noteCommentPrefix(teamID, noteID)
	objectNames, err := r.storage.List(ctx, r.bucket, prefix)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクト一覧の取得に失敗しました: %w", err)
	}

	comments := make([]*domain.NoteComment, 0, len(objectNames))
	for _, objectName := range objectNames {
		rc, err := r.storage.Read(ctx, r.bucket, objectName)
		if err != nil {
			return nil, fmt.Errorf("GCS からの読み込みに失敗しました (%s): %w", objectName, err)
		}

		data, err := func() ([]byte, error) {
			defer func() {
				if cerr := rc.Close(); cerr != nil {
					slog.Warn("GCS ReadCloser のクローズに失敗しました", "error", cerr)
				}
			}()
			return io.ReadAll(rc)
		}()
		if err != nil {
			return nil, fmt.Errorf("GCS データの読み取りに失敗しました (%s): %w", objectName, err)
		}

		var rec noteCommentRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil, fmt.Errorf("ノートコメントデータのJSONデコードに失敗しました (%s): %w", objectName, err)
		}

		comments = append(comments, toNoteComment(rec))
	}

	return comments, nil
}

// Save はノートコメントを新規作成または更新します。
func (r *GCSNoteCommentRepository) Save(ctx context.Context, teamID string, comment *domain.NoteComment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec := toNoteCommentRecord(comment)
	if err := r.saveNoteComment(ctx, teamID, rec); err != nil {
		return fmt.Errorf("ノートコメントデータ保存に失敗しました: %w", err)
	}

	return nil
}

// Delete はIDで指定したノートコメントを削除します。
func (r *GCSNoteCommentRepository) Delete(ctx context.Context, teamID, noteID, commentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	objectName := noteCommentObjectName(teamID, noteID, commentID)

	exists, err := r.storage.Exists(ctx, r.bucket, objectName)
	if err != nil {
		return fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}
	if !exists {
		return domain.ErrNoteCommentNotFound
	}

	if err := r.storage.Delete(ctx, r.bucket, objectName); err != nil {
		return fmt.Errorf("GCS からの削除に失敗しました: %w", err)
	}

	return nil
}

// GCSNoteCommentRepository が domain.NoteCommentRepository を実装していることをコンパイル時に保証します。
var _ domain.NoteCommentRepository = (*GCSNoteCommentRepository)(nil)
