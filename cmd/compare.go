package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/papidb/drift-detector/internal/drift-detectors"
	"github.com/papidb/drift-detector/internal/types"
	"github.com/papidb/drift-detector/pkg"

	awsRepository "github.com/papidb/drift-detector/pkg/cloud/aws/repository"
	"github.com/papidb/drift-detector/pkg/common"
	"github.com/papidb/drift-detector/pkg/logger"
	"github.com/papidb/drift-detector/pkg/output"
	"github.com/papidb/drift-detector/pkg/parser"
	"github.com/spf13/cobra"
)

func loadConfigs(ctx context.Context, app *pkg.App) ([]types.Resource, []types.Resource, error) {
	app.Logger.Debug("Loading AWS config")
	file, err := os.Open(app.Options.TFPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open state file: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read state file: %w", err)
	}

	stateResources, err := parser.ParseTerraformStateFile(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse tf config: %w", err)
	}

	var ec2Repo awsRepository.EC2Repository

	if app.Options.AWSPath != "" {
		file, err := os.Open(app.Options.AWSPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open AWS JSON file: %w", err)
		}
		defer file.Close()

		ec2Repo = awsRepository.NewJSONEC2Repo(file)
	} else {
		ec2Repo = awsRepository.NewEC2Repo(app.Session)
	}

	awsResources, err := ec2Repo.ListInstances(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch AWS instances: %w", err)
	}

	return stateResources, awsResources, nil
}

func runCompare(app *pkg.App) error {
	ctx := context.Background()

	terraformResources, cloudResources, err := loadConfigs(ctx, app)
	if err != nil {
		fmt.Println(err)
		return err
	}

	groupedTerraform := common.GroupResourcesByType(terraformResources)
	groupedCloud := common.GroupResourcesByType(cloudResources)

	type driftGroup struct {
		ResourceName string
		Drifts       []types.Drift
	}

	var (
		mu           sync.Mutex
		driftResults = make(map[types.ResourceType][]driftGroup)
		wg           sync.WaitGroup
	)

	seenTypes := map[types.ResourceType]struct{}{}
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
				continue // or log if desired
			}

			wg.Add(1)
			go func(rt types.ResourceType, old, new types.Resource) {
				defer wg.Done()
				result, err := drift.CompareEC2Configs(old, new)
				if err != nil {
					app.Logger.Debug("Failed to compare %s: %s", rt, err)
					// optionally log
					return
				}
				if len(result) == 0 {
					return
				}
				mu.Lock()
				driftResults[rt] = append(driftResults[rt], driftGroup{
					ResourceName: old.Name,
					Drifts:       result,
				})
				mu.Unlock()
			}(resourceType, tfRes, cloudRes)
		}
	}

	wg.Wait()

	// Print grouped drift results
	o := output.NewOutput(app.Output)
	for resourceType, groups := range driftResults {
		fmt.Printf("\n==== Resource Type: %s ====\n", resourceType)
		for _, group := range groups {
			fmt.Printf("\n  Resource: %s\n", group.ResourceName)
			o.PrintDrifts(group.Drifts)
		}
	}

	return nil
}

func NewCompareCmd() *cobra.Command {
	opts := &pkg.CompareOptions{}
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

	app := pkg.NewApp(outputType, log, opts, awsSession)

	compareCmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare a Terraform EC2 config against the actual AWS EC2 instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompare(app)
		},
	}

	compareCmd.Flags().StringVarP(&opts.InstanceID, "instance-id", "i", "", "AWS EC2 instance ID")
	compareCmd.Flags().StringVarP(&opts.AWSPath, "aws-json", "j", "", "Path to sample AWS EC2 JSON file")
	compareCmd.Flags().StringVarP(&opts.TFPath, "tf-path", "t", "", "Path to Terraform HCL or state file (required)")
	var outputFormat string
	compareCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "Output format (console, json, diff, html, etc)")
	outputType = common.OutputType(outputFormat)

	compareCmd.MarkFlagRequired("instance-id")
	// compareCmd.MarkFlagRequired("aws-json")
	compareCmd.MarkFlagRequired("tf-path")
	return compareCmd
}
