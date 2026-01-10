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
	"github.com/Asif-Faizal/SpiceLedger/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	mode := "up"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	if mode == "--down" || mode == "down" {
		defaultDSN := strings.Replace(cfg.DBDSN, "dbname=spice_db", "dbname=postgres", 1)
		dbAdmin, err := sql.Open("pgx", defaultDSN)
		if err != nil {
			log.Fatalf("Failed to open admin DB connection: %v", err)
		}
		defer dbAdmin.Close()
		_, _ = dbAdmin.ExecContext(context.Background(), "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'spice_db'")
		_, err = dbAdmin.ExecContext(context.Background(), "DROP DATABASE IF EXISTS spice_db")
		if err != nil {
			log.Fatalf("Failed to drop database: %v", err)
		}
		log.Println("Database spice_db dropped.")
		return
	}

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

	// 3. Migration Tracking
	// Create migrations table if not exists
	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// 4. Check for legacy state (Tables exist but no migrations recorded)
	var count int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to count migrations: %v", err)
	}

	if count == 0 {
		// Check if 'users' table exists
		var userTableExists bool
		err = db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'users')").Scan(&userTableExists)
		if err != nil {
			log.Printf("Warning: checking users table failed: %v", err)
		}

		if userTableExists {
			log.Println("Detected existing database without migration tracking. Attempting to bootstrap...")

			// Check if 004 is applied by looking for 'grade_id' in purchase_lots
			var hasGradeID bool
			err = db.QueryRowContext(context.Background(),
				"SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='purchase_lots' AND column_name='grade_id')").Scan(&hasGradeID)

			applied := []string{
				"000_initial_schema.sql",
				"001_add_roles_and_grades.sql",
				"002_seed_admin.sql",
				"003_fix_admin_password.sql",
			}

			if hasGradeID {
				log.Println("Detected schema post-004 (grade_id exists). Marking 000-004 as applied.")
				applied = append(applied, "004_refactor_grades_to_uuid.sql")
			} else {
				log.Println("Detected schema pre-004. Marking 000-003 as applied.")
			}

			tx, err := db.Begin()
			if err != nil {
				log.Fatalf("Failed to begin transaction for bootstrap: %v", err)
			}
			stmt, _ := tx.Prepare("INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING")
			for _, v := range applied {
				if _, err := stmt.Exec(v); err != nil {
					tx.Rollback()
					log.Fatalf("Failed to record bootstrap migration %s: %v", v, err)
				}
			}
			if err := tx.Commit(); err != nil {
				log.Fatalf("Failed to commit bootstrap: %v", err)
			}
		}
	}

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
		// Check if applied
		var exists bool
		err := db.QueryRowContext(context.Background(), "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", filename).Scan(&exists)
		if err != nil {
			log.Fatalf("Failed to check migration status for %s: %v", filename, err)
		}

		if exists {
			log.Printf("Skipping %s (already applied)", filename)
			continue
		}

		log.Printf("Applying migration: %s", filename)
		content, err := os.ReadFile(filepath.Join("migrations", filename))
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", filename, err)
		}

		// Execute migration
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}

		if _, err = tx.ExecContext(context.Background(), string(content)); err != nil {
			tx.Rollback()
			log.Fatalf("Failed to execute migration %s: %v", filename, err)
		}

		if _, err = tx.ExecContext(context.Background(), "INSERT INTO schema_migrations (version) VALUES ($1)", filename); err != nil {
			tx.Rollback()
			log.Fatalf("Failed to record migration %s: %v", filename, err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit migration %s: %v", filename, err)
		}

		log.Printf("Applied %s successfully", filename)
	}

	log.Println("All migrations applied successfully.")
}
