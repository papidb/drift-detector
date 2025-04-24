package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/papidb/drift-detector/internal/parser"
)

// EC2API is an interface that matches the aws's EC2 client's DescribeInstances method
type EC2API interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// Helper to replace the actual EC2 client creation
func NewEC2Client(ctx context.Context) (*ec2.Client, error) {
	config, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(config), nil
}

func FetchEC2Instance(ctx context.Context, client EC2API, reader io.Reader, instanceID string) (*parser.EC2Config, error) {
	// read from the reader if it's not nil
	if reader != nil {
		var output *ec2.DescribeInstancesOutput
		if err := json.NewDecoder(reader).Decode(&output); err != nil {
			return nil, fmt.Errorf("failed to decode reservations from reader: %w", err)
		}

		// find reservation with the instance ID
		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				if instance.InstanceId != nil && *instance.InstanceId == instanceID {
					return transformAWSInstanceToEC2(&instance), nil
				}
			}
		}

		return nil, fmt.Errorf("no instance found in the provided JSON")
	}

	// If reader is nil, fetch the instance from AWS
	output, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 instance: %w", err)
	}

	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("no instance found for ID %s", instanceID)
	}

	return transformAWSInstanceToEC2(&output.Reservations[0].Instances[0]), nil
}

func transformAWSInstanceToEC2(instance *types.Instance) *parser.EC2Config {
	tags := make(map[string]string)
	for _, tag := range instance.Tags {
		if tag.Key != nil && tag.Value != nil {
			tags[*tag.Key] = *tag.Value
		}
	}

	return &parser.EC2Config{
		InstanceType:    aws.ToString((*string)(&instance.InstanceType)),
		Tags:            tags,
		SourceDestCheck: aws.ToBool(instance.SourceDestCheck),
		MetadataOptions: parser.MetadataOptions{
			// HttpEndpoint: aws.ToString(instance.MetadataOptions.HttpEndpoint.),
			// HttpTokens:   aws.ToString(instance.MetadataOptions.HttpTokens),
		},
	}
}
