package cmd

import (
	"context"
	"fmt"

	"github.com/papidb/drift-detector/internal/tui"
	"github.com/spf13/cobra"
)

var (
	instanceID  string
	awsJSONPath string
	tfPath      string
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare a Terraform EC2 config against the actual AWS EC2 instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return tui.Run(ctx, instanceID, awsJSONPath, tfPath)
	},
}

func init() {
	compareCmd.Flags().StringVarP(&instanceID, "instance-id", "i", "", "AWS EC2 instance ID")
	compareCmd.Flags().StringVarP(&awsJSONPath, "aws-json", "j", "", "Path to sample AWS EC2 JSON file")
	compareCmd.Flags().StringVarP(&tfPath, "tf-path", "t", "", "Path to Terraform HCL or state file (required)")
	compareCmd.MarkFlagRequired("tf-path")

	compareCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if instanceID == "" && awsJSONPath == "" {
			return fmt.Errorf("either --instance-id or --aws-json must be provided")
		}
		if instanceID != "" && awsJSONPath != "" {
			return fmt.Errorf("only one of --instance-id or --aws-json can be provided")
		}
		return nil
	}

	rootCmd.AddCommand(compareCmd)
}
