package writer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func File(path string, buf *bytes.Buffer) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return fmt.Errorf("mkdir error: %w", err)
		}

		f, err = os.Create(path)
		if err != nil {
			return fmt.Errorf("create error: %w", err)
		}

	}
	if err != nil {
		return fmt.Errorf("open error: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, buf)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	return nil
}
