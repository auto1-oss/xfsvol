package main

import (
	"os"

	"github.com/cirocosta/xfsvol/xfsvolctl/commands"
	"gopkg.in/urfave/cli.v1"
)

var (
	version string = "v1.0.2"
)

func main() {
	app := cli.NewApp()
	app.Name = "xfsvolctl"
	app.Version = version
	app.Usage = "Controls the 'xfsvol' volume plugin"
	app.Commands = []cli.Command{
		commands.Ls,
		commands.Create,
		commands.Delete,
	}
	app.Run(os.Args)
}
