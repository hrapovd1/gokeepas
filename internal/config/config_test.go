package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewServerConf(t *testing.T) {
	conf, err := NewServerConf()
	require.NoError(t, err)
	assert.Equal(t, "redis://localhost:6379/0", conf.DBdsn)
	assert.Equal(t, ":5000", conf.ServerAddr)
	assert.Equal(t, zapcore.Level(0), conf.LogLevel)
	assert.NotEmpty(t, conf.ServerKey)
}

func TestLoggerConfig(t *testing.T) {
	assert.Equal(t, zap.InfoLevel, LoggerConfig(false))
	assert.Equal(t, zap.DebugLevel, LoggerConfig(true))
}
