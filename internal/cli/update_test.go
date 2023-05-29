package cli

import (
	"testing"

	"github.com/hrapovd1/gokeepas/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Test_runUpd(t *testing.T) {
	client := cliClient{
		logger: zap.New(nil),
		config: config.Config{},
	}
	secret := rawSecret{delim: `,`}
	cmd := cobra.Command{}
	t.Run("empty args", func(t *testing.T) {
		runUpd(&client, secret, &cmd, []string{})
	})
}
