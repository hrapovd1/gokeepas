/*
Package cli contents methods and types for KeepPas cli client.
*/
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newLoginCmd(clnt *cliClient) *cobra.Command {
	opts := loginOptions{}
	// loginCmd represents the login command
	loginCmd := &cobra.Command{
		Use:   "login -u username -p password",
		Short: "Login existed user on KeepPas server",
		Long:  `Login existed user on KeepPas server and get authenticated token.`,
		Run: func(cmd *cobra.Command, args []string) {
			runLogin(clnt, opts, cmd)
		},
	}
	loginCmd.Flags().StringVarP(&opts.user, "username", "u", "", "login of user")
	loginCmd.Flags().StringVarP(&opts.password, "password", "p", "", "password of user")

	return loginCmd
}

func runLogin(client *cliClient, options loginOptions, cmd *cobra.Command) {
	if options.user == "" {
		if err := cmd.Help(); err != nil {
			client.logger.Sugar().Fatal(err)
		}
		return
	}
	// process grpc client
	conn := client.transport(client.config.ServerAddr, client.logger)
	defer func(l *zap.Logger) {
		if err := conn.Close(); err != nil {
			l.Error(err.Error())
		}
	}(client.logger)
	transport := pb.NewKeepPasClient(conn)
	// call grpc method
	resp, err := transport.LogIn(cmd.Context(), &pb.AuthRequest{
		Login:    options.user,
		Password: options.password,
	})
	if err != nil {
		client.logger.Sugar().Fatalln(err)
	}
	if resp.AuthToken == "" && resp.Error != "" {
		client.logger.Fatal(resp.Error)
	}
	if err := writeToken(resp.AuthToken, client.config.TokenCache, client.logger); err != nil {
		client.logger.Sugar().Infof("write token error: %v", err)
	}
	fmt.Println("login success")
}

func writeToken(tok string, path string, l *zap.Logger) error {
	tCache, err := os.Create(path)
	if err != nil {
		l.Sugar().Debug(err)
		return err
	}
	defer func(l *zap.Logger) {
		if err := tCache.Close(); err != nil {
			l.Sugar().Debug(err)
		}
	}(l)
	n, err := tCache.WriteString(tok)
	if err != nil {
		l.Sugar().Debug(err)
		return err
	}
	if len(tok) != n {
		l.Sugar().Debug(err)
		return err
	}
	return nil
}

func readToken(path string, l *zap.Logger) (string, error) {
	// check token cache file
	cacheInfo, err := os.Stat(path)
	if err != nil {
		l.Sugar().Debug(err)
		return "", err
	}
	cacheSize := cacheInfo.Size()
	if cacheSize == 0 || cacheSize > 512 {
		l.Sugar().Debugf("abnormal token cache file size: %v", cacheSize)
		return "", errors.New("wrong token cache file")
	}
	// read token from cache file
	tCache, err := os.Open(path)
	if err != nil {
		l.Sugar().Debug(err)
		return "", err
	}
	defer func(l *zap.Logger) {
		if err := tCache.Close(); err != nil {
			l.Sugar().Debug(err)
		}
	}(l)
	scanner := bufio.NewScanner(tCache)
	if ok := scanner.Scan(); !ok {
		l.Sugar().Debug(scanner.Err())
		return "", scanner.Err()
	}
	return scanner.Text(), nil
}
