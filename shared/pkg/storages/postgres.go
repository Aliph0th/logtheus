package storages

import (
	"fmt"
	"log/slog"
	sl "logtheus/shared/pkg/utils/logger"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func NewPostgres(host string, port int, user string, password string, name string) (*Database, error) {
	dataSourceName := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, user, password, name, port)
	db, err := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to Postgres", sl.Error(err))
		os.Exit(1)
	}
	slog.Info("Connected to Postgres database", "host", host, "port", port, "dbname", name)
	return &Database{DB: db}, nil
}

func (d *Database) Migrate(models ...any) error {
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
