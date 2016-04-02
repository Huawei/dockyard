package main

import (
	"os"

	"github.com/codegangsta/cli"

	"github.com/containerops/dockyard/cmd"
	_ "github.com/containerops/dockyard/utils/setting"
)

func main() {
	app := cli.NewApp()

	app.Name = setting.AppName
	app.Usage = setting.Usage
	app.Version = setting.Version
	app.Author = setting.Author
	app.Email = setting.Email

	app.Commands = []cli.Command{
		cmd.CmdWeb,
		cmd.CmdDatabase,
		cmd.CmdOSS,
		cmd.CmdMonitor,
	}

	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
