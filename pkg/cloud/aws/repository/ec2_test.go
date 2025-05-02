package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/papidb/drift-detector/internal/types"
	"github.com/stretchr/testify/assert"
)

// mockEC2Client is a mock implementation of ec2iface.EC2API for testing
type mockEC2Client struct {
	ec2iface.EC2API
	describeInstancesOutput *ec2.DescribeInstancesOutput
	describeInstancesErr    error
}

func (m *mockEC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return m.describeInstancesOutput, m.describeInstancesErr
}

func TestEC2Repo_ListInstances(t *testing.T) {
	ctx := context.Background()

	// Sample EC2 instance data
	instance := &ec2.Instance{
		InstanceId:       aws.String("i-1234567890abcdef0"),
		InstanceType:     aws.String("t2.micro"),
		ImageId:          aws.String("ami-12345678"),
		KeyName:          aws.String("my-key"),
		SubnetId:         aws.String("subnet-12345678"),
		PrivateIpAddress: aws.String("10.0.0.1"),
		PublicIpAddress:  aws.String("203.0.113.1"),
		State:            &ec2.InstanceState{Name: aws.String("running")},
		Placement:        &ec2.Placement{AvailabilityZone: aws.String("us-west-2a")},
		Tags: []*ec2.Tag{
			{Key: aws.String("Name"), Value: aws.String("test-instance")},
			{Key: aws.String("Environment"), Value: aws.String("prod")},
		},
		SecurityGroups: []*ec2.GroupIdentifier{
			{GroupId: aws.String("sg-12345678")},
		},
	}

	tests := []struct {
		name           string
		output         *ec2.DescribeInstancesOutput
		err            error
		expected       []types.Resource
		expectedErrMsg string
	}{
		{
			name: "successful fetch",
			output: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					{Instances: []*ec2.Instance{instance}},
				},
			},
			expected: []types.Resource{
				types.NewResource(
					"i-1234567890abcdef0",
					types.EC2Instance,
					map[string]interface{}{
						"instance_id":       "i-1234567890abcdef0",
						"instance_type":     "t2.micro",
						"ami":               "ami-12345678",
						"key_name":          "my-key",
						"subnet_id":         "subnet-12345678",
						"availability_zone": "us-west-2a",
						"state":             "running",
						"tags": map[string]string{
							"Name":        "test-instance",
							"Environment": "prod",
						},
						"private_ip":      "10.0.0.1",
						"public_ip":       "203.0.113.1",
						"security_groups": []string{"sg-12345678"},
					},
				),
			},
		},
		{
			name:           "AWS API error",
			err:            errors.New("AWS error"),
			expectedErrMsg: "failed to fetch EC2 instances: AWS error",
		},
		{
			name:     "no instances",
			output:   &ec2.DescribeInstancesOutput{Reservations: []*ec2.Reservation{}},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock client
			mockClient := &mockEC2Client{
				describeInstancesOutput: tt.output,
				describeInstancesErr:    tt.err,
			}

			// Create repository
			repo := &ec2Repo{client: mockClient}

			// Call ListInstances
			result, err := repo.ListInstances(ctx)

			// Check error
			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJSONEC2Repo_ListInstances(t *testing.T) {
	ctx := context.Background()

	// Sample EC2 instance data
	instance := &ec2.Instance{
		InstanceId:       aws.String("i-1234567890abcdef0"),
		InstanceType:     aws.String("t2.micro"),
		ImageId:          aws.String("ami-12345678"),
		KeyName:          aws.String("my-key"),
		SubnetId:         aws.String("subnet-12345678"),
		PrivateIpAddress: aws.String("10.0.0.1"),
		PublicIpAddress:  aws.String("203.0.113.1"),
		State:            &ec2.InstanceState{Name: aws.String("running")},
		Placement:        &ec2.Placement{AvailabilityZone: aws.String("us-west-2a")},
		Tags: []*ec2.Tag{
			{Key: aws.String("Name"), Value: aws.String("test-instance")},
		},
		SecurityGroups: []*ec2.GroupIdentifier{
			{GroupId: aws.String("sg-12345678")},
		},
	}

	// Marshal test data to JSON
	output := ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{Instances: []*ec2.Instance{instance}},
		},
	}
	jsonData, err := json.Marshal(output)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		input          string
		expected       []types.Resource
		expectedErrMsg string
	}{
		{
			name:  "valid JSON",
			input: string(jsonData),
			expected: []types.Resource{
				types.NewResource(
					"i-1234567890abcdef0",
					types.EC2Instance,
					map[string]interface{}{
						"instance_id":       "i-1234567890abcdef0",
						"instance_type":     "t2.micro",
						"ami":               "ami-12345678",
						"key_name":          "my-key",
						"subnet_id":         "subnet-12345678",
						"availability_zone": "us-west-2a",
						"state":             "running",
						"tags":              map[string]string{"Name": "test-instance"},
						"private_ip":        "10.0.0.1",
						"public_ip":         "203.0.113.1",
						"security_groups":   []string{"sg-12345678"},
					},
				),
			},
		},
		{
			name:           "invalid JSON",
			input:          `{invalid json}`,
			expectedErrMsg: "failed to decode JSON EC2 data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create JSON repository
			reader := bytes.NewReader([]byte(tt.input))
			repo := NewJSONEC2Repo(reader)

			// Call ListInstances
			result, err := repo.ListInstances(ctx)

			// Check error
			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("awsString", func(t *testing.T) {
		assert.Equal(t, "", awsString(nil))
		assert.Equal(t, "test", awsString(aws.String("test")))
	})

	t.Run("tagsToMap", func(t *testing.T) {
		tags := []*ec2.Tag{
			{Key: aws.String("Name"), Value: aws.String("test")},
			{Key: aws.String("Env"), Value: aws.String("prod")},
			nil,
		}
		expected := map[string]string{
			"Name": "test",
			"Env":  "prod",
		}
		assert.Equal(t, expected, tagsToMap(tags))
		assert.Empty(t, tagsToMap(nil))
	})

	t.Run("findTag", func(t *testing.T) {
		tags := []*ec2.Tag{
			{Key: aws.String("Name"), Value: aws.String("test")},
			{Key: aws.String("Env"), Value: aws.String("prod")},
		}
		assert.Equal(t, "test", awsString(findTag(tags, "Name")))
		assert.Nil(t, findTag(tags, "Missing"))
		assert.Nil(t, findTag(nil, "Name"))
	})

	t.Run("securityGroupsToSlice", func(t *testing.T) {
		groups := []*ec2.GroupIdentifier{
			{GroupId: aws.String("sg-1")},
			{GroupId: aws.String("sg-2")},
			nil,
		}
		expected := []string{"sg-1", "sg-2"}
		assert.Equal(t, expected, securityGroupsToSlice(groups))
		assert.Empty(t, securityGroupsToSlice(nil))
	})
}
