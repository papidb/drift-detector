package drift

import (
	"fmt"
	"reflect"

	"github.com/papidb/drift-detector/internal/parser"
)

type Drift struct {
	Name     string
	OldValue interface{}
	NewValue interface{}
}

// CompareEC2Configs compares two ParsedEC2Config objects and returns a list of detected drifts.
func CompareEC2Configs(old parser.ParsedEC2Config, new parser.ParsedEC2Config) ([]Drift, error) {
	if old.Name != new.Name {
		return nil, fmt.Errorf("resource names do not match: %s vs %s", old.Name, new.Name)
	}
	return compareValues(reflect.ValueOf(old.Data), reflect.ValueOf(new.Data), ""), nil
}

// compareValues recursively compares two reflect.Values and records differences.
func compareValues(oldVal, newVal reflect.Value, path string) []Drift {
	var drifts []Drift

	// Unwrap pointers and interfaces
	for oldVal.Kind() == reflect.Ptr || oldVal.Kind() == reflect.Interface {
		if oldVal.IsNil() {
			break
		}
		oldVal = oldVal.Elem()
	}
	for newVal.Kind() == reflect.Ptr || newVal.Kind() == reflect.Interface {
		if newVal.IsNil() {
			break
		}
		newVal = newVal.Elem()
	}

	// If types differ, record drift at this path and stop recursion
	if oldVal.IsValid() && newVal.IsValid() && oldVal.Type() != newVal.Type() {
		drifts = append(drifts, Drift{
			Name:     path,
			OldValue: oldVal.Interface(),
			NewValue: newVal.Interface(),
		})
		return drifts
	}

	switch oldVal.Kind() {
	case reflect.Struct:
		// Iterate fields
		for i := 0; i < oldVal.NumField(); i++ {
			field := oldVal.Type().Field(i)
			// skip unexported fields
			if field.PkgPath != "" {
				continue
			}
			subPath := joinPath(path, field.Name)
			drifts = append(drifts, compareValues(oldVal.Field(i), newVal.Field(i), subPath)...)
		}

	case reflect.Map:
		// Keys removed or changed
		for _, key := range oldVal.MapKeys() {
			oldEntry := oldVal.MapIndex(key)
			newEntry := newVal.MapIndex(key)
			subPath := fmt.Sprintf("%s[%q]", path, key.String())

			if !newEntry.IsValid() {
				drifts = append(drifts, Drift{Name: subPath, OldValue: oldEntry.Interface(), NewValue: nil})
			} else {
				drifts = append(drifts, compareValues(oldEntry, newEntry, subPath)...)
			}
		}
		// Keys added
		for _, key := range newVal.MapKeys() {
			if !oldVal.MapIndex(key).IsValid() {
				subPath := fmt.Sprintf("%s[%q]", path, key.String())
				drifts = append(drifts, Drift{Name: subPath, OldValue: nil, NewValue: newVal.MapIndex(key).Interface()})
			}
		}

	case reflect.Slice, reflect.Array:
		oldLen := oldVal.Len()
		newLen := newVal.Len()
		maxLen := oldLen
		if newLen > maxLen {
			maxLen = newLen
		}
		for i := 0; i < maxLen; i++ {
			subPath := fmt.Sprintf("%s[%d]", path, i)
			if i >= oldLen {
				drifts = append(drifts, Drift{Name: subPath, OldValue: nil, NewValue: newVal.Index(i).Interface()})
			} else if i >= newLen {
				drifts = append(drifts, Drift{Name: subPath, OldValue: oldVal.Index(i).Interface(), NewValue: nil})
			} else {
				drifts = append(drifts, compareValues(oldVal.Index(i), newVal.Index(i), subPath)...)
			}
		}

	default:
		if !reflect.DeepEqual(oldVal.Interface(), newVal.Interface()) {
			drifts = append(drifts, Drift{Name: path, OldValue: oldVal.Interface(), NewValue: newVal.Interface()})
		}
	}

	return drifts
}

// joinPath constructs the new path segment
func joinPath(base, field string) string {
	if base == "" {
		return field
	}
	return base + "." + field
}
