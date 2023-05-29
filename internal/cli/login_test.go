package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/hrapovd1/gokeepas/internal/crypto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_readToken(t *testing.T) {
	token := "testToken"
	logger := zap.New(nil)
	t.Run("right", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("/tmp", "token")
		defer os.Remove(tmpFile.Name())
		require.NoError(t, err)
		_, err = tmpFile.WriteString(token)
		require.NoError(t, err)
		result, err := readToken(tmpFile.Name(), logger)
		require.NoError(t, err)
		require.Equal(t, token, result)
	})
	t.Run("empty token", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("/tmp", "token")
		defer os.Remove(tmpFile.Name())
		require.NoError(t, err)
		result, err := readToken(tmpFile.Name(), logger)
		require.Error(t, err)
		require.Empty(t, result)
	})
}

func Test_writeToken(t *testing.T) {
	token := "testToken"
	logger := zap.New(nil)
	// generate random token file name
	rStr, err := crypto.GenServerKey(7)
	require.NoError(t, err)
	ftName := "/tmp/" + rStr + ".token"
	t.Run("right", func(t *testing.T) {
		err := writeToken(token, ftName, logger)
		require.NoError(t, err)
		tokenFile, err := os.Open(ftName)
		defer os.Remove(ftName)
		require.NoError(t, err)
		result := make([]byte, 20)
		_, err = fmt.Fscan(tokenFile, &result)
		require.NoError(t, err)
		assert.Equal(t, token, string(result))
	})
	t.Run("wrong path", func(t *testing.T) {
		err := writeToken(token, "/", logger)
		require.Error(t, err)
	})
}

func Test_runLogin(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		options := loginOptions{}
		client := cliClient{logger: zap.New(nil)}
		cmd := cobra.Command{}
		runLogin(&client, options, &cmd)
	})
}
