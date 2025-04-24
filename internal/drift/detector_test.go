package drift_test

import (
	"reflect"
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

	drifts := drift.CompareEC2Configs(old, new)

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

	drifts := drift.CompareEC2Configs(old, new)

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

	drifts := drift.CompareEC2Configs(old, new)

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

	drifts := drift.CompareEC2Configs(old, new)

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

	drifts := drift.CompareEC2Configs(old, new)

	if !reflect.DeepEqual(drifts, expected) {
		t.Errorf("Expected %+v, got %+v", expected, drifts)
	}
}
