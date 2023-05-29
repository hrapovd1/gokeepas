/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func newKVCmdList(clnt *cliClient) *cobra.Command {
	getOutJSON := false
	// listCmd represents the list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Get list of secret keys from KeepPas server",
		Long: `Get list of secret keys from KeepPas server.
Default output format is text, you can change output to JSON format with flag -j.`,
		Run: func(cmd *cobra.Command, args []string) {
			runList(clnt, getOutJSON, cmd)
		},
	}
	listCmd.Flags().BoolVarP(&getOutJSON, "json", "j", false, "print output in json. Default text format.")

	return listCmd
}

func runList(client *cliClient, jsonOut bool, cmd *cobra.Command) {
	token, err := readToken(client.config.TokenCache, client.logger)
	if err != nil {
		client.logger.Sugar().Fatal(err)
	}
	client.token = token
	md := metadata.New(map[string]string{"bearer-token": token})
	cmd.SetContext(metadata.NewOutgoingContext(cmd.Context(), md))

	// getUserKey is defined in root.go
	if err := getUserKey(cmd, client); err != nil {
		client.logger.Sugar().Fatal(err)
	}
	// process request
	req := pb.BinRequest{}
	// process grpc client
	conn := client.transport(client.config.ServerAddr, client.logger)
	defer func(l *zap.Logger) {
		if err := conn.Close(); err != nil {
			l.Error(err.Error())
		}
	}(client.logger)
	transport := pb.NewKeepPasClient(conn)
	// call grpc method
	resp, err := transport.List(cmd.Context(), &req)
	if err != nil {
		client.logger.Sugar().Fatal(err)
	}
	if err := printList(resp, jsonOut, client.logger); err != nil {
		client.logger.Sugar().Fatal(err)
	}
}

func printList(resp *pb.ListResponse, jsonOut bool, log *zap.Logger) error {
	if !jsonOut {
		keys := strings.Split(resp.Keys, `','`)
		keysCount := len(keys)
		log.Sugar().Debugf("keys count: %v", keysCount)
		fmt.Println("===== Keys ======")
		for i, key := range keys {
			if i == 0 {
				if len(key) > 0 {
					key = string([]rune(key)[1:])
				}
			}
			if i == keysCount-1 {
				if len(key) > 0 {
					rKey := []rune(key)
					key = string(rKey[:len(rKey)-1])
				}
			}
			fmt.Println(key)
		}
	} else {
		keys := strings.Split(resp.Keys, `','`)
		keysCount := len(keys)
		for i, key := range keys {
			if i == 0 {
				if len(key) > 0 {
					key = string([]rune(key)[1:])
				}
			}
			if i == keysCount-1 {
				if len(key) > 0 {
					rKey := []rune(key)
					key = string(rKey[:len(rKey)-1])
				}
			}
			keys[i] = key
		}
		out, err := json.MarshalIndent(keys, ``, strings.Repeat(` `, indentCount))
		if err != nil {
			log.Sugar().Debug(err)
			return err
		}
		fmt.Println(string(out))
	}
	return nil
}
