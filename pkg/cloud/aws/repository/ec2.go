package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/papidb/drift-detector/internal/types"
)

// EC2Repository is an interface that allows us do what ever we want t
type EC2Repository interface {
	ListInstances(ctx context.Context) ([]types.Resource, error)
}

type ec2Repo struct {
	client ec2iface.EC2API
}

func NewEC2Repo(session *session.Session) *ec2Repo {
	client := ec2.New(session)
	return &ec2Repo{
		client: client,
	}
}

func (r *ec2Repo) ListInstances(ctx context.Context) ([]types.Resource, error) {
	output, err := r.client.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch EC2 instances: %w", err)
	}

	// do a transformation
	var instances []types.Resource
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, eC2InstanceToResource(instance))
		}
	}
	// return the transformed resource
	return instances, nil
}

type jsonEC2Repo struct {
	reader io.Reader
}

func NewJSONEC2Repo(reader io.Reader) *jsonEC2Repo {
	return &jsonEC2Repo{reader: reader}
}

func (r *jsonEC2Repo) ListInstances(ctx context.Context) ([]types.Resource, error) {
	var output ec2.DescribeInstancesOutput
	if err := json.NewDecoder(r.reader).Decode(&output); err != nil {
		return nil, fmt.Errorf("failed to decode JSON EC2 data: %w", err)
	}

	var instances []types.Resource
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, eC2InstanceToResource(instance))
		}
	}

	return instances, nil
}

// helpers

func eC2InstanceToResource(instance *ec2.Instance) types.Resource {
	data := map[string]interface{}{
		"instance_id":       awsString(instance.InstanceId),
		"instance_type":     awsString(instance.InstanceType),
		"ami":               awsString(instance.ImageId),
		"key_name":          awsString(instance.KeyName),
		"subnet_id":         awsString(instance.SubnetId),
		"availability_zone": awsString(instance.Placement.AvailabilityZone),
		"state":             awsString(instance.State.Name),
		"tags":              tagsToMap(instance.Tags),
		"private_ip":        awsString(instance.PrivateIpAddress),
		"public_ip":         awsString(instance.PublicIpAddress),
		"security_groups":   securityGroupsToSlice(instance.SecurityGroups),
	}

	// name := awsString(findTag(instance.Tags, "Name"))

	return types.NewResource(awsString(instance.InstanceId), types.EC2Instance, data)
}

// awsString safely dereferences an AWS SDK *string value.
func awsString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// tagsToMap converts []*ec2.Tag into a map[string]string.
func tagsToMap(tags []*ec2.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		if tag != nil && tag.Key != nil && tag.Value != nil {
			result[*tag.Key] = *tag.Value
		}
	}
	return result
}

// findTag retrieves the value of a tag with the given key (e.g., "Name").
func findTag(tags []*ec2.Tag, key string) *string {
	for _, tag := range tags {
		if tag != nil && aws.StringValue(tag.Key) == key {
			return tag.Value
		}
	}
	return nil
}

// securityGroupsToSlice extracts the group IDs from []*ec2.GroupIdentifier.
func securityGroupsToSlice(groups []*ec2.GroupIdentifier) []string {
	var result []string
	for _, group := range groups {
		if group != nil && group.GroupId != nil {
			result = append(result, *group.GroupId)
		}
	}
	return result
}
