package cfutil

import (
	"github.com/gemnasium/migrate/migrate"
)

// Migrate() starts a database migration for the given
// database specified in `connectString`. It looks in the `path`
// for the migration files.
func Migrate(connectString, path string) ([]error, bool) {
	return migrate.UpSync(connectString, path)
}
