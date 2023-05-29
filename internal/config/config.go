/*
Package config contents type and methods for server and client configuration
*/
package config

import (
	"fmt"

	"github.com/hrapovd1/gokeepas/internal/crypto"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config is general type for server and client configuration
type Config struct {
	DBdsn      string
	ServerAddr string
	ServerKey  []byte
	UserKey    string
	TokenCache string // path to file with cli user token
	LogLevel   zapcore.Level
}

// NewServerConf generates server configuration according flags
func NewServerConf() (*Config, error) {
	conf := &Config{}
	var dbg bool
	var addr string
	var dsn string
	var srvKey string
	pflag.BoolVar(&dbg, "debug", false, "Run server with debug logging")
	pflag.StringVarP(&addr, "address", "a", ":5000", "Server ADDRESS:PORT")
	pflag.StringVarP(&dsn, "redisDSN", "d", "redis://localhost:6379/0", "Redis DB address, format: 'redis://<user>:<pass>@<ip/dns>:<port>/<db>', default: redis://localhost:6379/0 ")
	pflag.StringVarP(&srvKey, "masterkey", "k", "", "Server encryption master key. If it isn't provided, server will generates new key and prints in stdout. You need to use the same key for existed DB.")
	pflag.Parse()

	conf.LogLevel = LoggerConfig(dbg)
	conf.ServerAddr = addr
	conf.DBdsn = dsn
	conf.ServerKey = []byte(srvKey)

	if len(conf.ServerKey) == 0 {
		srvKey, err := crypto.GenServerKey(crypto.SymmKeyLength)
		if err != nil {
			return conf, err
		}
		conf.ServerKey = []byte(srvKey)
		fmt.Printf("\t!!!! Not found server key, generate new: \n\n\t\t'%v'\t\n\n\tPlease remember and provide it in the next run with this db !!!\n", srvKey)
	}

	return conf, nil
}

// LoggerConfig return log level according debug flag
func LoggerConfig(debug bool) zapcore.Level {
	if debug {
		return zap.DebugLevel
	}
	return zap.InfoLevel
}
