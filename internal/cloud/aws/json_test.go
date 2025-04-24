package aws_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/papidb/drift-detector/internal/cloud/aws"
)

func TestReaderFromFilePath(t *testing.T) {
	tempDir := t.TempDir()
	validJSONPath := filepath.Join(tempDir, "test.json")
	if err := os.WriteFile(validJSONPath, []byte(`{"key": "value"}`), 0644); err != nil {
		t.Fatalf("Failed to create test JSON file: %v", err)
	}

	invalidExtPath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(invalidExtPath, []byte("not json"), 0644); err != nil {
		t.Fatalf("Failed to create test text file: %v", err)
	}

	// Setup: Path to non-existent file
	nonexistentPath := filepath.Join(tempDir, "nonexistent.json")

	tests := []struct {
		name        string
		path        string
		wantReader  bool
		wantCloseFn bool
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid json file",
			path:        validJSONPath,
			wantReader:  true,
			wantCloseFn: true,
			wantErr:     false,
		},
		{
			name:        "invalid file extension",
			path:        invalidExtPath,
			wantReader:  false,
			wantCloseFn: false,
			wantErr:     true,
			errContains: "invalid file type: expected .json",
		},
		{
			name:        "non-existent file",
			path:        nonexistentPath,
			wantReader:  false,
			wantCloseFn: false,
			wantErr:     true,
			errContains: "failed to open file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, closeFn, err := aws.ReaderFromFilePath(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReaderFromFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ReaderFromFilePath() error = %v, want containing %q", err, tt.errContains)
				}
			}

			if (reader != nil) != tt.wantReader {
				t.Errorf("ReaderFromFilePath() reader = %v, wantReader %v", reader != nil, tt.wantReader)
			}

			if (closeFn != nil) != tt.wantCloseFn {
				t.Errorf("ReaderFromFilePath() closeFn = %v, wantCloseFn %v", closeFn != nil, tt.wantCloseFn)
			}

			// Verify the reader actually works for successful cases
			if tt.wantReader && reader != nil {
				data, err := io.ReadAll(reader)
				if err != nil {
					t.Errorf("Failed to read from returned reader: %v", err)
				}
				if len(data) == 0 {
					t.Error("Reader returned empty data")
				}

				// Test the closer function
				if closeFn != nil {
					if err := closeFn(); err != nil {
						t.Errorf("Close function returned error: %v", err)
					}
				}
			}
		})
	}
}
