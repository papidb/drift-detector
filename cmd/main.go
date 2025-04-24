package main

import (
	"fmt"

	"github.com/papidb/drift-detector/internal/drift"
	"github.com/papidb/drift-detector/internal/parser"
)

func main() {
	println("Hello, Terraform Drift Detector!")

	// Example usage
	source := parser.ParsedEC2Config{
		Name: "app_server",
		Data: parser.EC2Config{
			AMI:          "ami-830c94e3",
			InstanceType: "t2.macro",
			Tags: map[string]string{
				"Name": "ExampleAppServerInstance",
			},
		},
	}

	target := parser.ParsedEC2Config{
		Name: "app_server",
		Data: parser.EC2Config{
			AMI:          "ami-830c94e3",
			InstanceType: "t2.micro",
			Tags: map[string]string{
				"Name": "fine-named",
				"Env":  "production",
			},
		},
	}

	drifts := drift.CompareEC2Configs(source, target)

	fmt.Println("Configuration Drifts:")
	drift.PrintDrifts(drifts)
}
