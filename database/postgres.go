package database

import (
	"fmt"

	"github.com/therealak12/api-health-check/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresInstance(cfg config.Database) (*gorm.DB, error) {
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Name)
	db, err := gorm.Open(postgres.Open(dataSourceName))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
