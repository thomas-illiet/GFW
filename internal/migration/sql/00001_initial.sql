-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE ecosystems (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE CHECK (name ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    display_name text NOT NULL CHECK (length(trim(display_name)) > 0),
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE machines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(), hostname text NOT NULL CHECK (length(trim(hostname)) > 0),
    ecosystem_id uuid NOT NULL REFERENCES ecosystems(id), ipv4 inet NOT NULL CHECK (family(ipv4) = 4 AND masklen(ipv4) = 32),
    last_used_at timestamptz, last_sync_at timestamptz NOT NULL DEFAULT now(), is_managed boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), UNIQUE (ecosystem_id, ipv4)
);

CREATE TABLE trusted_services (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text NOT NULL UNIQUE, description text NOT NULL DEFAULT '',
    is_active boolean NOT NULL DEFAULT true, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE trusted_hostnames (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(), service_id uuid NOT NULL REFERENCES trusted_services(id) ON DELETE CASCADE,
    hostname text NOT NULL, last_attempt_at timestamptz, last_success_at timestamptz, last_error text, UNIQUE(service_id, hostname)
);

CREATE TABLE trusted_resolved_ips (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(), hostname_id uuid NOT NULL REFERENCES trusted_hostnames(id) ON DELETE CASCADE,
    ipv4 inet NOT NULL CHECK (family(ipv4) = 4 AND masklen(ipv4) = 32), first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(), last_used_at timestamptz, is_current boolean NOT NULL DEFAULT true,
    UNIQUE(hostname_id, ipv4)
);

CREATE TABLE ingestion_keys (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text NOT NULL UNIQUE, prefix text NOT NULL UNIQUE, secret_hash text NOT NULL,
    expires_at timestamptz, revoked_at timestamptz, last_used_at timestamptz, created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE firewall_requests (
    id uuid NOT NULL DEFAULT gen_random_uuid(), occurred_at timestamptz NOT NULL DEFAULT now(),
    route_kind text NOT NULL CHECK (route_kind IN ('ecosystem', 'ip')), ecosystem_id uuid, ecosystem_name text,
    format text NOT NULL CHECK (format IN ('json', 'txt')), status_code integer NOT NULL, duration_ms integer NOT NULL,
    etag text, not_modified boolean NOT NULL DEFAULT false, machine_count integer NOT NULL DEFAULT 0,
    trusted_count integer NOT NULL DEFAULT 0, total_count integer NOT NULL DEFAULT 0, PRIMARY KEY (id, occurred_at)
) PARTITION BY RANGE (occurred_at);

CREATE TABLE firewall_request_ips (
    request_id uuid NOT NULL, request_occurred_at timestamptz NOT NULL, ipv4 inet NOT NULL CHECK (family(ipv4) = 4),
    source_type text NOT NULL CHECK (source_type IN ('machine', 'trusted')), source_id uuid NOT NULL,
    PRIMARY KEY (request_id, request_occurred_at, ipv4, source_type, source_id),
    FOREIGN KEY (request_id, request_occurred_at) REFERENCES firewall_requests(id, occurred_at) ON DELETE CASCADE
) PARTITION BY RANGE (request_occurred_at);

CREATE TABLE audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(), occurred_at timestamptz NOT NULL DEFAULT now(), actor text NOT NULL,
    action text NOT NULL, resource_type text NOT NULL, resource_id text, details jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX machines_ecosystem_idx ON machines(ecosystem_id);
CREATE INDEX machines_last_used_idx ON machines(last_used_at);
CREATE INDEX trusted_ips_current_idx ON trusted_resolved_ips(is_current, ipv4);
CREATE INDEX audit_events_occurred_idx ON audit_events(occurred_at DESC);
