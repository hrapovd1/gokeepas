/*
Package cli contents methods and types for KeepPas cli client.
*/
package cli

import (
	"strings"

	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func newKVCmdUpdate(clnt *cliClient) *cobra.Command {
	secrt := rawSecret{delim: `,`}
	// updCmd represents the update command
	updCmd := &cobra.Command{
		Use: `update [-d DELIM] -t TYPE [-e EXTRA] -k KEY VALUE_FIELDS

	DELIM: values delimiter, ',' is default
	Allowed TYPE: login | text | bin | cart
	KEY: name of secret
	VALUE_FIELDS:
		login: LOGIN,PASSWORD
		text: TEXT
		bin: DATA (any binary data)
		cart: CART_NUMBER,EXPIRED DATA,HOLDER NAME,CVC
	EXTRA: any text
	`,
		Short: "Update secret on KeepPas server",
		Long: `Update secret on KeepPas server.
It rewrite all existed values of secret, if you don't provide existed field(s), they will be empty.
	`,
		// Run: runUpd,
		Run: func(cmd *cobra.Command, args []string) {
			runUpd(clnt, secrt, cmd, args)
		},
	}
	updCmd.Flags().StringVarP(&secrt.secretType, "type", "t", "", "type of secret")
	updCmd.Flags().StringVarP(&secrt.name, "key", "k", "", "name of secret")
	updCmd.Flags().StringVarP(&secrt.extra, "extra", "e", "", "extra data of secret")
	updCmd.Flags().StringVarP(&secrt.delim, "delim", "d", `,`, "values delimiter")

	return updCmd
}

func runUpd(client *cliClient, secret rawSecret, cmd *cobra.Command, args []string) {
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
	// process request
	value := strings.Join(args, ` `)
	req, err := parseValue(client, secret, value) // defined in add.go
	if err != nil {
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
	client.logger.Sugar().Debugf("call add, req: %v", req)
	// call grpc method
	resp, err := transport.Update(cmd.Context(), req)
	if err != nil {
		client.logger.Sugar().Fatal(err)
	}
	if resp.Error != "" {
		client.logger.Sugar().Fatal(resp.Error)
	}
}
