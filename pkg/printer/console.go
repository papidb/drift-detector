package printer

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/papidb/drift-detector/internal/types"
)

type ConsolePrinter struct{}

func NewConsolePrinter() *ConsolePrinter {
	return &ConsolePrinter{}
}

func (o *ConsolePrinter) PrintDrifts(resourceType types.ResourceType, resourceName string, drifts []types.Drift) {
	fmt.Printf("\n==== Resource Type: %s ====\n", resourceType)
	fmt.Printf("\n  Resource: %s\n", resourceName)

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
