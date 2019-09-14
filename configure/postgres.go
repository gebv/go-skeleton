package configure

import (
	"database/sql"

	"github.com/gebv/go-skeleton/dbstat"
	"github.com/gebv/go-skeleton/settings"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// SetupPostgres setup connect to postgres.
func SetupPostgres(s *settings.Settings) *sql.DB {
	// Postgres init
	sqlDB, err := sql.Open("postgres", s.Postgres.URL)
	if err != nil {
		zap.L().Panic("Failed to connect to PostgreSQL.", zap.Error(err))
	}
	sqlDB.SetConnMaxLifetime(s.Postgres.ConnMaxLifetime)
	sqlDB.SetMaxOpenConns(s.Postgres.MaxOpenConns)
	sqlDB.SetMaxIdleConns(s.Postgres.MaxIdleConns)
	if err = sqlDB.Ping(); err != nil {
		zap.L().Panic("Failed to connect ping PostgreSQL.", zap.Error(err))
	}
	zap.L().Info("Postgres - Connected!")

	prometheus.MustRegister(dbstat.New(sqlDB, "postgres"))

	return sqlDB
}
