package parser_test

import (
	"testing"

	"github.com/papidb/drift-detector/pkg/parser"
)

func ParseTerraformHCLFile_ValidFile(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		resources, err := parser.ParseTerraformHCLFile("../../infrastructure/main.tf")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(resources)
	})
}

func ParseTerraformHCLFile_InvalidFile(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		_, err := parser.ParseTerraformHCLFile("../../infrastructure/invalid.tf")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		t.Log(err)
	})
}
