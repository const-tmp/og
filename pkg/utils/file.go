package utils

import (
	"errors"
	"io/fs"
	"os"
)

func Exists(filePath string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return false, err
		}
		return false, nil
	}
	return true, f.Close()
}
