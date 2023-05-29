package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Test_runSignup(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		client := cliClient{logger: zap.New(nil)}
		options := loginOptions{}
		runSignup(&client, options, &cobra.Command{})
	})
}
