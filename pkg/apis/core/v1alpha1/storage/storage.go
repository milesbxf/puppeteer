package storage

import (
	"io"
	"os"
)

func Store(in io.Reader, key string) error {
	f, err := os.OpenFile("/tmp/gittest/"+key, os.O_WRONLY|os.O_CREATE, 0666)
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
