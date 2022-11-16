package writer

import (
	"fmt"
	"os"
	"path/filepath"
)

func File(dir, name string, data []byte) error {
	writePath := filepath.Join(dir, name)
	f, err := os.OpenFile(writePath, os.O_WRONLY|os.O_CREATE, 0644)
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(writePath), 0755)
		if err != nil {
			return fmt.Errorf("mkdir error: %w", err)
		}

		f, err = os.Create(writePath)
		if err != nil {
			return fmt.Errorf("create error: %w", err)
		}

	}
	if err != nil {
		return fmt.Errorf("open error: %w", err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	return nil
}
