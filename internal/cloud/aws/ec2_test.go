package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	aws_internal "github.com/papidb/drift-detector/internal/cloud/aws"
	"github.com/papidb/drift-detector/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEC2Client struct {
	mock.Mock
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*ec2.DescribeInstancesOutput), args.Error(1)
}

func TestFetchEC2Instance(t *testing.T) {
	ctx := context.Background()
	instanceID := "random-instance-id"

	tests := []struct {
		name           string
		mockSetup      func(*MockEC2Client)
		expectedConfig *parser.EC2Config
		expectedError  string
	}{
		{
			name: "successful fetch from AWS",
			mockSetup: func(m *MockEC2Client) {
				m.On("DescribeInstances", ctx, &ec2.DescribeInstancesInput{
					InstanceIds: []string{instanceID},
				}).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									InstanceType: "t3.medium",
									Tags: []types.Tag{
										{Key: aws.String("Environment"), Value: aws.String("prod")},
									},
								},
							},
						},
					},
				}, nil)
			},
			expectedConfig: &parser.EC2Config{
				InstanceType: "t3.medium",
				Tags:         map[string]string{"Environment": "prod"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockEC2Client)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			config, err := aws_internal.FetchEC2Instance(ctx, mockClient, nil, instanceID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedConfig, config)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
