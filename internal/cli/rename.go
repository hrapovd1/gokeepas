/*
Package cli contents methods and types for KeepPas cli client.
*/
package cli

import (
	"fmt"
	"strings"

	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func newKVCmdRename(clnt *cliClient) *cobra.Command {
	delim := " "
	// mvCmd represents the rename command
	mvCmd := &cobra.Command{
		Use:   "rename OLD_KEY NEW_KEY",
		Short: "Rename secret on KeepPas server",
		Long:  `Rename secret on KeepPas server. It uses space as delimeter, but you can change it with flag -d.`,
		Run: func(cmd *cobra.Command, args []string) {
			runRename(clnt, delim, cmd, args)
		},
	}
	mvCmd.Flags().StringVarP(&delim, "delim", "d", ` `, "key names delimiter")
	return mvCmd
}

func runRename(client *cliClient, dlmtr string, cmd *cobra.Command, args []string) {
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

	if dlmtr == " " && len(args) < 2 {
		if err := cmd.Help(); err != nil {
			client.logger.Sugar().Debug(err)
		}
		fmt.Println("not enough parameters")
		return
	}
	values := make([]string, len(args))
	copy(values, args)
	if dlmtr != " " {
		value := strings.Join(args, ``)
		values = strings.Split(value, dlmtr)
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
	resp, err := transport.Rename(cmd.Context(), &pb.BinRequest{
		Key:    values[0],
		NewKey: values[1],
	})
	if err != nil {
		client.logger.Sugar().Fatalln(err)
	}
	client.logger.Sugar().Debug(resp)
}
