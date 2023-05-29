/*
Package cli contents methods and types for KeepPas cli client.
*/
package cli

import (
	"github.com/spf13/cobra"
)

func newKVCmd() *cobra.Command {
	// kvCmd represents the kv command
	return &cobra.Command{
		Use:   "kv",
		Short: "Manage secrets",
	}
}
