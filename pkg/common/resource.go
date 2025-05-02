package common

import "github.com/papidb/drift-detector/internal/types"

func GroupResourcesByType(resources []types.Resource) map[types.ResourceType][]types.Resource {
	grouped := make(map[types.ResourceType][]types.Resource)
	for _, res := range resources {
		grouped[res.Type] = append(grouped[res.Type], res)
	}
	return grouped
}
