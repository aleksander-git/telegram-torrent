package database_test

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

const (
	driver       = "postgres"
	migrationDir = "backend/migrations"
)

func connectToPostgres() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
		os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("USER"), os.Getenv("PASSWORD"))

	db, err := sql.Open(driver, connStr)
	if err != nil {
		return nil, fmt.Errorf("cannot connecting to db: %w", err)
	}

	return db, nil
}

func createDB(name string) error {
	db, err := connectToPostgres()
	if err != nil {
		return fmt.Errorf("connectToPostgres(): %w", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", name))
	if err != nil {
		return fmt.Errorf("Cannot create db %q: %w", name, err)
	}

	return nil
}

func dropDB(name string) error {
	db, err := connectToPostgres()
	if err != nil {
		return fmt.Errorf("connectToPostgres(): %w", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", name))
	if err != nil {
		return fmt.Errorf("Cannot drop db %q: %w", name, err)
	}

	return nil
}

func TestMain(m *testing.M) {
	dbName := "migration_test_db"
	os.Setenv("DB_NAME", dbName)

	if err := createDB(dbName); err != nil {
		slog.Error(fmt.Sprintf("createDB(%q): %s", dbName, err))
		os.Exit(1)
	}
	defer func() {
		if err := dropDB(dbName); err != nil {
			slog.Error(fmt.Sprintf("dropDB(%q): %s", dbName, err))
		}
	}()

	m.Run()
}

func TestMigration(t *testing.T) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("DB_NAME"))

	db, err := sql.Open(driver, connStr)
	if err != nil {
		t.Fatalf("cannot open db: %s", err)
	}
	defer db.Close()

	if err := goose.SetDialect(driver); err != nil {
		t.Fatalf("cannot set goose dialect: %s", err)
	}

	if err := goose.Up(db, migrationDir); err != nil {
		t.Fatalf("cannot raise migrations: %s", err)
	}
	t.Log("migrations were successfully raised")

	if err := goose.Down(db, migrationDir); err != nil {
		t.Fatalf("cannot roll back migrations: %s", err)
	}
	t.Log("migrations were successfully rolled back")
}
