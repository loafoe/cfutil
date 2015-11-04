package cfutil

import (
	"github.com/mattes/migrate/migrate"
)

func Migrate(connectString, path string) ([]error, bool) {
	return migrate.UpSync(connectString, path)
}
