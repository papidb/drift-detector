package parser

import "github.com/papidb/drift-detector/internal/types"

// Parser is an interface for parsing Terraform state files
type Parser interface {
	ParseTerraformStateFile(data []byte) ([]types.Resource, error)
}

// DefaultParser is the default implementation of Parser
type DefaultParser struct{}

func (p *DefaultParser) ParseTerraformStateFile(data []byte) ([]types.Resource, error) {
	return ParseTerraformStateFile(data)
}

func NewParser() *DefaultParser {
	return &DefaultParser{}
}
