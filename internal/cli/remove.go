/*
Package cli contents methods and types for KeepPas cli client.
*/
package cli

import (
	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func newKVCmdRm(clnt *cliClient) *cobra.Command {
	// rmCmd represents the remove command
	rmCmd := &cobra.Command{
		Use:   "remove KEY",
		Short: "Remove secret on KeepPas server",
		Long:  `Remove secret on KeepPas server, always return ok.`,
		Run: func(cmd *cobra.Command, args []string) {
			runRm(clnt, cmd, args)
		},
	}
	return rmCmd
}

func runRm(client *cliClient, cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		if err := cmd.Help(); err != nil {
			client.logger.Sugar().Fatal(err)
		}
		return
	}
	token, err := readToken(client.config.TokenCache, client.logger)
	if err != nil {
		client.logger.Sugar().Fatal(err)
	}
	client.token = token
	md := metadata.New(map[string]string{"bearer-token": token})
	cmd.SetContext(metadata.NewOutgoingContext(cmd.Context(), md))

	if err := getUserKey(cmd, client); err != nil {
		client.logger.Sugar().Fatal(err)
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
	resp, err := transport.Remove(cmd.Context(), &pb.BinRequest{
		Key: args[0],
	})
	if err != nil {
		client.logger.Sugar().Fatalln(err)
	}
	client.logger.Sugar().Debug(resp)
}
