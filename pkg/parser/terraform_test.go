package parser

import (
	"encoding/json"
	"testing"

	"github.com/papidb/drift-detector/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestParseTerraformStateFile(t *testing.T) {
	// Sample valid Terraform state JSON
	validState := map[string]interface{}{
		"resources": []interface{}{
			map[string]interface{}{
				"type": "aws_instance",
				"instances": []interface{}{
					map[string]interface{}{
						"attributes": map[string]interface{}{
							"id":                     "i-1234567890abcdef0",
							"instance_type":          "t2.micro",
							"ami":                    "ami-12345678",
							"key_name":               "my-key",
							"subnet_id":              "subnet-12345678",
							"availability_zone":      "us-west-2a",
							"instance_state":         "running",
							"private_ip":             "10.0.0.1",
							"public_ip":              "203.0.113.1",
							"tags":                   map[string]string{"Name": "test-instance", "Env": "prod"},
							"vpc_security_group_ids": []string{"sg-12345678"},
						},
					},
				},
			},
		},
	}
	validStateJSON, _ := json.Marshal(validState)

	// Empty resources state
	emptyResourcesState := map[string]interface{}{
		"resources": []interface{}{},
	}
	emptyResourcesJSON, _ := json.Marshal(emptyResourcesState)

	// No resources field
	noResourcesState := map[string]interface{}{
		"other_field": "value",
	}
	noResourcesJSON, _ := json.Marshal(noResourcesState)

	// Invalid instances field
	invalidInstancesState := map[string]interface{}{
		"resources": []interface{}{
			map[string]interface{}{
				"type":      "aws_instance",
				"instances": "not-an-array",
			},
		},
	}
	invalidInstancesJSON, _ := json.Marshal(invalidInstancesState)

	// Missing attributes
	missingAttributesState := map[string]interface{}{
		"resources": []interface{}{
			map[string]interface{}{
				"type": "aws_instance",
				"instances": []interface{}{
					map[string]interface{}{
						"other_field": "value",
					},
				},
			},
		},
	}
	missingAttributesJSON, _ := json.Marshal(missingAttributesState)

	tests := []struct {
		name           string
		input          []byte
		expected       []types.Resource
		expectedErrMsg string
	}{
		{
			name:  "valid state file",
			input: validStateJSON,
			expected: []types.Resource{
				types.NewResource(
					"i-1234567890abcdef0",
					types.ResourceType("aws_instance"),
					map[string]interface{}{
						"instance_id":       "i-1234567890abcdef0",
						"instance_type":     "t2.micro",
						"ami":               "ami-12345678",
						"key_name":          "my-key",
						"subnet_id":         "subnet-12345678",
						"availability_zone": "us-west-2a",
						"state":             "running",
						"private_ip":        "10.0.0.1",
						"public_ip":         "203.0.113.1",
						"tags":              map[string]string{"Name": "test-instance", "Env": "prod"},
						"security_groups":   []string{"sg-12345678"},
					},
				),
			},
		},
		{
			name:           "invalid JSON",
			input:          []byte(`{invalid json`),
			expectedErrMsg: "failed to parse state file",
		},
		{
			name:     "empty resources",
			input:    emptyResourcesJSON,
			expected: []types.Resource{},
		},
		{
			name:           "no resources field",
			input:          noResourcesJSON,
			expectedErrMsg: "no resources found in state file",
		},
		{
			name:     "invalid instances field",
			input:    invalidInstancesJSON,
			expected: []types.Resource{},
		},
		{
			name:     "missing attributes",
			input:    missingAttributesJSON,
			expected: []types.Resource{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerraformStateFile(tt.input)

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
