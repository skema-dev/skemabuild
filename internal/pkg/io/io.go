package io

import (
	"io"
	"os"
	"path/filepath"
)

func SaveToFile(fileName string, content []byte) error {
	dir := filepath.Dir(fileName)

	if err := TryMakeDir(dir); err != nil {
		return err
	}

	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	whence, _ := f.Seek(0, io.SeekEnd)
	_, err = f.WriteAt(content, whence)
	defer f.Close()
	return err
}

func TryMakeDir(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetHomePath() string {
	path := os.Getenv("SKEMA_HOME")
	if path == "" {
		panic("SKEMA_HOME not set!")
	}
	return path
}
