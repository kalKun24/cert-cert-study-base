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

// commentObjectName はGCS上の個別コメントオブジェクト名を返します。
// パス: questions/{questionID}/comments/{commentID}.json
func commentObjectName(questionID, commentID string) string {
	return fmt.Sprintf("questions/%s/comments/%s.json", questionID, commentID)
}

// commentPrefixByQuestion は指定した問題IDのコメント一覧取得用プレフィックスです。
func commentPrefixByQuestion(questionID string) string {
	return fmt.Sprintf("questions/%s/comments/", questionID)
}

// commentRecord はGCS上のJSONファイルに保存するコメントレコードです。
// domain.Comment と対応しており、JSON直列化のための構造体です。
type commentRecord struct {
	ID         string    `json:"id"`
	QuestionID string    `json:"question_id"`
	Body       string    `json:"body"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// toCommentRecord はドメインエンティティをJSONレコードに変換します。
func toCommentRecord(c *domain.Comment) commentRecord {
	return commentRecord{
		ID:         c.ID,
		QuestionID: c.QuestionID,
		Body:       c.Body,
		CreatedBy:  c.CreatedBy,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

// toComment はJSONレコードをドメインエンティティに変換します。
func toComment(r commentRecord) *domain.Comment {
	return &domain.Comment{
		ID:         r.ID,
		QuestionID: r.QuestionID,
		Body:       r.Body,
		CreatedBy:  r.CreatedBy,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// GCSCommentRepository はGCS上のJSONファイルにコメントデータを永続化するリポジトリです。
// domain.CommentRepository インターフェースを実装します。
// コメントは1件1ファイル形式で保存します。
// パス: questions/{questionID}/comments/{commentID}.json
type GCSCommentRepository struct {
	mu      sync.RWMutex
	storage storage.StorageClient
	bucket  string
}

// NewGCSCommentRepository は GCSCommentRepository を生成します。
func NewGCSCommentRepository(sc storage.StorageClient, bucket string) *GCSCommentRepository {
	return &GCSCommentRepository{
		storage: sc,
		bucket:  bucket,
	}
}

// loadComment はGCSから指定したコメントを読み込みます。
func (r *GCSCommentRepository) loadComment(ctx context.Context, questionID, commentID string) (commentRecord, error) {
	objectName := commentObjectName(questionID, commentID)

	exists, err := r.storage.Exists(ctx, r.bucket, objectName)
	if err != nil {
		return commentRecord{}, fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}
	if !exists {
		return commentRecord{}, domain.ErrCommentNotFound
	}

	rc, err := r.storage.Read(ctx, r.bucket, objectName)
	if err != nil {
		return commentRecord{}, fmt.Errorf("GCS からの読み込みに失敗しました: %w", err)
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			slog.Warn("GCS ReadCloser のクローズに失敗しました", "error", cerr)
		}
	}()

	data, err := io.ReadAll(rc)
	if err != nil {
		return commentRecord{}, fmt.Errorf("GCS データの読み取りに失敗しました: %w", err)
	}

	var rec commentRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return commentRecord{}, fmt.Errorf("コメントデータのJSONデコードに失敗しました: %w", err)
	}

	return rec, nil
}

// saveComment はコメントをGCSに書き込みます。
func (r *GCSCommentRepository) saveComment(ctx context.Context, rec commentRecord) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("コメントデータのJSONエンコードに失敗しました: %w", err)
	}

	objectName := commentObjectName(rec.QuestionID, rec.ID)
	if err := r.storage.Write(ctx, r.bucket, objectName, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("GCS への書き込みに失敗しました: %w", err)
	}

	return nil
}

// FindByID はIDでコメントを検索します。
func (r *GCSCommentRepository) FindByID(ctx context.Context, questionID, commentID string) (*domain.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rec, err := r.loadComment(ctx, questionID, commentID)
	if err != nil {
		return nil, err
	}

	return toComment(rec), nil
}

// ListByQuestionID は指定した問題IDのコメント一覧を返します。
// ソートはユースケース層で行うため、ここでは順序を保証しません。
func (r *GCSCommentRepository) ListByQuestionID(ctx context.Context, questionID string) ([]*domain.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := commentPrefixByQuestion(questionID)
	objectNames, err := r.storage.List(ctx, r.bucket, prefix)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクト一覧の取得に失敗しました: %w", err)
	}

	comments := make([]*domain.Comment, 0, len(objectNames))
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

		var rec commentRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil, fmt.Errorf("コメントデータのJSONデコードに失敗しました (%s): %w", objectName, err)
		}

		comments = append(comments, toComment(rec))
	}

	return comments, nil
}

// Save はコメントを新規作成または更新します。
func (r *GCSCommentRepository) Save(ctx context.Context, comment *domain.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rec := toCommentRecord(comment)
	if err := r.saveComment(ctx, rec); err != nil {
		return fmt.Errorf("コメントデータ保存に失敗しました: %w", err)
	}

	return nil
}

// Delete はIDで指定したコメントを削除します。
func (r *GCSCommentRepository) Delete(ctx context.Context, questionID, commentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	objectName := commentObjectName(questionID, commentID)

	exists, err := r.storage.Exists(ctx, r.bucket, objectName)
	if err != nil {
		return fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}
	if !exists {
		return domain.ErrCommentNotFound
	}

	if err := r.storage.Delete(ctx, r.bucket, objectName); err != nil {
		return fmt.Errorf("GCS からの削除に失敗しました: %w", err)
	}

	return nil
}

// GCSCommentRepository が domain.CommentRepository を実装していることをコンパイル時に保証します。
var _ domain.CommentRepository = (*GCSCommentRepository)(nil)
