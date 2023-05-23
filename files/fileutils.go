package files

import (
	"io/fs"
	"os"
)

func CountFiles(folder string) int {
	dir, err := fs.ReadDir(os.DirFS(folder), ".")
	if err != nil {
		return -1
	}

	count := 0
	for _, file := range dir {
		if file.Type().IsRegular() {
			count++
		}
	}

	return count
}
