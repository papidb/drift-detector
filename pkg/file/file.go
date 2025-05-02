package file

import (
	"fmt"
	"io"
	"os"
)

// FileReader is an interface for reading files
type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

// OSFileReader is the default file reader implementation
type OSFileReader struct{}

func (r *OSFileReader) ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}
