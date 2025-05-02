package json

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ReaderFromFilePath(path string) (io.Reader, func() error, error) {
	if !strings.EqualFold(filepath.Ext(path), ".json") {
		return nil, nil, fmt.Errorf("invalid file type: expected .json, got %s", filepath.Ext(path))
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	return file, file.Close, nil
}
