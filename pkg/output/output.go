package output

import (
	"github.com/papidb/drift-detector/internal/types"
	"github.com/papidb/drift-detector/pkg/common"
)

type Output interface {
	PrintDrifts(drifts []types.Drift)
}

func NewOutput(output common.OutputType) Output {
	switch output {
	case common.OutputConsole:
		fallthrough
	default:
		return NewConsoleOutPut()
	}
}
