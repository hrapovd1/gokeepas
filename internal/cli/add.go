/*
Package cli contents methods and types for KeepPas cli client.
*/
package cli

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/hrapovd1/gokeepas/internal/crypto"
	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/hrapovd1/gokeepas/internal/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func newKVCmdAdd(clnt *cliClient) *cobra.Command {
	secrt := rawSecret{delim: `,`}
	// addCmd represents the add command
	addCmd := &cobra.Command{
		Use: `add [-d DELIM] -t TYPE [-e EXTRA] -k KEY VALUE_FIELDS

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
		Short: "Add secret on KeepPas server",
		Long:  `Add secret on KeepPas server.`,
		Run: func(cmd *cobra.Command, args []string) {
			runAdd(clnt, secrt, cmd, args)
		},
	}
	addCmd.Flags().StringVarP(&secrt.secretType, "type", "t", "", "type of secret")
	addCmd.Flags().StringVarP(&secrt.name, "key", "k", "", "name of secret")
	addCmd.Flags().StringVarP(&secrt.extra, "extra", "e", "", "extra data of secret")
	addCmd.Flags().StringVarP(&secrt.delim, "delim", "d", `,`, "values delimiter")

	return addCmd
}

func runAdd(client *cliClient, secret rawSecret, cmd *cobra.Command, args []string) {
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
	if secret.delim != " " {
		value = strings.Join(args, ``)
	}
	req, err := parseValue(client, secret, value)
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
	resp, err := transport.Add(cmd.Context(), req)
	if err != nil {
		client.logger.Sugar().Fatal(err)
	}
	if resp.Error != "" {
		client.logger.Sugar().Fatal(resp.Error)
	}
}

func parseValue(client *cliClient, rSecret rawSecret, val string) (*pb.BinRequest, error) {
	request := pb.BinRequest{Key: rSecret.name}
	switch rSecret.secretType {
	case "login":
		request.Type = pb.Type_LOGIN
		secret, err := parseLogin(val, rSecret.delim)
		if err != nil {
			return &request, err
		}
		secret.Info = []string{rSecret.extra}
		rawJSON, err := json.Marshal(secret)
		if err != nil {
			return &request, err
		}
		encData, err := crypto.EncryptKey([]byte(client.config.UserKey), rawJSON)
		if err != nil {
			return &request, err
		}
		request.Data = encData
		return &request, nil

	case "text":
		request.Type = pb.Type_TEXT
		secret := parseText(val)
		secret.Info = []string{rSecret.extra}
		rawJSON, err := json.Marshal(secret)
		if err != nil {
			return &request, err
		}
		encData, err := crypto.EncryptKey([]byte(client.config.UserKey), rawJSON)
		if err != nil {
			return &request, err
		}
		request.Data = encData
		return &request, nil

	case "bin":
		request.Type = pb.Type_BINARY
		secret := parseBin(val)
		secret.Info = []string{rSecret.extra}
		rawJSON, err := json.Marshal(secret)
		if err != nil {
			return &request, err
		}
		encData, err := crypto.EncryptKey([]byte(client.config.UserKey), rawJSON)
		if err != nil {
			return &request, err
		}
		request.Data = encData
		return &request, nil

	case "cart":
		request.Type = pb.Type_CART
		secret, err := parseCart(val, rSecret.delim)
		if err != nil {
			return &request, err
		}
		secret.Info = []string{rSecret.extra}
		rawJSON, err := json.Marshal(secret)
		if err != nil {
			return &request, err
		}
		encData, err := crypto.EncryptKey([]byte(client.config.UserKey), rawJSON)
		if err != nil {
			return &request, err
		}
		request.Data = encData
		return &request, nil

	}
	return &request, errors.New("unknown type")
}

func parseLogin(val string, delim string) (*types.Login, error) {
	data := types.Login{}
	vals := strings.Split(val, delim)
	if len(vals) < 2 {
		return &data, errors.New("not enough fields")
	}
	data.Login = vals[0]
	data.Password = vals[1]
	return &data, nil
}

func parseText(val string) *types.Text {
	data := types.Text{}
	data.Text = val
	return &data
}

func parseBin(val string) *types.Binary {
	data := types.Binary{}
	data.Data = []byte(val)
	return &data
}

func parseCart(val string, delim string) (*types.Cart, error) {
	data := types.Cart{}
	vals := strings.Split(val, delim)
	if len(vals) < 4 {
		return &data, errors.New("not enough fields")
	}
	data.Number = vals[0]
	data.Expired = vals[1]
	data.Holder = vals[2]
	data.CVC = vals[3]
	// err := json.Unmarshal([]byte(val), &data)
	return &data, nil
}
