package platform

import (
	"errors"
	"os"
)

// EnsureDir creates the directory at dirPath along with any necessary parents.
// If the directory already exists, EnsureDir returns nil (EEXIST is ignored).
func EnsureDir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0o755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return nil
}




