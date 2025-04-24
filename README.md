
# Drift Detector

A CLI tool used to detect drift between Terraform and AWS.

---

## Approach

This tool compares the real-time configuration of an AWS EC2 instance against the configuration defined in Terraform files. It supports:

- Fetching instance data live from AWS (via instance ID)
- Loading sample JSON files (useful for testing)
- Parsing HCL files
- Performing a field-by-field comparison and reporting drift

The tool is structured into modules:
- `cmd/`: This holds our cli commands and entry point
- `internal/cloud/aws/`: Logic to fetch EC2 configuration
- `internal/parser/`: Unified interface to parse both AWS and Terraform sources
- `internal/drift/`: Core comparison logic
- `scripts/`: This holds scripts that helps with testing and development, such as `drift.sh`
- `infrastructure/`: This holds Terraform files that are used for testing

---

## 🧰 Decisions
The drift logic is implemented in `internal/drift/detector.go`. It uses a `Drift` struct to represent a detected drift: 
```go
type Drift struct {
	Name     string
	OldValue interface{}
	NewValue interface{}
}
```


### 💻 CLI
We use [Cobra](https://github.com/spf13/cobra) for a robust CLI experience.
### 🔍 Unified Parsing
All configurations (whether from AWS or Terraform) are parsed into a common struct:
```go
type EC2Config struct {
	AMI          string            `hcl:"ami,attr"`
	InstanceType string            `hcl:"instance_type,attr"`
	Tags         map[string]string `hcl:"tags,attr"`
}

type ParsedEC2Config struct {
	Name string
	Data EC2Config
}
```

### Running

You need to have AWS credentials configured in your environment. You can use the `AWS_PROFILE` environment variable to specify the profile to use or use anything configuration supported by the [AWS CLI for Go](https://docs.aws.amazon.com/cli/v1/userguide/cli-chap-install.html).

#### Compare

This command will compare the configuration of an AWS EC2 instance with the configuration defined in a Terraform file.
```bash
go run main.go compare --instance-id i-0b0c71db8f5d9497c --aws-json sample-data/ec2-instances.json --tf-path infrastructure/main.tf
```

#### Compare (live)

This command will compare the configuration of an AWS EC2 instance with the configuration defined in a Terraform file. It will fetch the instance configuration from AWS every time it is run.

```bash
go run main.go compare --instance-id i-0b0c71db8f5d9497c --tf-path infrastructure/main.tf
```

#### Drift

This command will drift the configuration of an AWS EC2 instance with the configuration defined in a Terraform file.

```bash
make drift
```


### 🧪 Testing

Testing is done using golang test module and `github.com/stretchr/testify`. 

This command will run all tests:
```bash
make test
```

This command will run all tests with verbose output:
```bash
make test-verbose
```

This command will run all tests with coverage output:
```bash
make test-coverage
```


### Limitations

- This tool only supports Terraform HCL files. It does not support Terraform JSON files.
- This tool only supports AWS EC2 instances.
- This tool doesn't display drifts information in very readable format.
- This tool is unable to compare for multiple instances at the same time.