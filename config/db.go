package config

import (
	"fmt"
	"jwt_go/graph/model"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDb() (*gorm.DB, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Get database credentials from environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "unictive"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "brandingku_graphql"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", host, user, password, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	// Auto-migrate only the existing models in this project
	err = db.AutoMigrate(
		&model.UserDB{},
		&model.RoleDB{},
		&model.PermissionDB{},
		&model.RolePermissionDB{},
	)

	if err != nil {
		return nil, err
	}

	// Run seeders (idempotent)
	if err := Seed(db); err != nil {
		return nil, err
	}

	return db, nil
}
