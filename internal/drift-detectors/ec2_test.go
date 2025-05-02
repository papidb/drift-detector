package drift_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/papidb/drift-detector/internal/drift-detectors"
	"github.com/papidb/drift-detector/internal/types"
)

func TestCompareEC2Configs_NoDrift(t *testing.T) {
	old := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"ami":           "ami-123",
			"instance_type": "t2.micro",
			"tags":          map[string]string{"env": "dev"},
		},
	)
	new := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{ // clone of old
			"ami":           "ami-123",
			"instance_type": "t2.micro",
			"tags":          map[string]string{"env": "dev"},
		},
	)

	drifts, err := drift.CompareEC2Configs(old, new)
	if err != nil {
		t.Errorf("Expected no error, got %+v", err)
	}

	if len(drifts) != 0 {
		t.Errorf("Expected no drift, got %+v", drifts)
	}
}

func TestCompareEC2Configs_FieldChange(t *testing.T) {
	old := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"ami":           "ami-123",
			"instance_type": "t2.micro",
			"tags":          map[string]string{"env": "dev"},
		},
	)
	new := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"ami":           "ami-456",
			"instance_type": "t2.micro",
			"tags":          map[string]string{"env": "dev"},
		},
	)

	expected := []types.Drift{
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
	old := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"tags": map[string]string{"env": "dev"},
		},
	)
	new := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"tags": map[string]string{"env": "prod"},
		},
	)

	expected := []types.Drift{
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
	old := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"tags": map[string]string{},
		},
	)
	new := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"tags": map[string]string{"env": "prod"},
		},
	)

	expected := []types.Drift{
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
	old := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"tags": map[string]string{"env": "prod"},
		},
	)
	new := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"tags": map[string]string{},
		},
	)

	expected := []types.Drift{
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
	old := types.NewResource(
		"web-server",
		types.EC2Instance,
		map[string]interface{}{
			"ami":           "ami-123",
			"instance_type": "t2.micro",
			"tags":          map[string]string{"env": "dev"},
		},
	)
	new := types.NewResource(
		"web-server-2",
		types.EC2Instance,
		map[string]interface{}{
			"ami":           "ami-456",
			"instance_type": "t2.micro",
			"tags":          map[string]string{"env": "dev"},
		},
	)

	_, err := drift.CompareEC2Configs(old, new)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "resource names do not match") {
		t.Errorf("Expected error to contain 'resource names do not match', got %+v", err)
	}
}
