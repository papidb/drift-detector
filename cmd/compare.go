package cmd

import (
	"context"
	"fmt"

	"github.com/papidb/drift-detector/internal/cloud/aws"
	"github.com/papidb/drift-detector/internal/drift"
	"github.com/papidb/drift-detector/internal/parser"
	"github.com/spf13/cobra"
)

var (
	instanceID  string
	awsJSONPath string
	tfPath      string
)

func loadConfigs(ctx context.Context, instanceID, awsJSONPath, tfPath string) (parser.ParsedEC2Config, parser.ParsedEC2Config, error) {
	nilParserConfig := parser.ParsedEC2Config{}
	configs, err := parser.ParseTerraformHCLFile(tfPath)
	config := parser.EC2Config{}
	for _, cfg := range configs {
		if cfg.Name == instanceID {
			config = cfg.Data
			break
		}
	}
	if err != nil {
		return parser.ParsedEC2Config{}, parser.ParsedEC2Config{}, fmt.Errorf("failed to parse tf config: %w", err)
	}

	// Load AWS config from file or fetch
	var awsCfg *parser.EC2Config
	var name string

	if awsJSONPath != "" {
		awsFileReader, fileCloser, err := aws.ReaderFromFilePath(awsJSONPath)
		if err != nil {
			return nilParserConfig, nilParserConfig, fmt.Errorf("failed to parse aws json: %w", err)
		}

		awsCfg, err = aws.FetchEC2Instance(ctx, nil, awsFileReader, instanceID)
		if err != nil {
			return nilParserConfig, nilParserConfig, fmt.Errorf("failed to fetch aws instance: %w", err)
		}
		name = instanceID
		defer fileCloser()
	} else {
		awsConfig, err := aws.NewEC2Client(ctx)

		if err != nil {
			return nilParserConfig, nilParserConfig, fmt.Errorf("failed to create aws client: %w", err)
		}
		awsCfg, err = aws.FetchEC2Instance(ctx, awsConfig, nil, instanceID)
		if err != nil {
			return nilParserConfig, nilParserConfig, fmt.Errorf("failed to fetch aws instance: %w", err)
		}
		name = instanceID
	}

	return parser.ParsedEC2Config{Name: name, Data: config},
		parser.ParsedEC2Config{Name: name, Data: *awsCfg},
		nil
}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare a Terraform EC2 config against the actual AWS EC2 instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		// return tui.Run(ctx, instanceID, awsJSONPath, tfPath)
		oldCfg, newCfg, err := loadConfigs(ctx, instanceID, awsJSONPath, tfPath)
		if err != nil {
			return err
		}
		drifts, err := drift.CompareEC2Configs(oldCfg, newCfg)
		if err != nil {
			return err
		}
		drift.PrintDrifts(drifts)
		return nil
	},
}

func init() {
	compareCmd.Flags().StringVarP(&instanceID, "instance-id", "i", "", "AWS EC2 instance ID")
	compareCmd.Flags().StringVarP(&awsJSONPath, "aws-json", "j", "", "Path to sample AWS EC2 JSON file")
	compareCmd.Flags().StringVarP(&tfPath, "tf-path", "t", "", "Path to Terraform HCL or state file (required)")
	compareCmd.MarkFlagRequired("instance-id")
	compareCmd.MarkFlagRequired("aws-json")
	compareCmd.MarkFlagRequired("tf-path")

	// compareCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
	// 	if instanceID == "" && awsJSONPath == "" {
	// 		return fmt.Errorf("either --instance-id or --aws-json must be provided")
	// 	}
	// 	if instanceID != "" && awsJSONPath != "" {
	// 		return fmt.Errorf("only one of --instance-id or --aws-json can be provided")
	// 	}
	// 	return nil
	// }

	rootCmd.AddCommand(compareCmd)
}
