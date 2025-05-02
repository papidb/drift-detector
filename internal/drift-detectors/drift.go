package drift

import "github.com/papidb/drift-detector/internal/types"

// DriftComparator is an interface for comparing resources
type DriftComparator interface {
	CompareEC2Configs(old, new types.Resource) ([]types.Drift, error)
}

// DefaultDriftComparator is the default implementation of DriftComparator
type DefaultDriftComparator struct{}

func (c *DefaultDriftComparator) CompareEC2Configs(old, new types.Resource) ([]types.Drift, error) {
	return CompareEC2Configs(old, new)
}

func NewDriftComparator() DriftComparator {
	return &DefaultDriftComparator{}
}
