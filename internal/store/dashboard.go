package store

import (
	"context"
	"fmt"
	"time"
)

// Dashboard returns request and address activity aggregates for a supported window.
func (s *Store) Dashboard(ctx context.Context, window string) (Dashboard, error) {
	duration := 24 * time.Hour
	bucket := "hour"
	bucketCount := 24
	if window == "7d" {
		duration = 7 * 24 * time.Hour
		bucket = "day"
		bucketCount = 7
	} else if window == "30d" {
		duration = 30 * 24 * time.Hour
		bucket = "day"
		bucketCount = 30
	}
	now := time.Now().UTC()
	start := now.Add(-duration)
	previous := start.Add(-duration)
	d := Dashboard{Window: window}
	var success, errorCount int64
	err := s.DB.WithContext(ctx).Raw(`SELECT count(*) FILTER(WHERE occurred_at>=?),count(*) FILTER(WHERE occurred_at>=? AND occurred_at<?),count(*) FILTER(WHERE occurred_at>=? AND status_code<400),count(*) FILTER(WHERE occurred_at>=? AND status_code>=400),count(*) FILTER(WHERE occurred_at>=? AND not_modified),coalesce(avg(duration_ms) FILTER(WHERE occurred_at>=?),0),coalesce(percentile_cont(.95) WITHIN GROUP(ORDER BY duration_ms) FILTER(WHERE occurred_at>=?),0) FROM firewall_requests WHERE occurred_at>=?`, start, previous, start, start, start, start, start, start, previous).Row().Scan(&d.Requests.Total, &d.Requests.PreviousTotal, &success, &errorCount, &d.Requests.NotModified, &d.Requests.AverageLatencyMS, &d.Requests.P95LatencyMS)
	if err != nil {
		return d, err
	}
	if d.Requests.Total > 0 {
		d.Requests.SuccessRate = float64(success) * 100 / float64(d.Requests.Total)
		d.Requests.ErrorRate = float64(errorCount) * 100 / float64(d.Requests.Total)
	}
	if d.Requests.PreviousTotal > 0 {
		d.Requests.ChangePercent = float64(d.Requests.Total-d.Requests.PreviousTotal) * 100 / float64(d.Requests.PreviousTotal)
	}
	err = s.DB.WithContext(ctx).Raw(`WITH current_ips AS (SELECT host(ipv4) ip,'machine' source FROM machines UNION ALL SELECT host(r.ipv4),'trusted' FROM trusted_resolved_ips r JOIN trusted_hostnames h ON h.id=r.hostname_id JOIN trusted_services s ON s.id=h.service_id WHERE r.is_current AND s.is_active), active AS (SELECT DISTINCT host(ipv4) ip,source_type source FROM firewall_request_ips WHERE request_occurred_at >= ?) SELECT count(DISTINCT c.ip),count(DISTINCT c.ip) FILTER(WHERE a.ip IS NOT NULL),count(DISTINCT c.ip) FILTER(WHERE a.ip IS NULL),count(*) FILTER(WHERE c.source='machine' AND a.ip IS NOT NULL),count(*) FILTER(WHERE c.source='machine' AND a.ip IS NULL),count(*) FILTER(WHERE c.source='trusted' AND a.ip IS NOT NULL),count(*) FILTER(WHERE c.source='trusted' AND a.ip IS NULL) FROM current_ips c LEFT JOIN active a USING(ip,source)`, start).Row().Scan(&d.Activity.Total, &d.Activity.Active, &d.Activity.Passive, &d.Activity.MachineActive, &d.Activity.MachinePassive, &d.Activity.TrustedActive, &d.Activity.TrustedPassive)
	if err != nil {
		return d, err
	}
	if d.Activity.Total > 0 {
		d.Activity.ActiveRate = float64(d.Activity.Active) * 100 / float64(d.Activity.Total)
	}
	seriesQuery := fmt.Sprintf(`WITH buckets AS (SELECT generate_series(date_trunc('%s',$1::timestamptz)-($2-1)*interval '1 %s',date_trunc('%s',$1::timestamptz),interval '1 %s') at), counts AS (SELECT date_trunc('%s',occurred_at) at,count(*) requests,count(*) FILTER(WHERE status_code>=400) errors FROM firewall_requests WHERE occurred_at>=date_trunc('%s',$1::timestamptz)-($2-1)*interval '1 %s' AND occurred_at<=$1 GROUP BY 1) SELECT b.at,coalesce(c.requests,0),coalesce(c.errors,0) FROM buckets b LEFT JOIN counts c USING(at) ORDER BY b.at`, bucket, bucket, bucket, bucket, bucket, bucket, bucket)
	rows, err := s.DB.WithContext(ctx).Raw(seriesQuery, now, bucketCount).Rows()
	if err != nil {
		return d, err
	}
	for rows.Next() {
		var p TimePoint
		if err := rows.Scan(&p.At, &p.Requests, &p.Errors); err != nil {
			rows.Close()
			return d, err
		}
		d.Series = append(d.Series, p)
	}
	rows.Close()
	rows, err = s.DB.WithContext(ctx).Raw(`SELECT coalesce(ecosystem_name,'Individual IP'),count(DISTINCT r.id),count(DISTINCT host(i.ipv4)) FROM firewall_requests r LEFT JOIN firewall_request_ips i ON i.request_id=r.id AND i.request_occurred_at=r.occurred_at WHERE r.occurred_at >= ? GROUP BY 1 ORDER BY 2 DESC`, start).Rows()
	if err != nil {
		return d, err
	}
	for rows.Next() {
		var v EcosystemStats
		if err := rows.Scan(&v.Name, &v.Requests, &v.UniqueIPs); err != nil {
			rows.Close()
			return d, err
		}
		d.Ecosystems = append(d.Ecosystems, v)
	}
	rows.Close()
	d.Formats, err = s.namedCounts(ctx, `SELECT format,count(*) FROM firewall_requests WHERE occurred_at >= $1 GROUP BY format ORDER BY format`, start)
	if err != nil {
		return d, err
	}
	d.Sources, err = s.namedCounts(ctx, `SELECT source_type,count(*) FROM firewall_request_ips WHERE request_occurred_at >= $1 GROUP BY source_type ORDER BY source_type`, start)
	return d, err
}

// namedCounts executes a two-column aggregate query.
func (s *Store) namedCounts(ctx context.Context, q string, start time.Time) ([]NamedCount, error) {
	rows, err := s.DB.WithContext(ctx).Raw(q, start).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []NamedCount
	for rows.Next() {
		var v NamedCount
		if err := rows.Scan(&v.Name, &v.Count); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
