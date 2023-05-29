package main

import (
	"github.com/hrapovd1/gokeepas/internal/cli"
)

func main() {
	cli.Execute(cli.NewRootCmd())
}
