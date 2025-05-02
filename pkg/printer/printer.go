package printer

import (
	"github.com/papidb/drift-detector/internal/types"
	"github.com/papidb/drift-detector/pkg/common"
)

type Printer interface {
	PrintDrifts(resourceType types.ResourceType, resourceName string, drifts []types.Drift)
}

func NewPrinter(output common.OutputType) Printer {
	switch output {
	case common.OutputConsole:
		fallthrough
	default:
		return NewConsolePrinter()
	}
}
