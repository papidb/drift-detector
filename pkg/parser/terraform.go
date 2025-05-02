package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/papidb/drift-detector/internal/types"
	"github.com/zclconf/go-cty/cty"
)

// ParseTerraformHCLFile parses a Terraform file and extracts resources as generic data.
func ParseTerraformHCLFile(path string) ([]types.Resource, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse file: %w", diags)
	}

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

	var results []types.Resource
	for _, block := range content.Blocks {
		resourceType := block.Labels[0]
		resourceName := block.Labels[1]

		// Convert body to a hclsyntax.Body to walk attributes
		syntaxBody, ok := block.Body.(*hclsyntax.Body)
		if !ok {
			return nil, fmt.Errorf("unexpected body type for resource %s.%s", resourceType, resourceName)
		}

		data := make(map[string]interface{})
		for attrName, attr := range syntaxBody.Attributes {
			val, diag := attr.Expr.Value(nil)
			if diag.HasErrors() {
				continue // skip attributes with errors
			}
			data[attrName] = ctyToInterface(val)
		}

		results = append(results, types.NewResource(resourceName, types.ResourceType(resourceType), data))
	}

	return results, nil
}

// ParseTerraformStateFile parses a Terraform state file and extracts only the attributes of each instance.
func ParseTerraformStateFile(path string) ([]types.Resource, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open state file: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	var results []types.Resource
	resources, ok := state["resources"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no resources found in state file")
	}

	for _, r := range resources {
		resourceMap, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		resourceType, _ := resourceMap["type"].(string)

		instances, ok := resourceMap["instances"].([]interface{})
		if !ok {
			continue
		}

		for _, inst := range instances {
			instanceMap, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}

			attributes, ok := instanceMap["attributes"].(map[string]interface{})
			if !ok {
				continue
			}

			normalized := map[string]interface{}{
				"instance_id":       attributes["id"], // Terraform's "id" is the AWS instance_id
				"instance_type":     attributes["instance_type"],
				"ami":               attributes["ami"],
				"key_name":          attributes["key_name"],
				"subnet_id":         attributes["subnet_id"],
				"availability_zone": attributes["availability_zone"],
				"state":             attributes["instance_state"], // TF state doesn’t capture instance state
				"tags":              attributes["tags"],
				"private_ip":        attributes["private_ip"],
				"public_ip":         attributes["public_ip"],
				"security_groups":   attributes["vpc_security_group_ids"],
			}

			results = append(results, types.NewResource(
				fmt.Sprintf("%v", attributes["id"]),
				types.ResourceType(resourceType),
				normalized,
			))
		}
	}

	return results, nil
}

func ctyToInterface(val cty.Value) interface{} {
	if !val.IsKnown() || val.IsNull() {
		return nil
	}

	switch val.Type().FriendlyName() {
	case "string":
		return val.AsString()
	case "number":
		n, _ := val.AsBigFloat().Float64()
		return n
	case "bool":
		return val.True()
	case "list", "tuple":
		var result []interface{}
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			result = append(result, ctyToInterface(v))
		}
		return result
	case "map", "object":
		result := make(map[string]interface{})
		for it := val.ElementIterator(); it.Next(); {
			k, v := it.Element()
			result[k.AsString()] = ctyToInterface(v)
		}
		return result
	default:
		return val.GoString()
	}
}
