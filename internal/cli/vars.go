/*
Package cli contents methods and types for KeepPas cli client.
*/
package cli

import (
	"github.com/hrapovd1/gokeepas/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const indentCount = 4
const version = "1.0"

type cliTransport func(string, *zap.Logger) *grpc.ClientConn

type cliClient struct {
	config    config.Config
	transport cliTransport
	token     string
	logger    *zap.Logger
}

type loginOptions struct {
	user     string
	password string
}

type rawSecret struct {
	secretType string
	name       string
	extra      string
	delim      string
}

var BuildTime string
