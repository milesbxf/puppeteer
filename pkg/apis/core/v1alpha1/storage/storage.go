package storage

import (
	"io"
	"os"
	"path/filepath"
)

func Store(rootDir string, in io.Reader, key string) error {
	err := os.MkdirAll(rootDir, 0666)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(rootDir, key), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, in)
	if err != nil {
		return err
	}
	return nil
}
