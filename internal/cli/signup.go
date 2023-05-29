/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"fmt"

	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newSignupCmd(clnt *cliClient) *cobra.Command {
	opts := loginOptions{}
	// signupCmd represents the signup command
	signupCmd := &cobra.Command{
		Use:   "signup -u username -p password",
		Short: "Register new user on KeepPas server",
		Long:  `Register new user on KeepPas server and get authenticated token.`,
		// Run:   runSignup,
		Run: func(cmd *cobra.Command, args []string) {
			runSignup(clnt, opts, cmd)
		},
	}
	signupCmd.Flags().StringVarP(&opts.user, "username", "u", "", "login of user")
	signupCmd.Flags().StringVarP(&opts.password, "password", "p", "", "password of user")

	return signupCmd
}

func runSignup(client *cliClient, options loginOptions, cmd *cobra.Command) {
	if options.user == "" {
		if err := cmd.Help(); err != nil {
			client.logger.Sugar().Fatal(err)
		}
		return
	}
	// process grpc client
	client.logger.Sugar().Debugf("server: %s", client.config.ServerAddr)
	conn := client.transport(client.config.ServerAddr, client.logger)
	defer func(l *zap.Logger) {
		if err := conn.Close(); err != nil {
			l.Error(err.Error())
		}
	}(client.logger)
	transport := pb.NewKeepPasClient(conn)
	// call grpc method
	resp, err := transport.SignUp(cmd.Context(), &pb.AuthRequest{
		Login:    options.user,
		Password: options.password,
	})
	client.logger.Sugar().Debugf("resp: %v, err: %v", resp, err)
	if err != nil {
		client.logger.Sugar().Fatalln(err)
	}
	if resp.AuthToken == "" && resp.Error != "" {
		client.logger.Sugar().Fatal(resp.Error)
	}
	if err := writeToken(resp.AuthToken, client.config.TokenCache, client.logger); err != nil {
		client.logger.Sugar().Infof("write token error: %v", err)
	}
	fmt.Println("login success")
}
