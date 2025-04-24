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

func CompareEC2Configs(old parser.ParsedEC2Config, new parser.ParsedEC2Config) ([]Drift, error) {
	if old.Name != new.Name {
		return nil, fmt.Errorf("resource names do not match: %s vs %s", old.Name, new.Name)
	}
	return compareValues(reflect.ValueOf(old.Data), reflect.ValueOf(new.Data), ""), nil
}

func compareValues(oldVal, newVal reflect.Value, path string) []Drift {
	var drifts []Drift

	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	switch oldVal.Kind() {
	case reflect.Struct:
		for i := range oldVal.NumField() {
			fieldName := oldVal.Type().Field(i).Name
			subPath := fieldName
			if path != "" {
				subPath = path + "." + fieldName
			}
			drifts = append(drifts, compareValues(oldVal.Field(i), newVal.Field(i), subPath)...)
		}
	case reflect.Map:
		// Compare keys in old map
		for _, key := range oldVal.MapKeys() {
			oldEntry := oldVal.MapIndex(key)
			newEntry := newVal.MapIndex(key)
			keyStr := fmt.Sprintf(`[%q]`, key.String())
			fullPath := path + keyStr

			if !newEntry.IsValid() {
				drifts = append(drifts, Drift{
					Name:     fullPath,
					OldValue: oldEntry.Interface(),
					NewValue: nil,
				})
			} else if !reflect.DeepEqual(oldEntry.Interface(), newEntry.Interface()) {
				drifts = append(drifts, Drift{
					Name:     fullPath,
					OldValue: oldEntry.Interface(),
					NewValue: newEntry.Interface(),
				})
			}
		}
		// Compare keys in new map (that weren't in old)
		for _, key := range newVal.MapKeys() {
			if !oldVal.MapIndex(key).IsValid() {
				keyStr := fmt.Sprintf(`[%q]`, key.String())
				fullPath := path + keyStr
				drifts = append(drifts, Drift{
					Name:     fullPath,
					OldValue: nil,
					NewValue: newVal.MapIndex(key).Interface(),
				})
			}
		}
	default:
		if !reflect.DeepEqual(oldVal.Interface(), newVal.Interface()) {
			drifts = append(drifts, Drift{
				Name:     path,
				OldValue: oldVal.Interface(),
				NewValue: newVal.Interface(),
			})
		}
	}

	return drifts
}
