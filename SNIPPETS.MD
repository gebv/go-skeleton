# Snippets

## Setup postgres

```go
// Postgres
sqlDB := configure.SetupPostgres(settings)
reformDB := reform.NewDB(sqlDB, postgresql.Dialect, reform.NewPrintfLogger(zap.L().Sugar().Debugf))
```
