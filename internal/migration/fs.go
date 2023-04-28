package migration

import (
	"io/fs"
	"os"

	"github.com/cockroachdb/errors"
)

func getMigrationsFS(path string) (fs.FS, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, errors.Newf("Migrations directory '%s' does not exist", path)
	}

	return os.DirFS(path), nil
}
