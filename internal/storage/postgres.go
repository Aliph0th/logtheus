package storage

import (
	"fmt"
	"log/slog"
	"logtheus/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func NewPostgres(cfg *config.AppConfig) (*Database, error) {
	dataSourceName := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", cfg.DB.Host, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.Port)
	db, err := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Postgres: %w", err)
	}
	slog.Info("Connected to Postgres database", "host", cfg.DB.Host, "port", cfg.DB.Port, "dbname", cfg.DB.Name)
	return &Database{DB: db}, nil
}

func (d *Database) Migrate(models ...interface{}) error {
	if err := d.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("Failed to migrate database: %w", err)
	}
	slog.Info("Database migration completed")
	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	slog.Info("Closing database connection")
	return sqlDB.Close()
}
