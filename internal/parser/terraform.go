package parser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type EC2Config struct {
	AMI          string            `hcl:"ami,attr"`
	InstanceType string            `hcl:"instance_type,attr"`
	Tags         map[string]string `hcl:"tags,attr"`
}

type ParsedEC2Config struct {
	Name string
	Data EC2Config
}

func ParseTerraformHCLFile(path string) ([]ParsedEC2Config, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse file: %w", diags)
	}

	// Prepare a schema to only decode resource blocks
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
			},
		},
	}

	content, _, diags := file.Body.PartialContent(schema)

	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to extract resource blocks: %w", diags)
	}

	results := make([]ParsedEC2Config, 0, len(content.Blocks))

	ctx := &hcl.EvalContext{}
	for _, block := range content.Blocks {
		if block.Type == "resource" && block.Labels[0] == "aws_instance" {
			var cfg EC2Config
			diags := gohcl.DecodeBody(block.Body, ctx, &cfg)
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to decode aws_instance block: %w", diags)
			}

			results = append(results, ParsedEC2Config{
				Name: block.Labels[1],
				Data: cfg,
			})
		}
	}

	return results, nil
}
