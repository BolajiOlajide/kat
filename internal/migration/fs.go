package migration

import (
	"embed"
	"io/fs"
	"os"

	"github.com/cockroachdb/errors"
)

// ErrMigrationsDirNotExist is returned when the migrations directory doesn't exist
var ErrMigrationsDirNotExist = errors.New("migrations directory does not exist")

//go:embed templates/init.tmpl
var templatesFS embed.FS

func getMigrationsFS(path string) (fs.FS, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, errors.Wrapf(ErrMigrationsDirNotExist, "path: %s", path)
	}

	return os.DirFS(path), nil
}
