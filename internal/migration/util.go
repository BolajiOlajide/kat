package migration

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"

	"github.com/cockroachdb/errors"
)

var migrationFileNameRegex = regexp.MustCompile(`^\d+_[^_]+$`)

func getMigrationsPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/migrations", wd), nil
}

func getMigrationsFS(path string) (fs.FS, error) {
	path, err := getMigrationsPath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(path)
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
