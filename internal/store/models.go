package store

import "time"

// Ecosystem describes a product ecosystem and its usage summary.
type Ecosystem struct {
	ID                   string `gorm:"type:uuid;default:gen_random_uuid()"`
	Name                 string
	DisplayName          string
	IsActive             bool
	MachineCount         int        `gorm:"-"`
	LastUsedAt           *time.Time `gorm:"-"`
	CreatedAt, UpdatedAt time.Time
}

// TableName returns the persisted ecosystem table.
func (Ecosystem) TableName() string { return "ecosystems" }

// Machine describes one IPv4 source associated with an ecosystem.
type Machine struct {
	ID, Hostname, EcosystemID, EcosystemName, IPv4 string
	LastUsedAt                                     *time.Time
	LastSyncAt                                     time.Time
	IsManaged                                      bool
	CreatedAt, UpdatedAt                           time.Time
}

// TableName returns the persisted machine table.
func (Machine) TableName() string { return "machines" }

// TrustedService groups hostnames whose resolved addresses are globally trusted.
type TrustedService struct {
	ID, Name, Description string
	IsActive              bool
	Hostnames             []TrustedHostname `gorm:"-"`
	CreatedAt, UpdatedAt  time.Time
}

// TrustedHostname describes one DNS name and its resolution state.
type TrustedHostname struct {
	ID, Hostname, Status         string
	LastAttemptAt, LastSuccessAt *time.Time
	LastError                    *string
	IPs                          []ResolvedIP `gorm:"-"`
}

// ResolvedIP describes an address observed for a trusted hostname.
type ResolvedIP struct {
	ID, IPv4                string
	FirstSeenAt, LastSeenAt time.Time
	LastUsedAt              *time.Time
	IsCurrent               bool
}

// EffectiveIP describes a published address and its provenance.
type EffectiveIP struct {
	IPv4, SourceType, SourceID, Hostname, Service string
	LastUsedAt, LastSyncAt                        *time.Time
	Stale                                         bool
}

// HistoryInput contains the data recorded for a firewall response.
type HistoryInput struct {
	RouteKind, EcosystemID, EcosystemName, Format string
	StatusCode, DurationMS                        int
	ETag                                          string
	NotModified                                   bool
	Entries                                       []EffectiveIP
}

// IngestionKey exposes non-secret API-key metadata.
type IngestionKey struct {
	ID, Name, Prefix                 string
	ExpiresAt, RevokedAt, LastUsedAt *time.Time
	CreatedAt                        time.Time
}

// TableName returns the persisted ingestion-key table.
func (IngestionKey) TableName() string { return "ingestion_keys" }

// Dashboard contains aggregated administration metrics.
type Dashboard struct {
	Window     string           `json:"window"`
	Requests   RequestStats     `json:"requests"`
	Activity   ActivityStats    `json:"activity"`
	Series     []TimePoint      `json:"series"`
	Ecosystems []EcosystemStats `json:"ecosystems"`
	Formats    []NamedCount     `json:"formats"`
	Sources    []NamedCount     `json:"sources"`
}

// RequestStats contains request volume and latency indicators.
type RequestStats struct {
	Total            int64   `json:"total"`
	PreviousTotal    int64   `json:"previousTotal"`
	ChangePercent    float64 `json:"changePercent"`
	SuccessRate      float64 `json:"successRate"`
	ErrorRate        float64 `json:"errorRate"`
	NotModified      int64   `json:"notModified"`
	AverageLatencyMS float64 `json:"averageLatencyMs"`
	P95LatencyMS     float64 `json:"p95LatencyMs"`
}

// ActivityStats contains active and passive IP counts.
type ActivityStats struct {
	Active         int64   `json:"active"`
	Passive        int64   `json:"passive"`
	Total          int64   `json:"total"`
	ActiveRate     float64 `json:"activeRate"`
	MachineActive  int64   `json:"machineActive"`
	MachinePassive int64   `json:"machinePassive"`
	TrustedActive  int64   `json:"trustedActive"`
	TrustedPassive int64   `json:"trustedPassive"`
}

// TimePoint is one dashboard time-series bucket.
type TimePoint struct {
	At       time.Time `json:"at"`
	Requests int64     `json:"requests"`
	Errors   int64     `json:"errors"`
}

// EcosystemStats contains activity grouped by ecosystem.
type EcosystemStats struct {
	Name      string `json:"name"`
	Requests  int64  `json:"requests"`
	UniqueIPs int64  `json:"uniqueIps"`
}

// NamedCount is a labelled aggregate count.
type NamedCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
