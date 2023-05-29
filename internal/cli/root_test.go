package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestExecute(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		Execute(NewRootCmd())
	})
}

func Test_loggerConfig(t *testing.T) {
	res := loggerConfig(zapcore.InfoLevel)
	require.NotNil(t, res)
}

func Test_getServerCert(t *testing.T) {
	logger := zap.New(nil)
	_, err := getServerCert(":5000", logger)
	require.Error(t, err)
}
