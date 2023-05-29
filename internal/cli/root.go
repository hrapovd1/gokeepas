/*
Package cli contents methods and types for KeepPas cli client.
*/
package cli

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"os"
	"time"

	"github.com/hrapovd1/gokeepas/internal/config"
	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/spf13/cobra"
)

// Execute process all cli commands.
func Execute(root *cobra.Command) {
	cobra.CheckErr(root.Execute())
}

func NewRootCmd() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use: `keeppas [--debug] [-s SERVER_ADDRESS] [-c TOKEN_CACHE]
	SERVER_ADDRESS: KeepPas server address
	TOKEN_CACHE: path where jwt token is kept`,
		Short:   "KeepPas cli client",
		Long:    `KeepPas cli client allows keep and return secrets in/from KeepPas server.`,
		Version: version,
	}

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	var (
		srvAddr string // for persistent flag
		dbg     bool   // for persistent flag
		tcache  string // for persistent flag
		client  = cliClient{}
	)

	rootCmd.PersistentFlags().BoolVar(&dbg, "debug", false, "Turn on debug messages output.")
	rootCmd.PersistentFlags().StringVarP(&srvAddr, "server", "s", "localhost:5000", "ip/dns:port")
	rootCmd.PersistentFlags().StringVarP(&tcache, "cache", "c", home+"/.keeppas.token", "token cache")
	cobra.OnInitialize(func() {
		client.config.LogLevel = config.LoggerConfig(dbg)
		client.config.ServerAddr = srvAddr
		client.config.TokenCache = tcache
		client.logger = loggerConfig(client.config.LogLevel)
		client.transport = newGRPCConnection
	})

	rootCmd.SetVersionTemplate(version + " Build at " + BuildTime + "\n")

	kvCmd := newKVCmd()
	kvCmd.AddCommand(newKVCmdAdd(&client))
	kvCmd.AddCommand(newKVCmdCP(&client))
	kvCmd.AddCommand(newKVCmdGet(&client))
	kvCmd.AddCommand(newKVCmdRm(&client))
	kvCmd.AddCommand(newKVCmdRename(&client))
	kvCmd.AddCommand(newKVCmdUpdate(&client))
	kvCmd.AddCommand(newKVCmdList(&client))

	rootCmd.AddCommand(newSignupCmd(&client))
	rootCmd.AddCommand(newLoginCmd(&client))
	rootCmd.AddCommand(kvCmd)

	return rootCmd
}

func loggerConfig(lvl zapcore.Level) *zap.Logger {
	// create logger
	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.TimeKey = "timestamp"
	logConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logConfig.Level.SetLevel(lvl)
	logger, err := logConfig.Build()
	if err != nil {
		log.Fatalf("when create zap logger got error: %v", err)
	}
	return logger
}

func getUserKey(cmd *cobra.Command, clnt *cliClient) error {
	// process grpc client
	conn := clnt.transport(clnt.config.ServerAddr, clnt.logger)
	defer func(l *zap.Logger) {
		if err := conn.Close(); err != nil {
			l.Error(err.Error())
		}
	}(clnt.logger)
	transport := pb.NewKeepPasClient(conn)
	// call grpc method
	resp, err := transport.GetKey(cmd.Context(), &pb.BinRequest{})
	if err != nil {
		clnt.logger.Sugar().Debug(err)
		return err
	}
	if resp.Error != "" {
		clnt.logger.Sugar().Debug(resp.Error)
		return errors.New(resp.Error)
	}
	clnt.config.UserKey = string(resp.SymmKey)
	return nil
}

func getServerCert(srvAddr string, l *zap.Logger) (x509.Certificate, error) {
	cert := x509.Certificate{}
	conn, err := tls.Dial("tcp", srvAddr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return cert, err
	}
	defer func(l *zap.Logger) {
		if err := conn.Close(); err != nil {
			l.Error(err.Error())
		}
	}(l)
	cert = *conn.ConnectionState().PeerCertificates[0]
	return cert, nil
}

func newGRPCConnection(srvAddr string, l *zap.Logger) *grpc.ClientConn {
	// get server cert
	srvCert, err := getServerCert(srvAddr, l)
	if err != nil {
		l.Fatal(err.Error())
	}
	crtPool := x509.NewCertPool()
	crtPool.AddCert(&srvCert)

	// grpc client
	creds := credentials.NewClientTLSFromCert(crtPool, "")
	conn, err := grpc.Dial(srvAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		l.Fatal(err.Error())
	}

	return conn
}
