package store

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ListIngestionKeys returns API-key metadata without secret hashes.
func (s *Store) ListIngestionKeys(ctx context.Context) ([]IngestionKey, error) {
	var out []IngestionKey
	err := s.DB.WithContext(ctx).Order("created_at DESC").Find(&out).Error
	return out, err
}

// CreateIngestionKey stores a hashed ingestion key.
func (s *Store) CreateIngestionKey(ctx context.Context, name, prefix, hash string, expiresAt *time.Time) (IngestionKey, error) {
	row := map[string]any{"name": name, "prefix": prefix, "secret_hash": hash, "expires_at": expiresAt}
	err := s.DB.WithContext(ctx).Table("ingestion_keys").Create(row).Error
	var v IngestionKey
	if err == nil {
		err = s.DB.WithContext(ctx).Where("prefix = ?", prefix).First(&v).Error
	}
	return v, err
}

// RevokeIngestionKey marks an active key as revoked.
func (s *Store) RevokeIngestionKey(ctx context.Context, id string) error {
	db := s.DB.WithContext(ctx).Table("ingestion_keys").Where("id = ? AND revoked_at IS NULL", id).Update("revoked_at", time.Now())
	if db.Error == nil && db.RowsAffected == 0 {
		return ErrNotFound
	}
	return db.Error
}

// IngestionKeyHash returns the identifier and hash for a currently valid prefix.
func (s *Store) IngestionKeyHash(ctx context.Context, prefix string) (string, string, error) {
	var row struct{ ID, SecretHash string }
	err := s.DB.WithContext(ctx).Table("ingestion_keys").Select("id,secret_hash").Where("prefix = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > now())", prefix).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrNotFound
	}
	return row.ID, row.SecretHash, err
}

// TouchIngestionKey records successful use of an ingestion key.
func (s *Store) TouchIngestionKey(ctx context.Context, id string) {
	_ = s.DB.WithContext(ctx).Table("ingestion_keys").Where("id = ?", id).Update("last_used_at", time.Now()).Error
}
