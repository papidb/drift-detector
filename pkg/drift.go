package pkg

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/papidb/drift-detector/pkg/common"
	"github.com/papidb/drift-detector/pkg/logger"
)

type CompareOptions struct {
	InstanceID string
	TFPath     string
	AWSPath    string
}

type App struct {
	Options *CompareOptions
	Logger  logger.Logger
	Output  common.OutputType
	Session *session.Session
}

func NewApp(
	output common.OutputType,
	logger logger.Logger,
	options *CompareOptions,
	session *session.Session,
) *App {
	return &App{
		Options: options,
		Logger:  logger,
		Output:  output,
		Session: session,
	}
}
