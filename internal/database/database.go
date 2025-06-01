package database

import "database/sql"

func OpenDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "test_database.db")
	return db, err
}
