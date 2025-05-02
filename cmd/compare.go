package cmd

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/papidb/drift-detector/internal/drift-detectors"
	"github.com/papidb/drift-detector/internal/types"
	awsRepository "github.com/papidb/drift-detector/pkg/cloud/aws/repository"
	"github.com/papidb/drift-detector/pkg/common"
	"github.com/papidb/drift-detector/pkg/file"
	"github.com/papidb/drift-detector/pkg/logger"
	"github.com/papidb/drift-detector/pkg/parser"
	"github.com/papidb/drift-detector/pkg/printer"
	"github.com/spf13/cobra"
)

type CompareOptions struct {
	InstanceID string
	TFPath     string
	AWSPath    string
	// Instances  []string
}

// AppConfig holds dependencies for the command
type AppConfig struct {
	Logger         logger.Logger
	Session        *session.Session
	Options        *CompareOptions
	OutputType     common.OutputType
	FileReader     file.FileReader
	DriftPrinter   printer.Printer
	Parser         parser.Parser
	Comparator     drift.DriftComparator
	EC2RepoFactory func(*session.Session, string, file.FileReader, logger.Logger) awsRepository.EC2Repository
}

// NewAppConfig creates a new AppConfig with default dependencies
func NewAppConfig(outputType common.OutputType, log logger.Logger, options *CompareOptions, sess *session.Session) *AppConfig {
	return &AppConfig{
		Logger:       log,
		Session:      sess,
		Options:      options,
		OutputType:   outputType,
		FileReader:   &file.OSFileReader{},
		DriftPrinter: printer.NewPrinter(outputType),
		Parser:       parser.NewParser(),
		Comparator:   drift.NewDriftComparator(),
		EC2RepoFactory: func(sess *session.Session, awsPath string, reader file.FileReader, log logger.Logger) awsRepository.EC2Repository {
			if awsPath != "" {
				data, err := reader.ReadFile(awsPath)
				if err != nil {
					log.Error("Failed to read AWS JSON file: %w", err)
					return nil
				}
				return awsRepository.NewJSONEC2Repo(bytes.NewReader(data))
			}
			return awsRepository.NewEC2Repo(sess)
		},
	}
}

// loadConfigs loads Terraform and AWS resources
func loadConfigs(ctx context.Context, config *AppConfig) ([]types.Resource, []types.Resource, error) {
	config.Logger.Debug("Loading configs")

	// Read Terraform state file
	data, err := config.FileReader.ReadFile(config.Options.TFPath)
	if err != nil {
		return nil, nil, err
	}

	stateResources, err := config.Parser.ParseTerraformStateFile(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse tf config: %w", err)
	}

	// Create EC2 repository
	ec2Repo := config.EC2RepoFactory(config.Session, config.Options.AWSPath, config.FileReader, config.Logger)
	if ec2Repo == nil {
		return nil, nil, fmt.Errorf("failed to create EC2 repository")
	}

	awsResources, err := ec2Repo.ListInstances(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch AWS instances: %w", err)
	}

	return stateResources, awsResources, nil
}

// compareResources compares Terraform and AWS resources, returning grouped drifts
func compareResources(tfResources, awsResources []types.Resource, comparator drift.DriftComparator, logger logger.Logger) map[types.ResourceType][]types.DriftGroup {
	groupedTerraform := common.GroupResourcesByType(tfResources)
	groupedCloud := common.GroupResourcesByType(awsResources)

	driftResults := make(map[types.ResourceType][]types.DriftGroup)

	seenTypes := make(map[types.ResourceType]struct{})
	for k := range groupedTerraform {
		seenTypes[k] = struct{}{}
	}
	for k := range groupedCloud {
		seenTypes[k] = struct{}{}
	}

	for resourceType := range seenTypes {
		tfResources := groupedTerraform[resourceType]
		cloudResources := groupedCloud[resourceType]

		cloudMap := make(map[string]types.Resource)
		for _, r := range cloudResources {
			cloudMap[r.Name] = r
		}

		for _, tfRes := range tfResources {
			cloudRes, ok := cloudMap[tfRes.Name]
			if !ok {
				continue
			}

			result, err := comparator.CompareEC2Configs(tfRes, cloudRes)
			if err != nil {
				logger.Debug("Failed to compare %s: %s", resourceType, err)
				continue
			}
			if len(result) > 0 {
				driftResults[resourceType] = append(driftResults[resourceType], types.DriftGroup{
					ResourceName: tfRes.Name,
					Drifts:       result,
				})
			}
		}
	}

	return driftResults
}

// runCompare executes the comparison and prints results
func runCompare(config *AppConfig) error {
	ctx := context.Background()

	tfResources, awsResources, err := loadConfigs(ctx, config)
	if err != nil {
		fmt.Println(err)
		return err
	}

	driftResults := compareResources(tfResources, awsResources, config.Comparator, config.Logger)

	for resourceType, groups := range driftResults {
		for _, group := range groups {
			config.DriftPrinter.PrintDrifts(resourceType, group.ResourceName, group.Drifts)
		}
	}

	return nil
}

// NewCompareCmd creates the Cobra command
func NewCompareCmd() *cobra.Command {
	opts := &CompareOptions{}
	var outputType common.OutputType

	log := logger.NewLogger()
	log.Info("Initializing app")

	awsSession, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Region: aws.String("us-east-1")},
	})
	if err != nil {
		log.Error("Failed to create AWS session: %w", err)
	}

	config := NewAppConfig(outputType, log, opts, awsSession)

	compareCmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare a Terraform EC2 config against the actual AWS EC2 instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			config.OutputType = common.OutputType(cmd.Flag("output").Value.String())
			config.DriftPrinter = printer.NewPrinter(outputType)
			return runCompare(config)
		},
	}

	compareCmd.Flags().StringVarP(&opts.InstanceID, "instance-id", "i", "", "AWS EC2 instance ID")
	compareCmd.Flags().StringVarP(&opts.AWSPath, "aws-json", "j", "", "Path to sample AWS EC2 JSON file")
	compareCmd.Flags().StringVarP(&opts.TFPath, "tf-path", "t", "", "Path to Terraform HCL or state file (required)")
	compareCmd.Flags().String("output", "console", "Output format (console, json, diff, html, etc)")
	compareCmd.MarkFlagRequired("instance-id")
	compareCmd.MarkFlagRequired("tf-path")

	return compareCmd
}
