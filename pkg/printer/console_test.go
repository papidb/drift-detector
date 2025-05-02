package printer

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/papidb/drift-detector/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestConsolePrinter_PrintDrifts(t *testing.T) {
	// Helper function to capture console output
	captureOutput := func(f func()) string {
		// Disable color output for consistent testing
		color.NoColor = true
		// Capture stdout
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		// Restore stdout and read captured output
		w.Close()
		os.Stdout = originalStdout
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		return strings.TrimSpace(buf.String())
	}

	tests := []struct {
		name           string
		resourceType   types.ResourceType
		resourceName   string
		drifts         []types.Drift
		expectedOutput string
	}{
		{
			name:         "no drifts",
			resourceType: types.ResourceType("aws_instance"),
			resourceName: "i-123",
			drifts:       []types.Drift{},
			expectedOutput: strings.Join([]string{
				"==== Resource Type: aws_instance ====",
				"",
				"  Resource: i-123",
				"No drift detected.",
			}, "\n"),
		},
		{
			name:         "single drift",
			resourceType: types.ResourceType("aws_instance"),
			resourceName: "i-123",
			drifts: []types.Drift{
				{
					Name:     "instance_type",
					OldValue: "t2.micro",
					NewValue: "t3.micro",
				},
			},
			expectedOutput: strings.Join([]string{
				"==== Resource Type: aws_instance ====",
				"",
				"  Resource: i-123",
				"Kindly note that green indicates the new value in AWS and red indicates the old value in Terraform.",
				"Detected drift in instance_type",
				"- t2.micro",
				"+ t3.micro",
			}, "\n"),
		},
		{
			name:         "multiple drifts",
			resourceType: types.ResourceType("aws_instance"),
			resourceName: "i-123",
			drifts: []types.Drift{
				{
					Name:     "instance_type",
					OldValue: "t2.micro",
					NewValue: "t3.micro",
				},
				{
					Name:     "tags",
					OldValue: map[string]string{"Env": "prod"},
					NewValue: map[string]string{"Env": "dev"},
				},
			},
			expectedOutput: strings.Join([]string{
				"==== Resource Type: aws_instance ====",
				"",
				"  Resource: i-123",
				"Kindly note that green indicates the new value in AWS and red indicates the old value in Terraform.",
				"Detected drift in instance_type",
				"- t2.micro",
				"+ t3.micro",
				"",
				"Detected drift in tags",
				"- map[Env:prod]",
				"+ map[Env:dev]",
			}, "\n"),
		},
		{
			name:         "drift with nil values",
			resourceType: types.ResourceType("aws_instance"),
			resourceName: "i-123",
			drifts: []types.Drift{
				{
					Name:     "public_ip",
					OldValue: nil,
					NewValue: "203.0.113.1",
				},
				{
					Name:     "key_name",
					OldValue: "my-key",
					NewValue: nil,
				},
			},
			expectedOutput: strings.Join([]string{
				"==== Resource Type: aws_instance ====",
				"",
				"  Resource: i-123",
				"Kindly note that green indicates the new value in AWS and red indicates the old value in Terraform.",
				"Detected drift in public_ip",
				"+ 203.0.113.1",
				"",
				"Detected drift in key_name",
				"- my-key",
			}, "\n"),
		},
		{
			name:         "different resource type",
			resourceType: types.ResourceType("aws_s3_bucket"),
			resourceName: "my-bucket",
			drifts: []types.Drift{
				{
					Name:     "acl",
					OldValue: "private",
					NewValue: "public-read",
				},
			},
			expectedOutput: strings.Join([]string{
				"==== Resource Type: aws_s3_bucket ====",
				"",
				"  Resource: my-bucket",
				"Kindly note that green indicates the new value in AWS and red indicates the old value in Terraform.",
				"Detected drift in acl",
				"- private",
				"+ public-read",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer := NewConsolePrinter()
			actual := captureOutput(func() {
				printer.PrintDrifts(tt.resourceType, tt.resourceName, tt.drifts)
			})
			assert.Equal(t, tt.expectedOutput, actual)
		})
	}
}
