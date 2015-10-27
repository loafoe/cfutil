package cfutil

import (
	"github.com/mattes/migrate/migrate"
)

func Migrate(connectString, path string) []error {
	errs, ok := migrate.UpSync(connectString, path)
	if !ok {
		return errs
	}
	return nil
}
