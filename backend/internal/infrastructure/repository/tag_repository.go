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

// tagsObjectName はGCSバケット内でタグデータを保存するオブジェクト名です。
const tagsObjectName = "tags.json"

// tagRecord はGCS上のJSONファイルに保存するタグレコードです。
// domain.Tag と対応しており、JSON直列化のための構造体です。
type tagRecord struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// toTagRecord はドメインエンティティをJSONレコードに変換します。
func toTagRecord(t *domain.Tag) tagRecord {
	return tagRecord{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}
}

// toTag はJSONレコードをドメインエンティティに変換します。
func toTag(r tagRecord) *domain.Tag {
	return &domain.Tag{
		ID:        r.ID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
	}
}

// GCSTagRepository はGCS上のJSONファイルにタグデータを永続化するリポジトリです。
// domain.TagRepository インターフェースを実装します。
// タグデータは GCS バケット内の tags.json に保存します。
type GCSTagRepository struct {
	mu      sync.RWMutex
	storage storage.StorageClient
	bucket  string
}

// NewGCSTagRepository は GCSTagRepository を生成します。
func NewGCSTagRepository(sc storage.StorageClient, bucket string) *GCSTagRepository {
	return &GCSTagRepository{
		storage: sc,
		bucket:  bucket,
	}
}

// loadTags はGCSからタグデータを読み込みます。
// オブジェクトが存在しない場合は空のスライスを返します。
func (r *GCSTagRepository) loadTags(ctx context.Context) ([]tagRecord, error) {
	exists, err := r.storage.Exists(ctx, r.bucket, tagsObjectName)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}

	if !exists {
		return []tagRecord{}, nil
	}

	rc, err := r.storage.Read(ctx, r.bucket, tagsObjectName)
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

	var records []tagRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("タグデータのJSONデコードに失敗しました: %w", err)
	}

	return records, nil
}

// saveTags はタグデータをGCSに書き込みます。
func (r *GCSTagRepository) saveTags(ctx context.Context, records []tagRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("タグデータのJSONエンコードに失敗しました: %w", err)
	}

	if err := r.storage.Write(ctx, r.bucket, tagsObjectName, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("GCS への書き込みに失敗しました: %w", err)
	}

	return nil
}

// FindByID はIDでタグを検索します。
func (r *GCSTagRepository) FindByID(ctx context.Context, id string) (*domain.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("タグデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.ID == id {
			return toTag(rec), nil
		}
	}

	return nil, domain.ErrTagNotFound
}

// FindByName はタグ名でタグを検索します。
func (r *GCSTagRepository) FindByName(ctx context.Context, name string) (*domain.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("タグデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.Name == name {
			return toTag(rec), nil
		}
	}

	return nil, domain.ErrTagNotFound
}

// List は全タグを返します。
func (r *GCSTagRepository) List(ctx context.Context) ([]*domain.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("タグデータ読み込みに失敗しました: %w", err)
	}

	tags := make([]*domain.Tag, 0, len(records))
	for _, rec := range records {
		tags = append(tags, toTag(rec))
	}

	return tags, nil
}

// Save はタグを新規作成または更新します。
// IDが一致するレコードが存在する場合は更新、存在しない場合は追加します。
func (r *GCSTagRepository) Save(ctx context.Context, tag *domain.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadTags(ctx)
	if err != nil {
		return fmt.Errorf("タグデータ読み込みに失敗しました: %w", err)
	}

	rec := toTagRecord(tag)
	updated := false
	for i, existing := range records {
		if existing.ID == tag.ID {
			records[i] = rec
			updated = true
			break
		}
	}

	if !updated {
		records = append(records, rec)
	}

	if err := r.saveTags(ctx, records); err != nil {
		return fmt.Errorf("タグデータ保存に失敗しました: %w", err)
	}

	return nil
}

// Delete はIDで指定したタグを削除します。
func (r *GCSTagRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadTags(ctx)
	if err != nil {
		return fmt.Errorf("タグデータ読み込みに失敗しました: %w", err)
	}

	newRecords := make([]tagRecord, 0, len(records))
	found := false
	for _, rec := range records {
		if rec.ID == id {
			found = true
			continue
		}
		newRecords = append(newRecords, rec)
	}

	if !found {
		return domain.ErrTagNotFound
	}

	if err := r.saveTags(ctx, newRecords); err != nil {
		return fmt.Errorf("タグデータ保存に失敗しました: %w", err)
	}

	return nil
}

// GCSTagRepository が domain.TagRepository を実装していることをコンパイル時に保証します。
var _ domain.TagRepository = (*GCSTagRepository)(nil)
