package migrate

import (
	"database/sql"
	"fmt"
	"go-wordpress/internal/storage/sql/sqlc"
	"log"
	"os"
	"sort"
	"strings"
)

type Seeder struct {
	db *sql.DB
}

func NewSeeder(dbtx sqlc.DBTX) *Seeder {
	db, ok := dbtx.(*sql.DB)
	if !ok {
		log.Fatal("seeder requires *sql.DB, got incompatible type")
	}
	return &Seeder{db: db}
}

func (s *Seeder) SeederRun() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS seed_history (
			id         SERIAL PRIMARY KEY,
			seed_name  TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create seed_history table: %w", err)
	}

	entries, err := os.ReadDir("internal/storage/sql/seeds")
	if err != nil {
		return fmt.Errorf("failed to read seeds directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		seedName := entry.Name()

		var exists bool
		err := s.db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM seed_history WHERE seed_name = $1)",
			seedName,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check seed history for %s: %w", seedName, err)
		}
		if exists {
			log.Printf("⏭️  Seed already applied, skipping: %s", seedName)
			continue
		}

		content, err := os.ReadFile("internal/storage/sql/seeds/" + seedName)
		if err != nil {
			return fmt.Errorf("failed to read seed file %s: %w", seedName, err)
		}

		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", seedName, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute seed %s: %w", seedName, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO seed_history (seed_name) VALUES ($1)", seedName,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record seed %s: %w", seedName, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit seed %s: %w", seedName, err)
		}

		log.Printf("✅ Seed applied: %s", seedName)
	}

	log.Println("✅ All seeds applied successfully")
	return nil
}
