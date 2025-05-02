package output

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/papidb/drift-detector/internal/types"
)

type ConsoleOutPut struct{}

func NewConsoleOutPut() *ConsoleOutPut {
	return &ConsoleOutPut{}
}

func (o *ConsoleOutPut) PrintDrifts(drifts []types.Drift) {
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if len(drifts) == 0 {
		fmt.Println(green("No drift detected."))
		return
	}

	fmt.Println("Kindly note that green indicates the new value in AWS and red indicates the old value in Terraform.")

	for _, d := range drifts {
		fmt.Println(yellow(fmt.Sprintf("Detected drift in %s", d.Name)))

		if d.OldValue != nil {
			fmt.Println(red(fmt.Sprintf("- %v", d.OldValue)))
		}
		if d.NewValue != nil {
			fmt.Println(green(fmt.Sprintf("+ %v", d.NewValue)))
		}

		fmt.Println()
	}
}
