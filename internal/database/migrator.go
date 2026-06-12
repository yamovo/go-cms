package database

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// Migration represents a database migration.
type Migration struct {
	Version     int
	Description string
	Up          func(tx *gorm.DB) error
	Down        func(tx *gorm.DB) error
}

// Migrator handles database schema migrations.
type Migrator struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrator creates a new database migrator.
func NewMigrator(db *gorm.DB) *Migrator {
	m := &Migrator{db: db}
	// Ensure migration tracking table exists.
	m.ensureTable()
	return m
}

// ensureTable creates the migrations tracking table if it doesn't exist.
func (m *Migrator) ensureTable() {
	m.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		description TEXT,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
}

// Register adds a migration to the list.
func (m *Migrator) Register(migrations ...Migration) {
	m.migrations = append(m.migrations, migrations...)
}

// Up runs all pending migrations.
func (m *Migrator) Up() error {
	applied, err := m.getApplied()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedSet := make(map[int]bool)
	for _, v := range applied {
		appliedSet[v] = true
	}

	count := 0
	for _, mig := range m.migrations {
		if appliedSet[mig.Version] {
			continue
		}

		slog.Info("applying migration", "version", mig.Version, "description", mig.Description)

		err := m.db.Transaction(func(tx *gorm.DB) error {
			if err := mig.Up(tx); err != nil {
				return err
			}
			return tx.Exec(
				"INSERT INTO schema_migrations (version, description) VALUES (?, ?)",
				mig.Version, mig.Description,
			).Error
		})

		if err != nil {
			return fmt.Errorf("migration %d failed: %w", mig.Version, err)
		}

		count++
		slog.Info("migration applied", "version", mig.Version)
	}

	if count == 0 {
		slog.Info("no pending migrations")
	} else {
		slog.Info("migrations completed", "applied", count)
	}

	return nil
}

// Down rolls back the last N migrations.
func (m *Migrator) Down(steps int) error {
	applied, err := m.getApplied()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		slog.Info("no migrations to rollback")
		return nil
	}

	// Build lookup map.
	migMap := make(map[int]Migration)
	for _, mig := range m.migrations {
		migMap[mig.Version] = mig
	}

	// Rollback from newest.
	rolled := 0
	for i := len(applied) - 1; i >= 0 && rolled < steps; i-- {
		version := applied[i]
		mig, exists := migMap[version]
		if !exists {
			return fmt.Errorf("migration %d not found in registered migrations", version)
		}

		slog.Info("rolling back migration", "version", version, "description", mig.Description)

		err := m.db.Transaction(func(tx *gorm.DB) error {
			if err := mig.Down(tx); err != nil {
				return err
			}
			return tx.Exec(
				"DELETE FROM schema_migrations WHERE version = ?", version,
			).Error
		})

		if err != nil {
			return fmt.Errorf("rollback of migration %d failed: %w", version, err)
		}

		rolled++
		slog.Info("migration rolled back", "version", version)
	}

	slog.Info("rollback completed", "rolled_back", rolled)
	return nil
}

// Status shows the current migration status.
func (m *Migrator) Status() ([]MigrationStatus, error) {
	applied, err := m.getApplied()
	if err != nil {
		return nil, err
	}

	appliedSet := make(map[int]bool)
	for _, v := range applied {
		appliedSet[v] = true
	}

	var statuses []MigrationStatus
	for _, mig := range m.migrations {
		status := MigrationStatus{
			Version:     mig.Version,
			Description: mig.Description,
			Applied:     appliedSet[mig.Version],
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// MigrationStatus represents the status of a migration.
type MigrationStatus struct {
	Version     int
	Description string
	Applied     bool
}

// getApplied returns a sorted list of applied migration versions.
func (m *Migrator) getApplied() ([]int, error) {
	var versions []int
	err := m.db.Raw("SELECT version FROM schema_migrations ORDER BY version").Scan(&versions).Error
	if err != nil {
		return nil, err
	}
	return versions, nil
}
