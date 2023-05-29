/*
Package cli contents methods and types for KeepPas cli client.
*/
package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hrapovd1/gokeepas/internal/crypto"
	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/hrapovd1/gokeepas/internal/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func newKVCmdGet(clnt *cliClient) *cobra.Command {
	getOutJSON := false
	// getCmd represents the get command
	getCmd := &cobra.Command{
		Use:   "get KEY",
		Short: "Get secret from KeepPas server",
		Long: `Get secret from KeepPas server.
Default output format is text, you can change output to JSON format with flag -j.`,
		Run: func(cmd *cobra.Command, args []string) {
			runGet(clnt, getOutJSON, cmd, args)
		},
	}
	getCmd.Flags().BoolVarP(&getOutJSON, "json", "j", false, "print output in json. Default text format.")

	return getCmd
}

func runGet(client *cliClient, jsonOut bool, cmd *cobra.Command, args []string) {
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
	value := strings.Join(args, ``)
	req := pb.BinRequest{
		Key: value,
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
	resp, err := transport.Get(cmd.Context(), &req)
	if err != nil {
		client.logger.Sugar().Fatal(err)
	}
	if err := printValue(resp, jsonOut, client.config.UserKey, client.logger); err != nil {
		client.logger.Sugar().Fatal(err)
	}
}

func printValue(r *pb.GetResponse, jsonFmt bool, key string, l *zap.Logger) error {
	if r.Key == "" || r.Type.String() == "" {
		l.Sugar().Debug("empty response")
		return errors.New("empty response")
	}
	secret, err := crypto.DecryptKey([]byte(key), string(r.Data))
	if err != nil {
		l.Sugar().Debug(err)
		return err
	}
	switch r.Type {
	case pb.Type_TEXT:
		if err := printText(secret, jsonFmt); err != nil {
			l.Sugar().Error(err)
			return err
		}
	case pb.Type_LOGIN:
		if err := printLogin(secret, jsonFmt); err != nil {
			l.Sugar().Error(err)
			return err
		}
	case pb.Type_BINARY:
		if err := printBin(secret, jsonFmt); err != nil {
			l.Sugar().Error(err)
			return err
		}
	case pb.Type_CART:
		if err := printCart(secret, jsonFmt); err != nil {
			l.Sugar().Error(err)
			return err
		}
	}
	return nil
}

func printText(data []byte, jsonFmt bool) error {
	if !jsonFmt {
		textData := types.Text{}
		if err := json.Unmarshal(data, &textData); err != nil {
			return err
		}
		fmt.Println("====== Text ======")
		fmt.Println("Field  Value")
		fmt.Println("-----  -----")
		fmt.Print("text   ")
		fmt.Println(textData.Text)
		if len(textData.Info) > 0 && textData.Info[0] != "" {
			fmt.Println("====== Extra ======")
			for _, s := range textData.Info {
				fmt.Println(s)
			}
		}
		return nil
	}
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", strings.Repeat(" ", indentCount)); err != nil {
		return err
	}
	if _, err := out.WriteTo(os.Stdout); err != nil {
		return err
	}
	return nil
}

func printLogin(data []byte, jsonFmt bool) error {
	if !jsonFmt {
		loginData := types.Login{}
		if err := json.Unmarshal(data, &loginData); err != nil {
			return err
		}
		fmt.Println("====== Login ======")
		fmt.Println("Field      Value")
		fmt.Println("-----      -----")
		fmt.Print("login      ")
		fmt.Println(loginData.Login)
		fmt.Print("password   ")
		fmt.Println(loginData.Password)
		if len(loginData.Info) > 0 && loginData.Info[0] != "" {
			fmt.Println("====== Extra ======")
			for _, s := range loginData.Info {
				fmt.Println(s)
			}
		}
		return nil
	}
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", strings.Repeat(" ", indentCount)); err != nil {
		return err
	}
	if _, err := out.WriteTo(os.Stdout); err != nil {
		return err
	}
	return nil
}

func printBin(data []byte, jsonFmt bool) error {
	if !jsonFmt {
		binData := types.Binary{}
		if err := json.Unmarshal(data, &binData); err != nil {
			return err
		}
		fmt.Println("====== Binary ======")
		fmt.Println("Field  Value")
		fmt.Println("-----  -----")
		fmt.Print("data   ")
		fmt.Println(binData.Data)
		if len(binData.Info) > 0 {
			fmt.Println("====== Extra ======")
			for _, s := range binData.Info {
				fmt.Println(s)
			}
		}
		return nil
	}
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", strings.Repeat(" ", indentCount)); err != nil {
		return err
	}
	if _, err := out.WriteTo(os.Stdout); err != nil {
		return err
	}
	return nil
}

func printCart(data []byte, jsonFmt bool) error {
	if !jsonFmt {
		cartData := types.Cart{}
		if err := json.Unmarshal(data, &cartData); err != nil {
			return err
		}
		fmt.Println("====== Cart ======")
		fmt.Println("Field     Value")
		fmt.Println("-----     -----")
		fmt.Print("number    ")
		fmt.Println(cartData.Number)
		fmt.Print("expired   ")
		fmt.Println(cartData.Expired)
		fmt.Print("holder    ")
		fmt.Println(cartData.Holder)
		fmt.Print("cvc       ")
		fmt.Println(cartData.CVC)
		if len(cartData.Info) > 0 {
			fmt.Println("====== Extra ======")
			for _, s := range cartData.Info {
				fmt.Println(s)
			}
		}
		return nil
	}
	var out bytes.Buffer
	if err := json.Indent(&out, data, "", strings.Repeat(" ", indentCount)); err != nil {
		return err
	}
	if _, err := out.WriteTo(os.Stdout); err != nil {
		return err
	}
	return nil
}
