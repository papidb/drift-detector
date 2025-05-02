package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/papidb/drift-detector/internal/types"
	"github.com/papidb/drift-detector/pkg/cloud/aws/repository"
	"github.com/papidb/drift-detector/pkg/common"
	"github.com/papidb/drift-detector/pkg/file"
	"github.com/papidb/drift-detector/pkg/logger"
	"github.com/papidb/drift-detector/pkg/printer"
	"github.com/stretchr/testify/assert"
)

// MockFileReader is a mock implementation of file.FileReader
type MockFileReader struct {
	Data map[string][]byte
	Err  map[string]error
}

func (m *MockFileReader) ReadFile(path string) ([]byte, error) {
	if err, ok := m.Err[path]; ok {
		return nil, err
	}
	if data, ok := m.Data[path]; ok {
		return data, nil
	}
	return nil, errors.New("file not found")
}

// MockPrinter is a mock implementation of printer.Printer
type MockPrinter struct {
	Calls []struct {
		ResourceType types.ResourceType
		ResourceName string
		Drifts       []types.Drift
	}
}

func (m *MockPrinter) PrintDrifts(resourceType types.ResourceType, resourceName string, drifts []types.Drift) {
	m.Calls = append(m.Calls, struct {
		ResourceType types.ResourceType
		ResourceName string
		Drifts       []types.Drift
	}{resourceType, resourceName, drifts})
}

// MockParser is a mock implementation of parser.Parser
type MockParser struct {
	Resources []types.Resource
	Err       error
}

func (m *MockParser) ParseTerraformStateFile(data []byte) ([]types.Resource, error) {
	return m.Resources, m.Err
}

// MockDriftComparator is a mock implementation of drift.DriftComparator
type MockDriftComparator struct {
	Drifts []types.Drift
	Err    error
}

func (m *MockDriftComparator) CompareEC2Configs(old, new types.Resource) ([]types.Drift, error) {
	return m.Drifts, m.Err
}

// MockEC2Repository is a mock implementation of awsRepository.EC2Repository
type MockEC2Repository struct {
	Resources []types.Resource
	Err       error
}

func (m *MockEC2Repository) ListInstances(ctx context.Context) ([]types.Resource, error) {
	return m.Resources, m.Err
}

// MockLogger is a mock implementation of logger.Logger
type MockLogger struct {
	Logs []struct {
		Level  string
		Fields logger.Fields
		Args   []any
	}
}

func (m *MockLogger) Debug(args ...any) {
	logEntry := struct {
		Level  string
		Fields logger.Fields
		Args   []any
	}{Level: "debug"}
	if len(args) > 0 {
		if fields, ok := args[0].(logger.Fields); ok && len(args) > 1 {
			logEntry.Fields = fields
			logEntry.Args = args[1:]
			m.Logs = append(m.Logs, logEntry)
			return
		}
	}
	logEntry.Args = args
	m.Logs = append(m.Logs, logEntry)
}

func (m *MockLogger) Info(args ...any) {
	logEntry := struct {
		Level  string
		Fields logger.Fields
		Args   []any
	}{Level: "info"}
	if len(args) > 0 {
		if fields, ok := args[0].(logger.Fields); ok && len(args) > 1 {
			logEntry.Fields = fields
			logEntry.Args = args[1:]
			m.Logs = append(m.Logs, logEntry)
			return
		}
	}
	logEntry.Args = args
	m.Logs = append(m.Logs, logEntry)
}

func (m *MockLogger) Warn(args ...any) {
	logEntry := struct {
		Level  string
		Fields logger.Fields
		Args   []any
	}{Level: "warn"}
	if len(args) > 0 {
		if fields, ok := args[0].(logger.Fields); ok && len(args) > 1 {
			logEntry.Fields = fields
			logEntry.Args = args[1:]
			m.Logs = append(m.Logs, logEntry)
			return
		}
	}
	logEntry.Args = args
	m.Logs = append(m.Logs, logEntry)
}

func (m *MockLogger) Error(args ...any) {
	logEntry := struct {
		Level  string
		Fields logger.Fields
		Args   []any
	}{Level: "error"}
	if len(args) > 0 {
		if fields, ok := args[0].(logger.Fields); ok && len(args) > 1 {
			logEntry.Fields = fields
			logEntry.Args = args[1:]
			m.Logs = append(m.Logs, logEntry)
			return
		}
	}
	logEntry.Args = args
	m.Logs = append(m.Logs, logEntry)
}

func TestLoadConfigs(t *testing.T) {
	ctx := context.Background()
	log := &MockLogger{}
	config := &AppConfig{
		Logger: log,
		Options: &CompareOptions{
			TFPath:  "terraform.tfstate",
			AWSPath: "aws.json",
		},
	}

	tests := []struct {
		name           string
		fileReader     *MockFileReader
		parser         *MockParser
		ec2Repo        *MockEC2Repository
		expectedTF     []types.Resource
		expectedAWS    []types.Resource
		expectedErrMsg string
	}{
		{
			name: "valid configs",
			fileReader: &MockFileReader{
				Data: map[string][]byte{
					"terraform.tfstate": []byte(`{"resources": []}`),
					"aws.json":          []byte(`{}`),
				},
			},
			parser: &MockParser{
				Resources: []types.Resource{{Name: "i-123", Type: types.ResourceType("aws_instance")}},
			},
			ec2Repo: &MockEC2Repository{
				Resources: []types.Resource{{Name: "i-123", Type: types.ResourceType("aws_instance")}},
			},
			expectedTF:  []types.Resource{{Name: "i-123", Type: types.ResourceType("aws_instance")}},
			expectedAWS: []types.Resource{{Name: "i-123", Type: types.ResourceType("aws_instance")}},
		},
		{
			name: "invalid Terraform file",
			fileReader: &MockFileReader{
				Err: map[string]error{
					"terraform.tfstate": errors.New("file not found"),
				},
			},
			parser:         &MockParser{},
			ec2Repo:        &MockEC2Repository{},
			expectedErrMsg: "file not found",
		},
		{
			name: "Terraform parse error",
			fileReader: &MockFileReader{
				Data: map[string][]byte{
					"terraform.tfstate": []byte(`{"resources": []}`),
				},
			},
			parser: &MockParser{
				Err: errors.New("invalid state"),
			},
			ec2Repo:        &MockEC2Repository{},
			expectedErrMsg: "failed to parse tf config: invalid state",
		},
		{
			name: "AWS fetch error",
			fileReader: &MockFileReader{
				Data: map[string][]byte{
					"terraform.tfstate": []byte(`{"resources": []}`),
				},
			},
			parser: &MockParser{
				Resources: []types.Resource{},
			},
			ec2Repo: &MockEC2Repository{
				Err: errors.New("AWS API error"),
			},
			expectedErrMsg: "failed to fetch AWS instances: AWS API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.FileReader = tt.fileReader
			config.Parser = tt.parser
			config.EC2RepoFactory = func(_ *session.Session, _ string, _ file.FileReader, _ logger.Logger) repository.EC2Repository {
				return tt.ec2Repo
			}

			tfResources, awsResources, err := loadConfigs(ctx, config)

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, tfResources)
				assert.Nil(t, awsResources)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTF, tfResources)
			assert.Equal(t, tt.expectedAWS, awsResources)
		})
	}
}

func TestCompareResources(t *testing.T) {
	log := &MockLogger{}
	tfResource := types.Resource{
		Name: "i-123",
		Type: types.ResourceType("aws_instance"),
		Data: map[string]interface{}{"instance_type": "t2.micro"},
	}
	awsResource := types.Resource{
		Name: "i-123",
		Type: types.ResourceType("aws_instance"),
		Data: map[string]interface{}{"instance_type": "t3.micro"},
	}

	tests := []struct {
		name           string
		tfResources    []types.Resource
		awsResources   []types.Resource
		comparator     *MockDriftComparator
		expectedDrifts map[types.ResourceType][]types.DriftGroup
	}{
		{
			name:         "detected drifts",
			tfResources:  []types.Resource{tfResource},
			awsResources: []types.Resource{awsResource},
			comparator: &MockDriftComparator{
				Drifts: []types.Drift{{Name: "instance_type", OldValue: "t2.micro", NewValue: "t3.micro"}},
			},
			expectedDrifts: map[types.ResourceType][]types.DriftGroup{
				types.ResourceType("aws_instance"): {
					{ResourceName: "i-123", Drifts: []types.Drift{{Name: "instance_type", OldValue: "t2.micro", NewValue: "t3.micro"}}},
				},
			},
		},
		{
			name:           "no matching resources",
			tfResources:    []types.Resource{{Name: "i-456", Type: types.ResourceType("aws_instance")}},
			awsResources:   []types.Resource{awsResource},
			comparator:     &MockDriftComparator{},
			expectedDrifts: map[types.ResourceType][]types.DriftGroup{},
		},
		{
			name:         "compare error",
			tfResources:  []types.Resource{tfResource},
			awsResources: []types.Resource{awsResource},
			comparator: &MockDriftComparator{
				Err: errors.New("compare failed"),
			},
			expectedDrifts: map[types.ResourceType][]types.DriftGroup{},
		},
		{
			name:         "no drifts",
			tfResources:  []types.Resource{tfResource},
			awsResources: []types.Resource{awsResource},
			comparator: &MockDriftComparator{
				Drifts: []types.Drift{},
			},
			expectedDrifts: map[types.ResourceType][]types.DriftGroup{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareResources(tt.tfResources, tt.awsResources, tt.comparator, log)
			assert.Equal(t, tt.expectedDrifts, result)
		})
	}
}

func TestRunCompare(t *testing.T) {
	// Helper to capture console output
	captureOutput := func(f func()) string {
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		w.Close()
		os.Stdout = originalStdout
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		return strings.TrimSpace(buf.String())
	}

	log := &MockLogger{}
	config := &AppConfig{
		Logger: log,
		Options: &CompareOptions{
			TFPath: "terraform.tfstate",
		},
		OutputType: common.OutputType("console"),
	}

	tfResource := types.Resource{
		Name: "i-123",
		Type: types.ResourceType("aws_instance"),
		Data: map[string]interface{}{"instance_type": "t2.micro"},
	}
	awsResource := types.Resource{
		Name: "i-123",
		Type: types.ResourceType("aws_instance"),
		Data: map[string]interface{}{"instance_type": "t3.micro"},
	}

	tests := []struct {
		name           string
		tfResources    []types.Resource
		awsResources   []types.Resource
		comparator     *MockDriftComparator
		drifts         []types.Drift
		printer        printer.Printer
		expectedOutput string
		expectedErrMsg string
	}{
		{
			name:         "detected drifts",
			tfResources:  []types.Resource{tfResource},
			awsResources: []types.Resource{awsResource},
			comparator: &MockDriftComparator{
				Drifts: []types.Drift{{Name: "instance_type", OldValue: "t2.micro", NewValue: "t3.micro"}},
			},
			drifts:  []types.Drift{{Name: "instance_type", OldValue: "t2.micro", NewValue: "t3.micro"}},
			printer: printer.NewPrinter(common.OutputType("console")),
			expectedOutput: strings.Join([]string{
				"==== Resource Type: aws_instance ====",
				"",
				"  Resource: i-123",
				"Kindly note that green indicates the new value in AWS and red indicates the old value in Terraform.",
				"Detected drift in instance_type",
				"- t2.micro",
				"+ t3.micro",
			}, "\n"),
		},
		{
			name:           "load configs error",
			tfResources:    []types.Resource{},
			awsResources:   []types.Resource{},
			comparator:     &MockDriftComparator{},
			drifts:         []types.Drift{},
			printer:        &MockPrinter{},
			expectedErrMsg: "file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Parser = &MockParser{
				Resources: tt.tfResources,
			}
			config.EC2RepoFactory = func(_ *session.Session, _ string, reader file.FileReader, _ logger.Logger) repository.EC2Repository {
				return &MockEC2Repository{
					Resources: tt.awsResources,
				}
			}
			config.Comparator = tt.comparator
			config.DriftPrinter = tt.printer
			config.FileReader = &MockFileReader{
				Data: map[string][]byte{
					"terraform.tfstate": []byte(`{"resources": []}`),
				},
				Err: map[string]error{
					"terraform.tfstate": nil,
				},
			}
			if tt.expectedErrMsg != "" {
				config.FileReader = &MockFileReader{
					Err: map[string]error{
						"terraform.tfstate": errors.New("file not found"),
					},
				}
			}

			output := captureOutput(func() {
				err := runCompare(config)
				if tt.expectedErrMsg != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				} else {
					assert.NoError(t, err)
				}
			})

			if tt.expectedErrMsg == "" {
				assert.Equal(t, tt.expectedOutput, output)
				if mockPrinter, ok := config.DriftPrinter.(*MockPrinter); ok && len(tt.drifts) > 0 {
					assert.Len(t, mockPrinter.Calls, 1)
					assert.Equal(t, types.ResourceType("aws_instance"), mockPrinter.Calls[0].ResourceType)
					assert.Equal(t, "i-123", mockPrinter.Calls[0].ResourceName)
					assert.Equal(t, tt.drifts, mockPrinter.Calls[0].Drifts)
				}
			}
		})
	}
}

func TestNewCompareCmd(t *testing.T) {
	cmd := NewCompareCmd()

	// Verify command metadata
	assert.Equal(t, "compare", cmd.Use)
	assert.Equal(t, "Compare a Terraform EC2 config against the actual AWS EC2 instance", cmd.Short)

	// Verify flags
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("instance-id"))
	assert.NotNil(t, flags.Lookup("aws-json"))
	assert.NotNil(t, flags.Lookup("tf-path"))
	assert.NotNil(t, flags.Lookup("output"))

	// Verify default output
	outputFlag, _ := flags.GetString("output")
	assert.Equal(t, "console", outputFlag)

	// Verify required flags
	// Instead of checking annotations, test that the flags are marked as required by attempting to run without them
	err := cmd.ValidateRequiredFlags()
	assert.Error(t, err, "expected error when required flags are not set")
	assert.Contains(t, err.Error(), "required flag(s) \"instance-id\", \"tf-path\"")
}
