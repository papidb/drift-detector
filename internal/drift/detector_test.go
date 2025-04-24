package drift_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/papidb/drift-detector/internal/drift"
	"github.com/papidb/drift-detector/internal/parser"
)

func TestCompareEC2Configs_NoDrift(t *testing.T) {
	old := parser.ParsedEC2Config{
		Name: "web-server",
		Data: parser.EC2Config{
			AMI:          "ami-123",
			InstanceType: "t2.micro",
			Tags:         map[string]string{"env": "dev"},
		},
	}
	new := old // identical

	drifts, err := drift.CompareEC2Configs(old, new)
	if err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if len(drifts) != 0 {
		t.Errorf("Expected no drift, got %+v", drifts)
	}
}

func TestCompareEC2Configs_FieldChange(t *testing.T) {
	old := parser.ParsedEC2Config{
		Name: "web-server",
		Data: parser.EC2Config{
			AMI:          "ami-123",
			InstanceType: "t2.micro",
			Tags:         map[string]string{"env": "dev"},
		},
	}
	new := parser.ParsedEC2Config{
		Name: "web-server",
		Data: parser.EC2Config{
			AMI:          "ami-456",
			InstanceType: "t2.micro",
			Tags:         map[string]string{"env": "dev"},
		},
	}

	expected := []drift.Drift{
		{
			Name:     "AMI",
			OldValue: "ami-123",
			NewValue: "ami-456",
		},
	}

	drifts, err := drift.CompareEC2Configs(old, new)
	if err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if !reflect.DeepEqual(drifts, expected) {
		t.Errorf("Expected %+v, got %+v", expected, drifts)
	}
}

func TestCompareEC2Configs_TagValueChanged(t *testing.T) {
	old := parser.ParsedEC2Config{
		Data: parser.EC2Config{
			Tags: map[string]string{"env": "dev"},
		},
	}
	new := parser.ParsedEC2Config{
		Data: parser.EC2Config{
			Tags: map[string]string{"env": "prod"},
		},
	}

	expected := []drift.Drift{
		{
			Name:     "Tags[\"env\"]",
			OldValue: "dev",
			NewValue: "prod",
		},
	}

	drifts, err := drift.CompareEC2Configs(old, new)
	if err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if !reflect.DeepEqual(drifts, expected) {
		t.Errorf("Expected %+v, got %+v", expected, drifts)
	}
}

func TestCompareEC2Configs_TagAdded(t *testing.T) {
	old := parser.ParsedEC2Config{
		Data: parser.EC2Config{
			Tags: map[string]string{},
		},
	}
	new := parser.ParsedEC2Config{
		Data: parser.EC2Config{
			Tags: map[string]string{"env": "prod"},
		},
	}

	expected := []drift.Drift{
		{
			Name:     "Tags[\"env\"]",
			OldValue: nil,
			NewValue: "prod",
		},
	}

	drifts, err := drift.CompareEC2Configs(old, new)
	if err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if !reflect.DeepEqual(drifts, expected) {
		t.Errorf("Expected %+v, got %+v", expected, drifts)
	}
}

func TestCompareEC2Configs_TagRemoved(t *testing.T) {
	old := parser.ParsedEC2Config{
		Data: parser.EC2Config{
			Tags: map[string]string{"env": "prod"},
		},
	}
	new := parser.ParsedEC2Config{
		Data: parser.EC2Config{
			Tags: map[string]string{},
		},
	}

	expected := []drift.Drift{
		{
			Name:     "Tags[\"env\"]",
			OldValue: "prod",
			NewValue: nil,
		},
	}

	drifts, err := drift.CompareEC2Configs(old, new)
	if err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if !reflect.DeepEqual(drifts, expected) {
		t.Errorf("Expected %+v, got %+v", expected, drifts)
	}
}

func TestCompareEC2Configs_NameMisMatch(t *testing.T) {
	old := parser.ParsedEC2Config{
		Name: "web-server",
		Data: parser.EC2Config{
			AMI:          "ami-123",
			InstanceType: "t2.micro",
			Tags:         map[string]string{"env": "dev"},
		},
	}
	new := parser.ParsedEC2Config{
		Name: "web-server-2",
		Data: parser.EC2Config{
			AMI:          "ami-456",
			InstanceType: "t2.micro",
			Tags:         map[string]string{"env": "dev"},
		},
	}

	_, err := drift.CompareEC2Configs(old, new)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "resource names do not match") {
		t.Errorf("Expected error to contain 'resource names do not match', got %+v", err)
	}
}
