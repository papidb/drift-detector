# Drift Detector

A CLI tool used to detect configuration drift between Terraform state files and AWS EC2 instances.

---

## Overview

The Drift Detector CLI compares the configuration of AWS EC2 instances (fetched live from AWS or loaded from JSON files) against the configuration defined in Terraform state files. It identifies and reports configuration drifts, helping users ensure their infrastructure aligns with their Infrastructure as Code (IaC) definitions.

---

## Tasks to be Completed

The following tasks outline the development and enhancement roadmap for the Drift Detector:

1. **Core Functionality**:
   - Implement comparison logic for EC2 instance attributes (e.g., AMI, instance type, tags).
   - Support parsing Terraform state files and AWS EC2 instance data into a unified format.
   - Develop a CLI command (`compare`) to trigger drift detection and report results.

2. **Multiple Instance Support**:
   - Modify the `compare` command to accept multiple instance IDs via the `--instance-ids` flag.
   - Filter Terraform and AWS resources to process only the specified instance IDs.

3. **Output Formats**:
   - Support multiple output formats (console) using the `--output` flag.
   - Enhance console output readability with color-coded drift indicators (green for AWS values, red for Terraform values).

4. **Testing and Validation**:
   - Write unit tests for all packages (`cmd`, `internal/drift`, `internal/parser`, `pkg/*`).
   - Achieve at least 80% test coverage across the codebase.
   - Include integration tests using sample Terraform state and AWS JSON files.

5. **Future Enhancements**:
   - Support additional AWS resources (e.g., S3, RDS) beyond EC2.
   - Add support for Terraform JSON files alongside HCL/state files.
   - Implement a `--fix` command to generate Terraform code to reconcile drifts.
   - Improve output readability with tabular or markdown formats.

---

## Approach

The Drift Detector is designed to be modular, extensible, and testable. The approach focuses on:

1. **Unified Data Model**:
   - Parse both Terraform state files and AWS EC2 instance data into a common `types.Resource` struct, enabling consistent comparison logic.
   - Use `types.Drift` to represent detected configuration differences.

2. **Modular Architecture**:
   - Separate concerns into distinct packages:
     - `cmd/`: CLI entry point and command logic.
     - `internal/drift/`: Core drift comparison logic.
     - `internal/parser/`: Parsing Terraform and AWS data.
     - `pkg/cloud/aws/repository/`: Fetching AWS EC2 instance data.
     - `pkg/printer/`: Output formatting (console, JSON, etc.).
     - `pkg/file/`: File reading utilities.
     - `pkg/logger/`: Structured logging with `logrus`.

3. **CLI**:
   - Use Cobra for a robust CLI with flags for instance IDs, file paths, and output formats.
   - Support both live AWS data fetching and local JSON files for testing.

4. **Resource Filtering**:
   - Allow users to specify multiple instance IDs via `--instance-ids`.
   - Filter resources before comparison to reduce processing overhead and focus on relevant instances.

5. **Extensibility**:
   - Design interfaces (`parser.Parser`, `printer.Printer`, `drift.DriftComparator`) to support future resource types and output formats.
   - Use dependency injection in `AppConfig` to facilitate testing and mocking.

---

## Decisions

Key decisions made during development to ensure functionality, maintainability, and scalability:

1. **CLI Framework**:
   - **Choice**: Use [Cobra](https://github.com/spf13/cobra) for CLI implementation.
   - **Reason**: Cobra provides a robust framework for command hierarchies, flags, and help text, simplifying CLI development and maintenance.

2. **Unified Resource Model**:
   - **Choice**: Define `types.Resource` and `types.Drift` in `internal/types` for all configurations.
   - **Reason**: A common data model simplifies comparison logic and enables extensibility for additional resource types.

3. **Multiple Instance IDs**:
   - **Choice**: Replace `--instance-id` (single ID) with `--instance-ids` (multiple IDs, comma-separated or repeated flags).
   - **Reason**: Supports comparing multiple EC2 instances in one run, improving usability for users managing multiple instances.
   - **Implementation**: Filter resources in `compareResources` using a `map[string]struct{}` for efficient lookup.

4. **Output Handling**:
   - **Choice**: Implement a `printer.Printer` interface with a `ConsolePrinter` for colorful, human-readable output.
   - **Reason**: Decouples output formatting from comparison logic, allowing easy addition of new formats (e.g., JSON, HTML).
   - **Detail**: Console output uses `github.com/fatih/color` for green (AWS) and red (Terraform) indicators.

5. **Logging**:
   - **Choice**: Use `github.com/sirupsen/logrus` with a custom `logger.Logger` interface supporting structured fields.
   - **Reason**: Structured logging with fields enhances debugging and monitoring capabilities.

6. **Dependency Injection**:
   - **Choice**: Use `AppConfig` to hold dependencies (`FileReader`, `Parser`, `DriftPrinter`, etc.) with defaults set in `NewAppConfig`.
   - **Reason**: Facilitates unit testing by allowing mocks to be injected, improving test isolation.

7. **AWS SDK**:
   - **Choice**: Use `github.com/aws/aws-sdk-go` for live AWS data fetching.
   - **Reason**: Official SDK provides reliable access to AWS APIs, supporting authentication via environment variables or profiles.

---

## Limitations

Current limitations of the Drift Detector, with potential mitigations:

1. **EC2-Only Support**:
   - **Limitation**: Only EC2 instances are supported.
   - **Mitigation**: Extend `drift.DriftComparator` and `parser.Parser` to support other AWS resources (e.g., S3, RDS) in future iterations.

2. **Terraform State Files Only**:
   - **Limitation**: Only Terraform state files tfstate are supported.
   - **Mitigation**: Add support for parsing Terraform module files or converting them to state files via `terraform plan`.

3. **Output Readability**:
   - **Limitation**: Console output, while color-coded, may not be sufficiently structured for complex drifts.
   - **Mitigation**: Implement tabular or markdown output formats and enhance JSON output for machine-readable reports.

4. **Single Resource Type per Run**:
   - **Limitation**: The tool processes one resource type (e.g., `aws_instance`) per run, even with multiple instance IDs.
   - **Mitigation**: Extend comparison logic to handle multiple resource types in a single run.

5. **No Automated Drift Resolution**:
   - **Limitation**: The tool only detects drifts, not resolves them.
   - **Mitigation**: Add a `--fix` command to generate Terraform code to align AWS configurations with Terraform state.

6. **Performance with Large Resource Sets**:
   - **Limitation**: Filtering and comparison may be slow for large numbers of resources or instance IDs.
   - **Mitigation**: Optimize filtering with indexing or parallel processing for large-scale environments.

---

## Architecture

The Drift Detector follows a modular, layered architecture to ensure separation of concerns and testability:

1. **CLI Layer (`cmd/`)**:
   - Entry point for the CLI, implemented with Cobra.
   - Defines the `compare` command, handling flags (`--instance-ids`, `--tf-path`, `--aws-json`, `--output`).
   - Initializes `AppConfig` with dependencies and orchestrates the comparison process.

2. **Core Logic Layer**:
   - **Drift Comparison (`internal/drift/`)**:
     - Implements `DriftComparator` interface for comparing `types.Resource` pairs.
     - Produces `types.Drift` structs for detected differences.
   - **Parsing (`internal/parser/`)**:
     - Implements `Parser` interface to parse Terraform state files into `types.Resource`.
   - **AWS Data Fetching (`pkg/cloud/aws/repository/`)**:
     - Implements `EC2Repository` interface to fetch EC2 instance data (live or from JSON).
     - Supports both AWS SDK calls and local JSON files.

3. **Utility Layer (`pkg/`)**:
   - **File Reading (`pkg/file/`)**: `FileReader` interface for reading Terraform and JSON files.
   - **Logging (`pkg/logger/`)**: `Logger` interface with `logrus` backend for structured logging.
   - **Printing (`pkg/printer/`)**: `Printer` interface for outputting drifts in various formats.
   - **Common Utilities (`pkg/common/`)**: Helper functions like `GroupResourcesByType`.

4. **Data Model (`internal/types/`)**:
   - Defines `Resource`, `Drift`, `DriftGroup`, and `ResourceType` structs for unified data handling.

5. **Scripts and Infrastructure**:
   - **Scripts (`scripts/`)**: Contains `drift.sh` for testing and development tasks.
   - **Infrastructure (`infrastructure/`)**: Terraform files for creating test environments.

**Dependency Flow**:
- `cmd/` orchestrates the process, injecting dependencies into `AppConfig`.
- `AppConfig` wires together `FileReader`, `Parser`, `DriftPrinter`, `Comparator`, and `EC2RepoFactory`.
- Core logic (`drift`, `parser`, `aws/repository`) operates on `types` structs, unaware of CLI or output concerns.
- Utilities (`file`, `logger`, `printer`) provide reusable services.

**Diagram** (simplified):
```
CLI (cmd/)
  ↓
AppConfig (wires dependencies)
  ↓
[Parser] ↔ [DriftComparator] ↔ [EC2Repository]
  ↓           ↓                  ↓
[FileReader] [Logger]         [AWS SDK/JSON]
  ↓           ↓                  ↓
[Terraform] [Printer]        [AWS Data]
```

---

## Testing and Coverage

Testing is a critical component of the Drift Detector to ensure reliability and correctness.

### Testing Approach
- **Unit Tests**:
  - Cover all packages (`cmd`, `internal/drift`, `internal/parser`, `pkg/*`).
  - Use `github.com/stretchr/testify` for assertions and test utilities.
  - Mock interfaces (`FileReader`, `Parser`, `Printer`, `DriftComparator`, `EC2Repository`, `Logger`) to isolate dependencies.
- **Integration Tests**:
  - Use sample Terraform state files and AWS JSON files in `sample-data/` to test end-to-end workflows.
  - Validate against real AWS environments using test instances (requires AWS credentials).
- **Test Cases**:
  - `cmd`: Test `loadConfigs`, `compareResources`, `runCompare`, and `NewCompareCmd` for various scenarios (valid configs, errors, filtering).
  - `internal/drift`: Test drift detection logic for EC2 attributes.
  - `internal/parser`: Test parsing of Terraform state files.
  - `pkg/printer`: Test output formats (console, JSON).
  - `pkg/cloud/aws/repository`: Test live AWS fetching and JSON parsing.

### Test Commands
- **Run All Tests**:
  ```bash
  make test
  ```
  Executes `go test ./...`.

- **Run Tests with Verbose Output**:
  ```bash
  make test-verbose
  ```
  Executes `go test -v ./...`.

- **Run Tests with Coverage**:
  ```bash
  make test-coverage
  ```
  Executes `go test -cover ./...` and generates a coverage report.

### Current Coverage
- **Target**: Achieve at least 80% code coverage across all packages.
- **Current Status**: 
  - `cmd`: High coverage due to comprehensive tests for `compare` command, including resource filtering and error cases.
  - `internal/drift`: Moderate coverage; needs more edge cases for complex attribute comparisons.
  - `internal/parser`: High coverage for state file parsing; needs tests for malformed inputs.
  - `pkg/printer`: High coverage for console output; needs tests for JSON and other formats.
  - **Overall**: Estimated at ~75%; additional tests for edge cases and integration scenarios needed to reach 80%.

### Test Suite Details
- **cmd/compare_test.go**:
  - Tests `loadConfigs` for valid and invalid Terraform/AWS inputs.
  - Tests `compareResources` for drift detection, filtering by instance IDs, and error handling.
  - Tests `runCompare` for end-to-end execution and console output.
  - Tests `NewCompareCmd` for flag validation and required flags (`instance-ids`, `tf-path`).
- **Mocks**:
  - Use structs (`MockFileReader`, `MockPrinter`, etc.) to implement interfaces, avoiding external mocking libraries.
  - `MockLogger` supports structured logging with `Fields`.
- **Coverage Gaps**:
  - Edge cases for large resource sets or invalid instance IDs.
  - Integration tests for live AWS data fetching.
  - Tests for alternative output formats (JSON, HTML).

---

## Running the Tool

### Prerequisites
- **Go**: Version 1.18 or higher.
- **AWS Credentials**: Configured via environment variables, AWS CLI profiles, or shared config files (see [AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)).
- **Dependencies**:
  ```bash
  go get github.com/spf13/cobra
  go get github.com/sirupsen/logrus
  go get github.com/fatih/color
  go get github.com/aws/aws-sdk-go
  go get github.com/stretchr/testify
  ```

### Commands

#### Compare (with JSON)
Compare Terraform state against AWS EC2 instance data from a JSON file:
```bash
go run . compare --instance-ids i-123,i-456 --tf-path sample-data/terraform.tfstate --aws-json sample-data/ec2-instances.json
```

#### Compare (Live)
Fetch EC2 instance data live from AWS:
```bash
go run . compare --instance-ids i-123,i-456 --tf-path sample-data/terraform.tfstate
```

#### Drift (Test Script)
Apply intentional drifts for testing (using `scripts/drift.sh`):
```bash
make drift
```

### Example Output (Console)
```plaintext
==== Resource Type: aws_instance ====

  Resource: i-123
Kindly note that green indicates the new value in AWS and red indicates the old value in Terraform.
Detected drift in instance_type
- t2.micro
+ t3.micro
```

---

## Development

### Directory Structure
```
drift-detector/
├── cmd/                    # CLI commands and entry point
├── internal/               # Internal packages
│   ├── drift/             # Drift comparison logic
│   ├── parser/            # Terraform state parsing
│   ├── types/             # Common data models
├── pkg/                    # Reusable utilities
│   ├── cloud/aws/repository/ # AWS EC2 data fetching
│   ├── common/            # Helper functions
│   ├── file/              # File reading utilities
│   ├── logger/            # Structured logging
│   ├── parser/            # Parser interface
│   ├── printer/           # Output formatting
├── scripts/                # Development and testing scripts
├── infrastructure/         # Terraform files for testing
├── sample-data/            # Sample Terraform state and AWS JSON files
├── Makefile                # Build and test commands
└── README.md               # Project documentation
```

### Adding New Features
1. **New Resource Type**:
   - Extend `types.Resource` and `drift.DriftComparator` for new AWS resources.
   - Update `parser.Parser` to handle new resource schemas.
2. **New Output Format**:
   - Implement a new `printer.Printer` (e.g., `JSONPrinter`).
   - Register it in `printer.NewPrinter`.
3. **New Command**:
   - Add a new Cobra command in `cmd/`.
   - Update `AppConfig` with any new dependencies.

### Contributing
- Follow Go coding standards and run `gofmt` on all code.
- Write unit tests for new functionality, targeting 80% coverage.
- Update `README.md` with new features or changes.

---

## Conclusion

The Drift Detector provides a robust foundation for detecting configuration drift between Terraform and AWS EC2 instances. Its modular architecture, comprehensive test suite, and extensible design make it suitable for production use and future enhancements. By addressing current limitations and expanding resource support, the tool can become a comprehensive solution for infrastructure drift detection.

For issues, feature requests, or contributions, please open a pull request or issue on the project repository.


---

### Key Updates to the README

1. **Tasks to be Completed**:
   - Added a new section listing explicit tasks, including core functionality, multiple instance support, output formats, testing, and future enhancements.
   - Highlighted the recent addition of `--instance-ids` and resource filtering.

2. **Approach**:
   - Expanded to describe the unified data model, modular architecture, CLI experience, resource filtering, and extensibility.
   - Emphasized dependency injection and interface-based design for testability.

3. **Decisions**:
   - Added details on the decision to use `--instance-ids` for multiple instance support, including implementation details (filtering with `map[string]struct{}`).
   - Clarified choices for Cobra, unified resource model, output handling, logging, dependency injection, and AWS SDK.
   - Removed the outdated `EC2Config` and `ParsedEC2Config` structs, replacing them with references to `types.Resource` and `types.Drift`.

4. **Limitations**:
   - Updated to reflect current limitations, including EC2-only support, Terraform state file restriction, output readability, and performance concerns.
   - Added potential mitigations for each limitation to guide future development.

5. **Architecture**:
   - Provided a detailed breakdown of the layered architecture (CLI, core logic, utilities, data model, scripts).
   - Included a simplified dependency flow diagram to visualize component interactions.
   - Described the role of each package and how they interact via `AppConfig`.

6. **Testing and Coverage**:
   - Expanded to include the testing approach (unit and integration tests), test commands, and current coverage status.
   - Detailed the test suite for `cmd/compare_test.go`, including coverage of resource filtering and error cases.
   - Identified coverage gaps and areas for improvement (e.g., edge cases, integration tests).

7. **Running the Tool**:
   - Updated the `compare` command examples to use `--instance-ids` with multiple IDs.
   - Clarified AWS credential requirements and dependency installation.
   - Kept the example console output, noting the extra newline (consistent with `ConsolePrinter`).

8. **Development**:
   - Added a directory structure overview for clarity.
   - Provided guidance on adding new features (resource types, output formats, commands).
   - Included contributing guidelines emphasizing coding standards and test coverage.

---

### Notes

- **Alignment with Code Changes**:
  - The README reflects the recent change from `--instance-id` to `--instance-ids` and the resource filtering logic in `compareResources`.
  - It assumes the `ConsolePrinter` output includes an extra newline after `==== Resource Type: aws_instance ====`. If you applied the change to remove the newline (second `console.go` version), update the example output in the README by removing the empty line.

- **Assumptions**:
  - Assumes the project repository is hosted at `github.com/papidb/drift-detector`.
  - Assumes `sample-data/` contains `terraform.tfstate` and `ec2-instances.json` for testing.
  - Assumes `scripts/drift.sh` and `infrastructure/` are used for testing, as mentioned in the original README.

- **Output Format**:
  - The example console output matches the `ConsolePrinter` format. If you need a different format (e.g., JSON output example), I can add it.

- **Future Enhancements**:
  - The tasks section includes forward-looking goals (e.g., `--fix` command, additional resource types) to guide development.
  - If you have specific priorities or additional tasks, I can refine the list.

- **Testing Coverage**:
  - The coverage estimate (~75%) is based on the test suite provided. To get an accurate number, run `make test-coverage` and update the README with the result.
  - If you need help writing additional tests to reach 80% coverage, I can provide specific test cases.

If you need further refinements to the README (e.g., specific task details, additional sections, or alignment with other project changes), or if you encounter issues with the updated `compare.go` or `compare_test.go`, please share the details, and I’ll provide the necessary updates!