package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func InitSQLiteDatabase(p string) (*sqlx.DB, error) {
	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		log.Println("db does not exist, creating new")
	}

	db, err := sql.Open("sqlite3", p)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	driver, err := sqlite3.WithInstance(
		db,
		&sqlite3.Config{
			DatabaseName: "gomments",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("creating driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("creating migrations: %w", err)
	}

	if err := m.Up(); err != nil {
		if err.Error() != "no change" {
			return nil, fmt.Errorf("migrating up: %w", err)
		}
	}

	return sqlx.NewDb(db, "sqlite3"), nil
}
