package database

import (
	"fmt"
	"log"
	"time"

	"github.com/terminaloj/backend/internal/config"
	"github.com/terminaloj/backend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// Init opens the MySQL connection, configures the pool and migrates schema.
func Init(cfg config.MySQLConfig) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn),
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	conn, err := gorm.Open(mysql.Open(cfg.DSN), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if cfg.AutoMigrate {
		if err := conn.AutoMigrate(
			&models.User{},
			&models.Problem{},
			&models.Submission{},
			&models.SubmissionCaseResult{},
			&models.Announcement{},
			&models.Conversation{},
			&models.Message{},
		); err != nil {
			return nil, fmt.Errorf("auto migrate: %w", err)
		}
		log.Println("[db] auto migration complete")
	}

	db = conn

	if cfg.Seed {
		if err := Seed(conn); err != nil {
			log.Printf("[db] seed warn: %v", err)
		}
	}
	return conn, nil
}

// DB returns the shared *gorm.DB instance.
func DB() *gorm.DB { return db }
