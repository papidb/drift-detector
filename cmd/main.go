package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "drift-detector",
	Short: "Detect infrastructure drift between Terraform and AWS",
	Long:  `A tool to detect and report drift between Terraform-managed infrastructure and actual AWS resources.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
