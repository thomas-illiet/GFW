# Implementation guide

## Cobra

The root command defaults to `serve`. Subcommands share persistent configuration flags and use injected dependencies in tests. Release builds inject version, commit, and date through linker flags.

## Viper

Every command creates an isolated loader. Precedence is flags, `GFW_*` environment variables, YAML, then defaults. Validation happens after merging so every source follows the same safety rules.

## GORM

GORM owns connection pooling, CRUD, filters, and transactions. SQL is deliberately retained for partition DDL, PostgreSQL `inet` projections, unions, percentile aggregates, and bulk history operations where an ORM expression would obscure behavior. `AutoMigrate` is never used.

Repositories translate missing rows to `store.ErrNotFound`. Handlers map domain outcomes to the existing HTTP status and error envelopes.

## Goose

`internal/migration` owns the only production migration directory. Its Provider
uses embedded, forward-only SQL and a PostgreSQL session lock. Test fixtures use
in-memory expectations rather than additional migration directories. The serve
preflight reads Goose history directly so startup remains schema-read-only.

## Source organization

Code is grouped by domain rather than technical catch-all packages. Production files are kept below the automated size ceiling; new features should normally use smaller files targeting roughly 250 lines. Every named production function and method requires an adjacent explanatory comment.
