package main

import (
	"os"
	"runtime"

	"github.com/codegangsta/cli"

	"github.com/containerops/dockyard/cmd"
	"github.com/containerops/dockyard/setting"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	app := cli.NewApp()

	app.Name = setting.AppName
	app.Usage = setting.Usage
	app.Version = setting.Version
	app.Author = setting.Author
	app.Email = setting.Email

	app.Commands = []cli.Command{
		cmd.CmdWeb,
		cmd.CmdBackend,
		cmd.CmdDatabase,
	}

	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
