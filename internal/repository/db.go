package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"maid-recruitment-tracking/internal/config"
)

var (
	sharedDBMu sync.Mutex
	sharedDBs  = map[string]*gorm.DB{}
)

func openDatabase(cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	dsn := strings.TrimSpace(cfg.DatabaseURL)
	if dsn == "" {
		return nil, fmt.Errorf("database url is empty")
	}

	sharedDBMu.Lock()
	existing := sharedDBs[dsn]
	sharedDBMu.Unlock()
	if existing != nil {
		return existing, nil
	}

	driverConfig := postgres.Config{DSN: dsn}
	if shouldPreferSimpleProtocol(dsn) {
		driverConfig.PreferSimpleProtocol = true
	}

	db, err := gorm.Open(postgres.New(driverConfig), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("connect postgres: get sql db: %w", err)
	}
	configureConnectionPool(sqlDB, dsn)

	sharedDBMu.Lock()
	defer sharedDBMu.Unlock()
	if existing = sharedDBs[dsn]; existing != nil {
		_ = sqlDB.Close()
		return existing, nil
	}

	sharedDBs[dsn] = db
	return db, nil
}

func shouldPreferSimpleProtocol(dsn string) bool {
	normalized := strings.ToLower(strings.TrimSpace(dsn))
	return strings.Contains(normalized, ".pooler.supabase.com:6543")
}

func configureConnectionPool(sqlDB *sql.DB, dsn string) {
	maxOpen := 10
	maxIdle := 5

	switch {
	case strings.Contains(strings.ToLower(dsn), ".pooler.supabase.com:5432"):
		maxOpen = 4
		maxIdle = 2
	case strings.Contains(strings.ToLower(dsn), ".pooler.supabase.com:6543"):
		maxOpen = 8
		maxIdle = 4
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
}
