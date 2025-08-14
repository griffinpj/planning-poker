package database

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type DB struct {
	*sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	// Use SQLite connection string with performance optimizations
	connectionString := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=on", dbPath)
	
	sqlDB, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings for better performance
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{DB: sqlDB}

	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func (db *DB) runMigrations() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		return err
	}

	return nil
}