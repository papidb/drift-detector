package drift

import (
	"fmt"

	"github.com/papidb/drift-detector/internal/types"
)

func CompareEC2Configs(old, new types.Resource) ([]types.Drift, error) {
	if old.Type != new.Type {
		return nil, fmt.Errorf("resource types do not match: %s vs %s", old.Type, new.Type)
	}

	oldData, okOld := old.Data.(map[string]interface{})
	newData, okNew := new.Data.(map[string]interface{})
	if !okOld || !okNew {
		return nil, fmt.Errorf("resource data is not a map[string]interface{}")
	}

	var drifts []types.Drift

	fields := []string{
		"instance_id",
		"ami",
		"key_name",
		"instance_type",
		"subnet_id",
		"availability_zone",
		"state",
		"private_ip",
		"public_ip",
	}

	for _, field := range fields {
		if oldData[field] != newData[field] {
			drifts = append(drifts, types.Drift{
				Name:     field,
				OldValue: oldData[field],
				NewValue: newData[field],
			})
		}
	}
	// Compare tags
	oldTags, oldOk := oldData["tags"].(map[string]string)
	newTags, newOk := newData["tags"].(map[string]string)
	if oldOk && newOk {
		if !areTagsEqual(oldTags, newTags) {
			drifts = append(drifts, types.Drift{
				Name:     "tags",
				OldValue: oldTags,
				NewValue: newTags,
			})
		}
	} else if oldData["tags"] != newData["tags"] {
		// Fallback for when tags are not map[string]string or one is nil
		drifts = append(drifts, types.Drift{
			Name:     "tags",
			OldValue: oldData["tags"],
			NewValue: newData["tags"],
		})
	}
	return drifts, nil
}

// areTagsEqual compares two tag maps for semantic equality
func areTagsEqual(oldTags, newTags map[string]string) bool {
	if len(oldTags) != len(newTags) {
		return false
	}

	for key, oldValue := range oldTags {
		newValue, exists := newTags[key]
		if !exists || oldValue != newValue {
			return false
		}
	}

	return true
}
