package gziping

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func UnGzip(source, target string) (string, error) {
	reader, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return "", err
	}
	defer archive.Close()

	filename := "access.log"
	// filename := archive.Name
	target = filepath.Join(target, filename)
	writer, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return target, err
}
