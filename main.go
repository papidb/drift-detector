package main

import (
	"github.com/papidb/drift-detector/cmd"
	"github.com/papidb/drift-detector/pkg/logger"
)

func main() {
	logger.Init()
	cmd.Execute()
}
