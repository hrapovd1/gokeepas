package cli

import (
	"testing"

	"github.com/hrapovd1/gokeepas/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Test_runCp(t *testing.T) {
	client := cliClient{logger: zap.New(nil), config: config.Config{}}
	t.Run("empty args", func(t *testing.T) {
		runCp(&client, " ", &cobra.Command{}, []string{})
	})
}
