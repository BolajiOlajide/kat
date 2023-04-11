package migration

import (
	"io/fs"
	"os"

	"github.com/cockroachdb/errors"
)

func getMigrationsFS(path string) (fs.FS, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, errors.New("Directory 'migrations' does not exist")
	}

	// Open the directory and return a fs.FS for it.
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	return os.DirFS(path), nil
}
