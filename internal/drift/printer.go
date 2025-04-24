package drift

import (
	"fmt"

	"github.com/fatih/color"
)

func PrintDrifts(drifts []Drift) {
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if len(drifts) == 0 {
		fmt.Println(green("No drift detected."))
		return
	}

	for _, d := range drifts {
		fmt.Println(yellow(
			fmt.Sprintf("%s %s", "Detected drift in", d.Name),
		))

		if d.OldValue != nil {
			fmt.Println(red(fmt.Sprintf(
				"- %v",
				d.OldValue,
			)))
		}
		if d.NewValue != nil {
			fmt.Println(green(fmt.Sprintf(
				"+ %v",
				d.NewValue,
			)))
		}

		fmt.Println()
	}
}
