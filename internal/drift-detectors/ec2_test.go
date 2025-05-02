package drift

import (
	"fmt"
	"testing"

	"github.com/papidb/drift-detector/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestCompareEC2Configs(t *testing.T) {
	// Common resource data for testing
	baseData := map[string]interface{}{
		"instance_id":       "i-1234567890abcdef0",
		"ami":               "ami-12345678",
		"key_name":          "my-key",
		"instance_type":     "t2.micro",
		"subnet_id":         "subnet-12345678",
		"availability_zone": "us-west-2a",
		"state":             "running",
		"private_ip":        "10.0.0.1",
		"public_ip":         "203.0.113.1",
		"tags": map[string]string{
			"Name": "test-instance",
			"Env":  "prod",
		},
	}

	tests := []struct {
		name           string
		oldResource    types.Resource
		newResource    types.Resource
		expectedDrifts []types.Drift
		expectedErrMsg string
	}{
		{
			name: "no changes",
			oldResource: types.Resource{
				Type: types.EC2Instance,
				Data: baseData,
			},
			newResource: types.Resource{
				Type: types.EC2Instance,
				Data: baseData,
			},
			expectedDrifts: []types.Drift{},
		},
		{
			name: "changed instance_type and tags",
			oldResource: types.Resource{
				Type: types.EC2Instance,
				Data: baseData,
			},
			newResource: types.Resource{
				Type: types.EC2Instance,
				Data: map[string]interface{}{
					"instance_id":       "i-1234567890abcdef0",
					"ami":               "ami-12345678",
					"key_name":          "my-key",
					"instance_type":     "t3.micro", // Changed
					"subnet_id":         "subnet-12345678",
					"availability_zone": "us-west-2a",
					"state":             "running",
					"private_ip":        "10.0.0.1",
					"public_ip":         "203.0.113.1",
					"tags": map[string]string{
						"Name": "test-instance",
						"Env":  "dev", // Changed
					},
				},
			},
			expectedDrifts: []types.Drift{
				{
					Name:     "instance_type",
					OldValue: "t2.micro",
					NewValue: "t3.micro",
				},
				{
					Name:     "tags",
					OldValue: map[string]string{"Name": "test-instance", "Env": "prod"},
					NewValue: map[string]string{"Name": "test-instance", "Env": "dev"},
				},
			},
		},
		{
			name: "mismatched resource types",
			oldResource: types.Resource{
				Type: types.EC2Instance,
				Data: baseData,
			},
			newResource: types.Resource{
				Type: "other-type",
				Data: baseData,
			},
			expectedErrMsg: "resource types do not match: aws_instance vs other-type",
		},
		{
			name: "invalid data type",
			oldResource: types.Resource{
				Type: types.EC2Instance,
				Data: "invalid-data",
			},
			newResource: types.Resource{
				Type: types.EC2Instance,
				Data: baseData,
			},
			expectedErrMsg: "resource data is not a map[string]interface{}",
		},
		{
			name: "nil tags in new resource",
			oldResource: types.Resource{
				Type: types.EC2Instance,
				Data: baseData,
			},
			newResource: types.Resource{
				Type: types.EC2Instance,
				Data: map[string]interface{}{
					"instance_id":       "i-1234567890abcdef0",
					"ami":               "ami-12345678",
					"key_name":          "my-key",
					"instance_type":     "t2.micro",
					"subnet_id":         "subnet-12345678",
					"availability_zone": "us-west-2a",
					"state":             "running",
					"private_ip":        "10.0.0.1",
					"public_ip":         "203.0.113.1",
					"tags":              nil,
				},
			},
			expectedDrifts: []types.Drift{
				{
					Name:     "tags",
					OldValue: map[string]string{"Name": "test-instance", "Env": "prod"},
					NewValue: nil,
				},
			},
		},
		{
			name: "tags not map[string]string",
			oldResource: types.Resource{
				Type: types.EC2Instance,
				Data: map[string]interface{}{
					"instance_id":       "i-1234567890abcdef0",
					"ami":               "ami-12345678",
					"key_name":          "my-key",
					"instance_type":     "t2.micro",
					"subnet_id":         "subnet-12345678",
					"availability_zone": "us-west-2a",
					"state":             "running",
					"private_ip":        "10.0.0.1",
					"public_ip":         "203.0.113.1",
					"tags":              "invalid-tags",
				},
			},
			newResource: types.Resource{
				Type: types.EC2Instance,
				Data: baseData,
			},
			expectedDrifts: []types.Drift{
				{
					Name:     "tags",
					OldValue: "invalid-tags",
					NewValue: map[string]string{"Name": "test-instance", "Env": "prod"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drifts, err := CompareEC2Configs(tt.oldResource, tt.newResource)

			if tt.expectedErrMsg != "" {
				fmt.Println(err)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, drifts)
				return
			}

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedDrifts, drifts)
		})
	}
}

func TestAreTagsEqual(t *testing.T) {
	tests := []struct {
		name     string
		oldTags  map[string]string
		newTags  map[string]string
		expected bool
	}{
		{
			name: "identical tags",
			oldTags: map[string]string{
				"Name": "test-instance",
				"Env":  "prod",
			},
			newTags: map[string]string{
				"Name": "test-instance",
				"Env":  "prod",
			},
			expected: true,
		},
		{
			name: "different values",
			oldTags: map[string]string{
				"Name": "test-instance",
				"Env":  "prod",
			},
			newTags: map[string]string{
				"Name": "test-instance",
				"Env":  "dev",
			},
			expected: false,
		},
		{
			name: "different keys",
			oldTags: map[string]string{
				"Name": "test-instance",
				"Env":  "prod",
			},
			newTags: map[string]string{
				"Name": "test-instance",
				"Role": "web",
			},
			expected: false,
		},
		{
			name:     "empty tags",
			oldTags:  map[string]string{},
			newTags:  map[string]string{},
			expected: true,
		},
		{
			name: "different lengths",
			oldTags: map[string]string{
				"Name": "test-instance",
			},
			newTags: map[string]string{
				"Name": "test-instance",
				"Env":  "prod",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := areTagsEqual(tt.oldTags, tt.newTags)
			assert.Equal(t, tt.expected, result)
		})
	}
}
