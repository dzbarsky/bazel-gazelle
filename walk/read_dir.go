//go:build !darwin

package walk

import (
	"io/fs"
	"os"
)

func readDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}
