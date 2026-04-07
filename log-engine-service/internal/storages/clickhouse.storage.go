package storages

import (
	"fmt"
	"log/slog"
	"logtheus/logengine/internal/config"
	sl "logtheus/shared/pkg/utils/logger"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	gormClickhouse "gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type ClickHouse struct {
	DB *gorm.DB
}

func NewClickHouseStorage(cfg *config.AppConfig) (*ClickHouse, error) {
	db, err := gorm.Open(gormClickhouse.New(gormClickhouse.Config{
		Conn: clickhouse.OpenDB(&clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouse.Host, cfg.ClickHouse.Port)},
			Auth: clickhouse.Auth{
				Database: cfg.ClickHouse.Name,
				Username: cfg.ClickHouse.User,
				Password: cfg.ClickHouse.Password,
			},
			Settings: clickhouse.Settings{
				"max_execution_time": 60,
			},
			Compression: &clickhouse.Compression{Method: clickhouse.CompressionLZ4},
		}),
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to connect to ClickHouse", slog.String("host", cfg.ClickHouse.Host), slog.Int("port", cfg.ClickHouse.Port), slog.String("user", cfg.ClickHouse.User), sl.Error(err))
		os.Exit(1)
	}
	if err := sqlDB.Ping(); err != nil {
		slog.Error("Failed to ping ClickHouse", slog.String("host", cfg.ClickHouse.Host), slog.Int("port", cfg.ClickHouse.Port), slog.String("user", cfg.ClickHouse.User), sl.Error(err))
		os.Exit(1)
	}
	slog.Info("Successfully connected to ClickHouse", slog.String("host", cfg.ClickHouse.Host), slog.Int("port", cfg.ClickHouse.Port), slog.String("user", cfg.ClickHouse.User))
	return &ClickHouse{DB: db}, nil
}

func (c *ClickHouse) Migrate(model any, options string) error {
	if err := c.DB.Set("gorm:table_options", options).AutoMigrate(model); err != nil {
		return fmt.Errorf("Failed to migrate database: %w", err)
	}
	slog.Info("Database migration completed")
	return nil
}

func (c *ClickHouse) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	slog.Info("Closing ClickHouse connection")
	return sqlDB.Close()
}
