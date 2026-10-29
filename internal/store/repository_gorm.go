package store

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"gorm.io/gorm"
)

// ListEcosystems returns ecosystems enriched with machine and usage counters.
func (s *Store) ListEcosystems(ctx context.Context) ([]Ecosystem, error) {
	var out []Ecosystem
	err := s.DB.WithContext(ctx).Raw(`SELECT e.id,e.name,e.display_name,e.is_active,(SELECT count(*) FROM machines m WHERE m.ecosystem_id=e.id) machine_count,(SELECT max(r.occurred_at) FROM firewall_requests r WHERE r.ecosystem_id=e.id OR (r.ecosystem_id IS NULL AND r.ecosystem_name=e.name)) last_used_at,e.created_at,e.updated_at FROM ecosystems e ORDER BY e.name`).Scan(&out).Error
	return out, err
}

// CreateEcosystem persists a new ecosystem.
func (s *Store) CreateEcosystem(ctx context.Context, name, displayName string) (Ecosystem, error) {
	v := Ecosystem{Name: name, DisplayName: displayName, IsActive: true}
	err := s.DB.WithContext(ctx).Omit("MachineCount", "LastUsedAt").Create(&v).Error
	return v, err
}

// UpdateEcosystem updates mutable ecosystem attributes.
func (s *Store) UpdateEcosystem(ctx context.Context, id, displayName string, active bool) (Ecosystem, error) {
	var v Ecosystem
	db := s.DB.WithContext(ctx).Model(&Ecosystem{}).Where("id = ?", id).Updates(map[string]any{"display_name": displayName, "is_active": active, "updated_at": gorm.Expr("now()")})
	if db.Error != nil {
		return v, db.Error
	}
	if db.RowsAffected == 0 {
		return v, ErrNotFound
	}
	err := s.DB.WithContext(ctx).First(&v, "id = ?", id).Error
	return v, err
}

// DeleteEcosystem removes an ecosystem and lets the database cascade to its machines.
func (s *Store) DeleteEcosystem(ctx context.Context, id string) error {
	db := s.DB.WithContext(ctx).Where("id = ?", id).Delete(&Ecosystem{})
	if db.Error == nil && db.RowsAffected == 0 {
		return ErrNotFound
	}
	return db.Error
}

// ListMachines filters machines by text and ecosystem.
func (s *Store) ListMachines(ctx context.Context, search, ecosystemID string) ([]Machine, error) {
	var out []Machine
	q := s.DB.WithContext(ctx).Table("machines m").Select("m.id,m.hostname,m.ecosystem_id,e.name ecosystem_name,host(m.ipv4) ipv4,m.last_used_at,m.last_sync_at,m.is_managed,m.created_at,m.updated_at").Joins("JOIN ecosystems e ON e.id=m.ecosystem_id")
	if search != "" {
		q = q.Where("m.hostname ILIKE ? OR host(m.ipv4) ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if ecosystemID != "" {
		q = q.Where("m.ecosystem_id = ?", ecosystemID)
	}
	err := q.Order("m.hostname").Scan(&out).Error
	return out, err
}

// UpsertMachine creates or updates a machine identified by ecosystem and IPv4.
func (s *Store) UpsertMachine(ctx context.Context, hostname, ecosystemID, ip string, managed bool) (Machine, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is4() {
		return Machine{}, fmt.Errorf("invalid IPv4")
	}
	var v Machine
	err = s.DB.WithContext(ctx).Raw(`INSERT INTO machines(hostname,ecosystem_id,ipv4,is_managed,last_sync_at) VALUES(?,?,?::inet,?,now()) ON CONFLICT(ecosystem_id,ipv4) DO UPDATE SET hostname=excluded.hostname,is_managed=excluded.is_managed,last_sync_at=now(),updated_at=now() RETURNING id,hostname,ecosystem_id,(SELECT name FROM ecosystems WHERE id=ecosystem_id) ecosystem_name,host(ipv4) ipv4,last_used_at,last_sync_at,is_managed,created_at,updated_at`, hostname, ecosystemID, addr.String(), managed).Scan(&v).Error
	return v, err
}

// DeleteMachine removes a manually managed machine.
func (s *Store) DeleteMachine(ctx context.Context, id string) error {
	db := s.DB.WithContext(ctx).Where("id = ? AND is_managed = false", id).Delete(&Machine{})
	if db.Error == nil && db.RowsAffected == 0 {
		return ErrNotFound
	}
	return db.Error
}

// EffectiveIPs returns the effective machine and trusted IPv4 entries for an ecosystem.
func (s *Store) EffectiveIPs(ctx context.Context, ecosystem string) (string, []EffectiveIP, error) {
	var identity struct{ ID, Name string }
	err := s.DB.WithContext(ctx).Table("ecosystems").Select("id,name").Where("(id::text = ? OR name = ?) AND is_active", ecosystem, ecosystem).Take(&identity).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil, ErrNotFound
	}
	if err != nil {
		return "", nil, err
	}
	var out []EffectiveIP
	err = s.DB.WithContext(ctx).Raw(`SELECT host(m.ipv4) ipv4,'machine' source_type,m.id source_id,m.hostname,'' service,m.last_used_at,m.last_sync_at,false stale FROM machines m WHERE m.ecosystem_id=? UNION ALL SELECT host(r.ipv4),'trusted',r.id,h.hostname,s.name,r.last_used_at,h.last_success_at,(h.last_error IS NOT NULL) FROM trusted_resolved_ips r JOIN trusted_hostnames h ON h.id=r.hostname_id JOIN trusted_services s ON s.id=h.service_id WHERE r.is_current AND s.is_active ORDER BY 1,2`, identity.ID).Scan(&out).Error
	return identity.Name, out, err
}

// IPDetails returns all provenance entries for one IPv4 address.
func (s *Store) IPDetails(ctx context.Context, ip string) ([]EffectiveIP, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is4() {
		return nil, fmt.Errorf("invalid IPv4")
	}
	var out []EffectiveIP
	err = s.DB.WithContext(ctx).Raw(`SELECT host(m.ipv4) ipv4,'machine' source_type,m.id source_id,m.hostname,'' service,m.last_used_at,m.last_sync_at,false stale FROM machines m WHERE m.ipv4=?::inet UNION ALL SELECT host(r.ipv4),'trusted',r.id,h.hostname,s.name,r.last_used_at,h.last_success_at,(h.last_error IS NOT NULL) FROM trusted_resolved_ips r JOIN trusted_hostnames h ON h.id=r.hostname_id JOIN trusted_services s ON s.id=h.service_id WHERE r.is_current AND s.is_active AND r.ipv4=?::inet`, addr.String(), addr.String()).Scan(&out).Error
	if err == nil && len(out) == 0 {
		err = ErrNotFound
	}
	return out, err
}

// RecordRequest atomically records a firewall response and its individual IP entries.
func (s *Store) RecordRequest(ctx context.Context, input HistoryInput) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		machine, trusted := 0, 0
		for _, entry := range input.Entries {
			if entry.SourceType == "machine" {
				machine++
			} else {
				trusted++
			}
		}
		var record struct {
			ID         string
			OccurredAt time.Time
		}
		err := tx.Raw(`INSERT INTO firewall_requests(route_kind,ecosystem_id,ecosystem_name,format,status_code,duration_ms,etag,not_modified,machine_count,trusted_count,total_count) VALUES(?,COALESCE(NULLIF(?,'')::uuid,(SELECT id FROM ecosystems WHERE name=NULLIF(?,''))),NULLIF(?,''),?,?,?,?,?,?,?,?) RETURNING id,occurred_at`, input.RouteKind, input.EcosystemID, input.EcosystemName, input.EcosystemName, input.Format, input.StatusCode, input.DurationMS, input.ETag, input.NotModified, machine, trusted, len(input.Entries)).Scan(&record).Error
		if err != nil {
			return err
		}
		for _, entry := range input.Entries {
			if err := tx.Exec(`INSERT INTO firewall_request_ips(request_id,request_occurred_at,ipv4,source_type,source_id) VALUES(?,?,?::inet,?,?)`, record.ID, record.OccurredAt, entry.IPv4, entry.SourceType, entry.SourceID).Error; err != nil {
				return err
			}
			table := "machines"
			if entry.SourceType == "trusted" {
				table = "trusted_resolved_ips"
			}
			if err := tx.Table(table).Where("id = ?", entry.SourceID).Update("last_used_at", record.OccurredAt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
