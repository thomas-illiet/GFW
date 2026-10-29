package store

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ListTrustedServices returns services with nested hostnames and resolved addresses.
func (s *Store) ListTrustedServices(ctx context.Context) ([]TrustedService, error) {
	var out []TrustedService
	err := s.DB.WithContext(ctx).Table("trusted_services").Order("name").Find(&out).Error
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Hostnames, err = s.hostnames(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// hostnames returns nested hostnames for a trusted service.
func (s *Store) hostnames(ctx context.Context, serviceID string) ([]TrustedHostname, error) {
	var out []TrustedHostname
	err := s.DB.WithContext(ctx).Table("trusted_hostnames").Select("id,hostname,last_attempt_at,last_success_at,last_error,CASE WHEN last_error IS NOT NULL THEN 'stale' WHEN last_success_at IS NULL THEN 'pending' ELSE 'fresh' END status").Where("service_id = ?", serviceID).Order("hostname").Scan(&out).Error
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].IPs, err = s.resolvedIPs(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// resolvedIPs returns addresses previously observed for a hostname.
func (s *Store) resolvedIPs(ctx context.Context, hostnameID string) ([]ResolvedIP, error) {
	var out []ResolvedIP
	err := s.DB.WithContext(ctx).Table("trusted_resolved_ips").Select("id,host(ipv4) ipv4,first_seen_at,last_seen_at,last_used_at,is_current").Where("hostname_id = ?", hostnameID).Order("ipv4").Scan(&out).Error
	return out, err
}

// CreateTrustedService atomically creates a service and its normalized hostnames.
func (s *Store) CreateTrustedService(ctx context.Context, name, description string, active bool, hosts []string) (TrustedService, error) {
	var v TrustedService
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("trusted_services").Create(map[string]any{"name": name, "description": description, "is_active": active}).Error; err != nil {
			return err
		}
		if err := tx.Table("trusted_services").Where("name = ?", name).Take(&v).Error; err != nil {
			return err
		}
		for _, host := range hosts {
			if err := tx.Table("trusted_hostnames").Create(map[string]any{"service_id": v.ID, "hostname": strings.ToLower(host)}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		v.Hostnames, err = s.hostnames(ctx, v.ID)
	}
	return v, err
}

// UpdateTrustedService updates a service and replaces its hostname set.
func (s *Store) UpdateTrustedService(ctx context.Context, id, name, description string, active bool, hosts []string) (TrustedService, error) {
	var v TrustedService
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		db := tx.Table("trusted_services").Where("id = ?", id).Updates(map[string]any{"name": name, "description": description, "is_active": active, "updated_at": time.Now()})
		if db.Error != nil {
			return db.Error
		}
		if db.RowsAffected == 0 {
			return ErrNotFound
		}
		for _, host := range hosts {
			if err := tx.Exec(`INSERT INTO trusted_hostnames(service_id,hostname) VALUES(?,lower(?)) ON CONFLICT DO NOTHING`, id, host).Error; err != nil {
				return err
			}
		}
		deleteQuery := tx.Table("trusted_hostnames").Where("service_id = ?", id)
		if len(hosts) > 0 {
			deleteQuery = deleteQuery.Where("hostname NOT IN ?", hosts)
		}
		return deleteQuery.Delete(map[string]any{}).Error
	})
	if err != nil {
		return v, err
	}
	err = s.DB.WithContext(ctx).Table("trusted_services").Where("id = ?", id).Take(&v).Error
	if err == nil {
		v.Hostnames, err = s.hostnames(ctx, id)
	}
	return v, err
}

// DeleteTrustedService removes a trusted service and its dependent rows.
func (s *Store) DeleteTrustedService(ctx context.Context, id string) error {
	db := s.DB.WithContext(ctx).Table("trusted_services").Where("id = ?", id).Delete(map[string]any{})
	if db.Error == nil && db.RowsAffected == 0 {
		return ErrNotFound
	}
	return db.Error
}

// ListActiveHostnames returns hostnames belonging to active services.
func (s *Store) ListActiveHostnames(ctx context.Context) ([]TrustedHostname, error) {
	var out []TrustedHostname
	err := s.DB.WithContext(ctx).Table("trusted_hostnames h").Select("h.id,h.hostname,h.last_attempt_at,h.last_success_at,h.last_error").Joins("JOIN trusted_services s ON s.id=h.service_id").Where("s.is_active").Order("h.hostname").Scan(&out).Error
	return out, err
}

// SaveResolution atomically stores a DNS success or failure.
func (s *Store) SaveResolution(ctx context.Context, hostnameID string, ips []string, resolveErr error) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if resolveErr != nil {
			return tx.Table("trusted_hostnames").Where("id = ?", hostnameID).Updates(map[string]any{"last_attempt_at": time.Now(), "last_error": resolveErr.Error()}).Error
		}
		if err := tx.Table("trusted_hostnames").Where("id = ?", hostnameID).Updates(map[string]any{"last_attempt_at": time.Now(), "last_success_at": time.Now(), "last_error": nil}).Error; err != nil {
			return err
		}
		if err := tx.Table("trusted_resolved_ips").Where("hostname_id = ?", hostnameID).Update("is_current", false).Error; err != nil {
			return err
		}
		for _, ip := range ips {
			if err := tx.Exec(`INSERT INTO trusted_resolved_ips(hostname_id,ipv4) VALUES(?,?::inet) ON CONFLICT(hostname_id,ipv4) DO UPDATE SET is_current=true,last_seen_at=now()`, hostnameID, ip).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Audit records a best-effort administrative event.
func (s *Store) Audit(ctx context.Context, actor, action, resourceType, resourceID string) {
	_ = s.DB.WithContext(ctx).Exec(`INSERT INTO audit_events(actor,action,resource_type,resource_id) VALUES(?,?,?,NULLIF(?,''))`, actor, action, resourceType, resourceID).Error
}
