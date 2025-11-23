package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

func ApplyMigrations(db *sql.DB) error {
	migrationDir := "internal/repository/migrations"

	files, err := filepath.Glob(filepath.Join(migrationDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationDir)
	}

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		_, err = db.Exec(string(sqlBytes))
		if err != nil {
			return fmt.Errorf("apply migration %s: %w", file, err)
		}
	}

	return nil
}
