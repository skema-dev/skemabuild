package io

import (
	"github.com/skema-dev/skemabuild/internal/pkg/console"
	"github.com/skema-dev/skemabuild/internal/pkg/http"
	"github.com/skema-dev/skemabuild/internal/pkg/pattern"
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

func GetContentFromUri(uri string) string {
	var content string
	if pattern.IsHttpUrl(uri) {
		content = http.GetTextContent(uri)
	} else if pattern.IsGithubUrl(uri) {
		console.Fatalf("please use the raw content link for github proto file")
	} else {
		data, err := os.ReadFile(uri)
		console.FatalIfError(err, "failed to read from "+uri)
		content = string(data)
	}
	return content
}
