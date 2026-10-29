// Package store implements PostgreSQL persistence through GORM.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ErrNotFound identifies a missing domain resource.
var ErrNotFound = errors.New("not found")

// Store provides domain repositories backed by GORM.
type Store struct{ DB *gorm.DB }

// Open connects to PostgreSQL without mutating its schema.
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), TranslateError: true})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return &Store{DB: db}, nil
}

// New returns a store around an existing GORM connection.
func New(db *gorm.DB) *Store { return &Store{DB: db} }

// Close releases the underlying SQL connection pool.
func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}
	db, err := s.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

// Ping verifies that the database is reachable.
func (s *Store) Ping(ctx context.Context) error {
	db, err := s.DB.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

// SQLDB exposes the shared database/sql pool to infrastructure adapters.
func (s *Store) SQLDB() (*sql.DB, error) { return s.DB.DB() }

// EnsureHistoryPartitions creates monthly request partitions around the current date.
func (s *Store) EnsureHistoryPartitions(ctx context.Context, now time.Time, monthsAhead int) error {
	start := time.Date(now.Year(), now.Month()-12, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i <= 12+monthsAhead; i++ {
		from, to := start.AddDate(0, i, 0), start.AddDate(0, i+1, 0)
		suffix := from.Format("2006_01")
		queries := []string{
			fmt.Sprintf(`CREATE TABLE IF NOT EXISTS firewall_requests_%s PARTITION OF firewall_requests FOR VALUES FROM ('%s') TO ('%s')`, suffix, from.Format(time.RFC3339), to.Format(time.RFC3339)),
			fmt.Sprintf(`CREATE TABLE IF NOT EXISTS firewall_request_ips_%s PARTITION OF firewall_request_ips FOR VALUES FROM ('%s') TO ('%s')`, suffix, from.Format(time.RFC3339), to.Format(time.RFC3339)),
			fmt.Sprintf(`CREATE INDEX IF NOT EXISTS firewall_requests_%s_time_idx ON firewall_requests_%s(occurred_at, ecosystem_id, format, status_code)`, suffix, suffix),
			fmt.Sprintf(`CREATE INDEX IF NOT EXISTS firewall_request_ips_%s_lookup_idx ON firewall_request_ips_%s(request_occurred_at, source_type, ipv4)`, suffix, suffix),
		}
		for _, query := range queries {
			if err := s.DB.WithContext(ctx).Exec(query).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// DropExpiredPartitions removes monthly history partitions older than retentionDays.
func (s *Store) DropExpiredPartitions(ctx context.Context, now time.Time, retentionDays int) error {
	var names []string
	err := s.DB.WithContext(ctx).Raw(`SELECT inhrelid::regclass::text FROM pg_inherits WHERE inhparent IN ('firewall_requests'::regclass,'firewall_request_ips'::regclass)`).Scan(&names).Error
	if err != nil {
		return err
	}
	cutoff := now.AddDate(0, 0, -retentionDays)
	for _, name := range names {
		parts := strings.Split(name, "_")
		if len(parts) < 3 {
			continue
		}
		month, err := time.Parse("2006_01", strings.Join(parts[len(parts)-2:], "_"))
		if err == nil && month.AddDate(0, 1, 0).Before(cutoff) {
			if err := s.DB.WithContext(ctx).Exec(`DROP TABLE IF EXISTS ` + name).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
