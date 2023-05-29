package cli

import (
	"testing"

	"github.com/hrapovd1/gokeepas/internal/config"
	"github.com/hrapovd1/gokeepas/internal/types"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	pb "github.com/hrapovd1/gokeepas/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_runAdd(t *testing.T) {
	client := cliClient{logger: zap.New(nil), config: config.Config{}}
	secret := rawSecret{delim: `,`}
	t.Run("empty args", func(t *testing.T) {
		_ = client.logger
		runAdd(&client, secret, &cobra.Command{}, []string{})
	})
	// t.Run("empty token path", func(t *testing.T) {
	// 	client.config.TokenCache = ""
	// 	client.config.ServerAddr = ":5000"
	// 	client.transport = newGRPCConnection
	// 	// conn := client.transport(client.config.ServerAddr, client.logger)
	// 	// conn.Close()
	// 	cmd := cobra.Command{}
	// 	cmd.SetContext(context.Background())
	// 	runAdd(&cmd, []string{"one"})
	// })
}

func Test_parseLogin(t *testing.T) {
	t.Run("right", func(t *testing.T) {
		value := "one_two"
		delim := "_"
		res, err := parseLogin(value, delim)
		require.NoError(t, err)
		assert.IsType(t, &types.Login{}, res)
	})
	t.Run("wrong", func(t *testing.T) {
		value := "one two"
		delim := "_"
		res, err := parseLogin(value, delim)
		require.Error(t, err)
		assert.IsType(t, &types.Login{}, res)
	})
}

func Test_parseText(t *testing.T) {
	value := "one_two"
	res := parseText(value)
	assert.IsType(t, &types.Text{}, res)
}

func Test_parseBin(t *testing.T) {
	value := "one_two"
	res := parseBin(value)
	assert.IsType(t, &types.Binary{}, res)
}

func Test_parseCart(t *testing.T) {
	t.Run("right", func(t *testing.T) {
		value := "vllone_two_one two_123"
		delim := "_"
		res, err := parseCart(value, delim)
		require.NoError(t, err)
		assert.IsType(t, &types.Cart{}, res)
		assert.Equal(t, "one two", res.Holder)
	})
	t.Run("wrong", func(t *testing.T) {
		value := "one two"
		delim := "_"
		res, err := parseCart(value, delim)
		require.Error(t, err)
		assert.IsType(t, &types.Cart{}, res)
	})
}

func Test_parseValue(t *testing.T) {
	client := cliClient{}
	client.config.UserKey = "1234567890poiuyt"
	tests := []struct {
		name  string
		val   string
		rScrt rawSecret
	}{
		{"login", "one two", rawSecret{secretType: "login", name: "login", extra: "l f", delim: " "}},
		{"text", "one two", rawSecret{secretType: "text", name: "text", extra: "l & f"}},
		{"bin", "one two", rawSecret{secretType: "bin", name: "bin", extra: "l & f"}},
		{"cart", "vllone_two_one two_123,", rawSecret{secretType: "cart", name: "cart", extra: "l & f", delim: "_"}},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			res, err := parseValue(&client, tst.rScrt, tst.val)
			require.NoError(t, err)
			assert.NotEmpty(t, res.Data)
		})
	}
	t.Run("wrong type", func(t *testing.T) {
		wrongScrt := rawSecret{secretType: "new", name: "wrong"}
		res, err := parseValue(&client, wrongScrt, "")
		require.Error(t, err)
		assert.Equal(t, &pb.BinRequest{Key: "wrong"}, res)
	})
}
