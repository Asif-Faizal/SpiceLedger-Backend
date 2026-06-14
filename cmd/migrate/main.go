package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	"github.com/pressly/goose/v3"
	_ "github.com/go-sql-driver/mysql"
)

const migrationsDir = "migrations"

func main() {
	config := util.LoadConfig()
	dir := flag.String("dir", migrationsDir, "migrations directory")
	flag.Parse()

	db, err := sql.Open("mysql", config.DSN())
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	if err := waitForDB(db); err != nil {
		log.Fatalf("database not ready: %v", err)
	}

	goose.SetTableName("goose_migrations")
	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("set dialect: %v", err)
	}

	cmd := "up"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	switch cmd {
	case "up":
		if err := goose.Up(db, *dir); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		log.Println("migrations applied successfully")
	case "status":
		if err := goose.Status(db, *dir); err != nil {
			log.Fatalf("migrate status: %v", err)
		}
	case "down":
		if err := goose.Down(db, *dir); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		log.Println("rolled back last migration")
	case "version":
		version, err := goose.GetDBVersion(db)
		if err != nil {
			log.Fatalf("get version: %v", err)
		}
		fmt.Printf("current migration version: %d\n", version)
	default:
		log.Fatalf("unknown command %q (use: up, down, status, version)", cmd)
	}
}

func waitForDB(db *sql.DB) error {
	for i := 1; i <= 30; i++ {
		if err := db.Ping(); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timed out waiting for mysql")
}
