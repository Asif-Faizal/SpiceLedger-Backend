package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/saravanan/spice_backend/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Create DB if not exists
	// Connect to default 'postgres' database
	defaultDSN := strings.Replace(cfg.DBDSN, "dbname=spice_db", "dbname=postgres", 1)
	if !strings.Contains(defaultDSN, "dbname=postgres") {
		// Fallback if Replace didn't work (e.g. if string is different), just append or simpler:
		// assuming standard format, but let's try to be smart.
		// If DSN doesn't have dbname, we add it.
		// Simpler: Just try connecting to postgres DB explicitly construction DSN if needed.
		// For now, let's assume the string replace works as user provided DSN in .env matches.
	}

	dbAdmin, err := sql.Open("pgx", defaultDSN)
	if err != nil {
		log.Fatalf("Failed to open admin DB connection: %v", err)
	}
	defer dbAdmin.Close()

	// Check if DB exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'spice_db')"
	err = dbAdmin.QueryRowContext(context.Background(), query).Scan(&exists)
	if err != nil {
		// If we can't check, maybe we just try to create or fail.
		// If connecting to 'postgres' failed (e.g. auth), we might crash here.
		log.Printf("Warning: Failed to check if DB exists: %v", err)
	}

	if !exists {
		log.Println("Database spice_db does not exist. Creating...")
		_, err = dbAdmin.ExecContext(context.Background(), "CREATE DATABASE spice_db")
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		log.Println("Database spice_db created.")
	} else {
		log.Println("Database spice_db already exists.")
	}
	dbAdmin.Close()

	// 2. Connect to the target App DB
	db, err := sql.Open("pgx", cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	log.Println("Connected to spice_db.")

	// Read migration files
	files, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, filename := range sqlFiles {
		log.Printf("Applying migration: %s", filename)
		content, err := os.ReadFile(filepath.Join("migrations", filename))
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", filename, err)
		}

		// Execute migration
		_, err = db.ExecContext(context.Background(), string(content))
		if err != nil {
			log.Fatalf("Failed to execute migration %s: %v", filename, err)
		}
		log.Printf("Applied %s successfully", filename)
	}

	log.Println("All migrations applied successfully.")
}
