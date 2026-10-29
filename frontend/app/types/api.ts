export type Role = 'viewer' | 'editor' | 'admin'
export interface User {
  subject: string
  name: string
  email: string
  role: Role
  expiresAt: string
}
export interface Ecosystem {
  ID: string
  Name: string
  DisplayName: string
  IsActive: boolean
  MachineCount: number
  LastUsedAt?: string
  CreatedAt: string
  UpdatedAt: string
}
export interface Machine {
  ID: string
  Hostname: string
  EcosystemID: string
  EcosystemName: string
  IPv4: string
  LastUsedAt?: string
  LastSyncAt: string
  IsManaged: boolean
  CreatedAt: string
  UpdatedAt: string
}
export type MachineSource = '' | 'managed' | 'manual'
export interface ResolvedIP {
  ID: string
  IPv4: string
  FirstSeenAt: string
  LastSeenAt: string
  LastUsedAt?: string
  IsCurrent: boolean
}
export interface TrustedHostname {
  ID: string
  Hostname: string
  Status: 'fresh' | 'stale' | 'pending'
  LastAttemptAt?: string
  LastSuccessAt?: string
  LastError?: string
  IPs: ResolvedIP[]
}
export interface TrustedService {
  ID: string
  Name: string
  Description: string
  IsActive: boolean
  Hostnames: TrustedHostname[]
  CreatedAt: string
  UpdatedAt: string
}
export interface IngestionKey {
  ID: string
  Name: string
  Prefix: string
  ExpiresAt?: string
  RevokedAt?: string
  LastUsedAt?: string
  CreatedAt: string
}
export interface EffectiveIP {
  IPv4: string
  SourceType: 'machine' | 'trusted'
  SourceID: string
  Hostname: string
  Service: string
  LastUsedAt?: string
  LastSyncAt?: string
  Stale: boolean
}
export interface Dashboard {
  window: '24h' | '7d' | '30d'
  requests: {
    total: number
    previousTotal: number
    changePercent: number
    successRate: number
    errorRate: number
    notModified: number
    averageLatencyMs: number
    p95LatencyMs: number
  }
  activity: {
    active: number
    passive: number
    total: number
    activeRate: number
    machineActive: number
    machinePassive: number
    trustedActive: number
    trustedPassive: number
  }
  series: Array<{ at: string; requests: number; errors: number }>
  ecosystems: Array<{ name: string; requests: number; uniqueIps: number }>
  formats: Array<{ name: string; count: number }>
  sources: Array<{ name: string; count: number }>
}
export interface ListResponse<T> {
  items: T[]
  total: number
}
