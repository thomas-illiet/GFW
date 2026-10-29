# Architecture decisions

## GORM with explicit PostgreSQL SQL

GORM is the default persistence API because it centralizes connection management, transactions, CRUD, and test seams. PostgreSQL-specific SQL remains explicit when translating it into ORM calls would reduce clarity or correctness, especially partitions, `inet`, unions, and percentile aggregates.

## SQL migrations remain authoritative

Production schema changes must be reviewable, repeatable, forward-only and
independent from application startup. Embedded Goose SQL provides transactions,
numeric history and PostgreSQL session locking. `AutoMigrate` is disabled and
`serve` is read-only with respect to schema. Operational corrections use a new
up migration rather than a destructive rollback.

## Database-free unit suite

The default suite favors fast deterministic mocks. SQL expectations cover query
shape and failure behavior, while the manual release check adds a real Compose
PostgreSQL acceptance run. The database-backed test complements rather than
replaces the unit coverage gate.
