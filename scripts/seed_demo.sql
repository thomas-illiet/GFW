\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE demo_ecosystems (
    name text PRIMARY KEY,
    display_name text NOT NULL,
    position integer NOT NULL,
    is_active boolean NOT NULL
) ON COMMIT DROP;

INSERT INTO demo_ecosystems(name, display_name, position, is_active) VALUES
    ('customer-identity', 'Customer Identity', 1, true),
    ('checkout', 'Checkout & Orders', 2, true),
    ('merchant-platform', 'Merchant Platform', 3, true),
    ('data-platform', 'Data Platform', 4, true),
    ('observability', 'Observability', 5, true),
    ('developer-tools', 'Developer Tools', 6, true),
    ('corporate-services', 'Corporate Services', 7, true),
    ('legacy-core', 'Legacy Core', 8, false);

INSERT INTO ecosystems(name, display_name, is_active, created_at, updated_at)
SELECT
    name,
    display_name,
    is_active,
    now() - make_interval(days => 210 - position * 11),
    now() - make_interval(days => position % 5)
FROM demo_ecosystems
ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    is_active = EXCLUDED.is_active,
    updated_at = EXCLUDED.updated_at;

WITH roles(role, role_position) AS (
    VALUES
        ('api', 1), ('worker', 2), ('web', 3), ('scheduler', 4),
        ('events', 5), ('batch', 6), ('admin', 7), ('gateway', 8),
        ('cache', 9), ('database', 10), ('queue', 11), ('reporting', 12),
        ('sync', 13), ('search', 14), ('metrics', 15), ('backup', 16)
), generated AS (
    SELECT
        e.id AS ecosystem_id,
        d.name AS ecosystem_name,
        d.position AS ecosystem_position,
        r.role,
        r.role_position
    FROM demo_ecosystems d
    JOIN ecosystems e USING (name)
    CROSS JOIN roles r
)
INSERT INTO machines(
    hostname,
    ecosystem_id,
    ipv4,
    last_used_at,
    last_sync_at,
    is_managed,
    created_at,
    updated_at
)
SELECT
    format('%s-%s-%s', ecosystem_name, role, to_char(role_position, 'FM00')),
    ecosystem_id,
    format('10.%s.%s.%s', 40 + ecosystem_position, ecosystem_position, 20 + role_position)::inet,
    NULL,
    now() - make_interval(mins => ecosystem_position * 7 + role_position),
    role_position % 3 <> 0,
    now() - make_interval(days => 75 - ecosystem_position, hours => role_position),
    now() - make_interval(mins => ecosystem_position * 7 + role_position)
FROM generated
ON CONFLICT (ecosystem_id, ipv4) DO UPDATE SET
    hostname = EXCLUDED.hostname,
    last_sync_at = EXCLUDED.last_sync_at,
    is_managed = EXCLUDED.is_managed,
    updated_at = EXCLUDED.updated_at;

CREATE TEMP TABLE demo_services (
    name text PRIMARY KEY,
    description text NOT NULL,
    hostname text NOT NULL,
    position integer NOT NULL
) ON COMMIT DROP;

INSERT INTO demo_services(name, description, hostname, position) VALUES
    ('Demo Identity Provider', 'OIDC and SSO endpoints shared by all products.', 'login.demo.example.net', 1),
    ('Demo Artifact Registry', 'Container images and build artifacts.', 'registry.demo.example.net', 2),
    ('Demo Observability SaaS', 'Metrics, traces and alert delivery.', 'telemetry.demo.example.net', 3),
    ('Demo Security Updates', 'Operating system and security update mirrors.', 'updates.demo.example.net', 4),
    ('Demo Payment Gateway', 'External payment processor API.', 'payments.demo.example.net', 5),
    ('Demo Source Control', 'Git and continuous integration endpoints.', 'git.demo.example.net', 6);

INSERT INTO trusted_services(name, description, is_active, created_at, updated_at)
SELECT
    name,
    description,
    true,
    now() - make_interval(days => 160 - position * 9),
    now() - make_interval(hours => position)
FROM demo_services
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    is_active = true,
    updated_at = EXCLUDED.updated_at;

INSERT INTO trusted_hostnames(
    service_id,
    hostname,
    last_attempt_at,
    last_success_at,
    last_error
)
SELECT
    s.id,
    d.hostname,
    now() - make_interval(mins => d.position * 3),
    now() - make_interval(mins => d.position * 3),
    NULL
FROM demo_services d
JOIN trusted_services s USING (name)
ON CONFLICT (service_id, hostname) DO UPDATE SET
    last_attempt_at = EXCLUDED.last_attempt_at,
    last_success_at = EXCLUDED.last_success_at,
    last_error = NULL;

WITH generated AS (
    SELECT
        h.id AS hostname_id,
        d.position,
        ip_position
    FROM demo_services d
    JOIN trusted_services s USING (name)
    JOIN trusted_hostnames h ON h.service_id = s.id AND h.hostname = d.hostname
    CROSS JOIN generate_series(1, 2) AS ip_position
)
INSERT INTO trusted_resolved_ips(
    hostname_id,
    ipv4,
    first_seen_at,
    last_seen_at,
    last_used_at,
    is_current
)
SELECT
    hostname_id,
    CASE
        WHEN ip_position = 1 THEN format('192.0.2.%s', position * 10 + ip_position)::inet
        ELSE format('198.51.100.%s', position * 10 + ip_position)::inet
    END,
    now() - make_interval(days => 120 - position),
    now() - make_interval(mins => position * 3),
    NULL,
    true
FROM generated
ON CONFLICT (hostname_id, ipv4) DO UPDATE SET
    last_seen_at = EXCLUDED.last_seen_at,
    is_current = true;

INSERT INTO ingestion_keys(
    name,
    prefix,
    secret_hash,
    expires_at,
    revoked_at,
    last_used_at,
    created_at
) VALUES
    ('Demo CMDB sync', 'demo_cmdb', crypt('demo-secret-not-for-production', gen_salt('bf')), now() + interval '120 days', NULL, now() - interval '18 minutes', now() - interval '90 days'),
    ('Demo discovery agent', 'demo_discovery', crypt('demo-secret-not-for-production', gen_salt('bf')), now() + interval '30 days', NULL, now() - interval '2 hours', now() - interval '45 days'),
    ('Demo retired importer', 'demo_retired', crypt('demo-secret-not-for-production', gen_salt('bf')), now() - interval '10 days', now() - interval '12 days', now() - interval '40 days', now() - interval '180 days')
ON CONFLICT (name) DO UPDATE SET
    expires_at = EXCLUDED.expires_at,
    revoked_at = EXCLUDED.revoked_at,
    last_used_at = EXCLUDED.last_used_at;

-- Only replace history generated by this demo script.
DELETE FROM firewall_requests WHERE etag LIKE '"demo-%';

CREATE TEMP TABLE demo_requests (
    id uuid NOT NULL,
    occurred_at timestamptz NOT NULL,
    ecosystem_id uuid NOT NULL,
    PRIMARY KEY (id, occurred_at)
) ON COMMIT DROP;

WITH request_slots AS (
    SELECT
        e.id AS ecosystem_id,
        e.name AS ecosystem_name,
        d.position,
        slot,
        row_number() OVER (ORDER BY slot, d.position) AS sequence
    FROM demo_ecosystems d
    JOIN ecosystems e USING (name)
    CROSS JOIN generate_series(
        date_trunc('hour', now() - interval '30 days'),
        date_trunc('hour', now()) - interval '2 hours',
        interval '2 hours'
    ) AS slot
    WHERE d.is_active OR extract(day FROM slot)::integer % 4 = 0
), inserted AS (
    INSERT INTO firewall_requests(
        occurred_at,
        route_kind,
        ecosystem_id,
        ecosystem_name,
        format,
        status_code,
        duration_ms,
        etag,
        not_modified,
        machine_count,
        trusted_count,
        total_count
    )
    SELECT
        slot + make_interval(mins => (position * 7 + sequence % 41)::integer),
        'ecosystem',
        ecosystem_id,
        ecosystem_name,
        CASE WHEN sequence % 4 = 0 THEN 'txt' ELSE 'json' END,
        CASE
            WHEN sequence % 53 = 0 THEN 503
            WHEN sequence % 37 = 0 THEN 500
            WHEN sequence % 5 = 0 THEN 304
            ELSE 200
        END,
        8 + (sequence * 17 % 185)::integer,
        format('"demo-%s"', md5(ecosystem_name || slot::text || sequence::text)),
        sequence % 5 = 0 AND sequence % 37 <> 0 AND sequence % 53 <> 0,
        10,
        4,
        14
    FROM request_slots
    RETURNING id, occurred_at, ecosystem_id
)
INSERT INTO demo_requests(id, occurred_at, ecosystem_id)
SELECT id, occurred_at, ecosystem_id FROM inserted;

INSERT INTO firewall_request_ips(
    request_id,
    request_occurred_at,
    ipv4,
    source_type,
    source_id
)
SELECT
    r.id,
    r.occurred_at,
    selected.ipv4,
    selected.source_type,
    selected.source_id
FROM demo_requests r
JOIN LATERAL (
    SELECT m.ipv4, 'machine'::text AS source_type, m.id AS source_id
    FROM machines m
    WHERE m.ecosystem_id = r.ecosystem_id
    ORDER BY md5(m.id::text || r.id::text)
    LIMIT 10
) selected ON true;

INSERT INTO firewall_request_ips(
    request_id,
    request_occurred_at,
    ipv4,
    source_type,
    source_id
)
SELECT
    r.id,
    r.occurred_at,
    selected.ipv4,
    selected.source_type,
    selected.source_id
FROM demo_requests r
JOIN LATERAL (
    SELECT ip.ipv4, 'trusted'::text AS source_type, ip.id AS source_id
    FROM trusted_resolved_ips ip
    JOIN trusted_hostnames h ON h.id = ip.hostname_id
    JOIN trusted_services s ON s.id = h.service_id
    WHERE ip.is_current AND s.is_active
    ORDER BY md5(ip.id::text || r.id::text)
    LIMIT 4
) selected ON true;

UPDATE machines m
SET last_used_at = usage.last_used_at
FROM (
    SELECT source_id, max(request_occurred_at) AS last_used_at
    FROM firewall_request_ips
    WHERE source_type = 'machine'
    GROUP BY source_id
) usage
WHERE usage.source_id = m.id;

UPDATE trusted_resolved_ips ip
SET last_used_at = usage.last_used_at
FROM (
    SELECT source_id, max(request_occurred_at) AS last_used_at
    FROM firewall_request_ips
    WHERE source_type = 'trusted'
    GROUP BY source_id
) usage
WHERE usage.source_id = ip.id;

DELETE FROM audit_events WHERE actor = 'demo-seeder';

INSERT INTO audit_events(occurred_at, actor, action, resource_type, resource_id, details)
SELECT
    now() - make_interval(days => day_offset, hours => day_offset % 8),
    'demo-seeder',
    CASE day_offset % 3
        WHEN 0 THEN 'machine.updated'
        WHEN 1 THEN 'trusted_service.resolved'
        ELSE 'ecosystem.updated'
    END,
    CASE day_offset % 3
        WHEN 0 THEN 'machine'
        WHEN 1 THEN 'trusted_service'
        ELSE 'ecosystem'
    END,
    'demo-' || day_offset,
    jsonb_build_object('demo', true, 'result', 'success')
FROM generate_series(1, 30) AS day_offset;

COMMIT;

SELECT
    (SELECT count(*) FROM ecosystems) AS ecosystems,
    (SELECT count(*) FROM machines) AS machines,
    (SELECT count(*) FROM trusted_services) AS trusted_services,
    (SELECT count(*) FROM firewall_requests WHERE occurred_at >= now() - interval '30 days') AS requests_30d,
    (SELECT count(*) FROM firewall_request_ips WHERE request_occurred_at >= now() - interval '30 days') AS served_ips_30d;
