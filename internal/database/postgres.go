package database

import (
	"car_catalog/internal/config"
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func MigrateDatabase(cfg *config.Config) error {
	log.Println("[INFO] MigrateDatabase - Starting database migration...")

	databaseUrl := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		log.Fatalf("error open connection to apply migration: %s", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("could not init driver: %s", err)
	}

	migrate, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx", driver)
	if err != nil {
		log.Fatalf("could not apply the migration: %s", err)
	}

	migrate.Up()

	log.Println("[INFO] MigrateDatabase - Database migration completed successfully")
	return nil
}

func DatabaseConnection(cfg *config.Config) *pgxpool.Pool {
	log.Println("[INFO] DatabaseConnection - Connecting to database...")

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	log.Println("[INFO] DatabaseConnection - Successfully connected to database")

	return conn
}
