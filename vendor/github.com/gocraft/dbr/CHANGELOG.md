# Change log

Not all minor changes may be noted here, but all large and/or breaking changes
should be.

## v2.0 - 2015-10-09

### Added
- PostgreSQL support!
- `Open(driver, dsn string, log EventReceiver)` creates an underlying connection for you based on a supplied driver and dsn string
- All builders are now available without a `Session` facilitating much more complex queries
- More common SQL support: Subqueries, Unions, Joins, Aliases
- More complex condition building support: And/Or/Eq/Neq/Gt/Gte/Lt/Lte

### Deprecated
- `NewConnection` is deprecated. It assumes MySQL driver. Please use `Open` instead

### Changed
- `NullTime` no longer relies on the mysql package. E.g. instead of `NullTime{mysql.NullTime{Time: t, Valid: true}}` it's now simply `NullTime{Time: t, Valid: true}`
- All `*Builder` structs now embed a corresponding `*Stmt` struct (E.g. `SelectBuilder` embeds `SelectStmt`). All non-`Session` specific properies have been moved the `*Stmt` structs
