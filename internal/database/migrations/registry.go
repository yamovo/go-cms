package migrations

import "github.com/vortexcms/go-cms/internal/database"

var pendingMigrations []database.Migration

// RegisterMigrations adds migrations to the pending list.
func RegisterMigrations(migs ...database.Migration) {
	pendingMigrations = append(pendingMigrations, migs...)
}

// GetAll returns all registered migrations.
func GetAll() []database.Migration {
	return pendingMigrations
}
