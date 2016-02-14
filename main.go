package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"

	"github.com/containerops/dockyard/cmd"
	_ "github.com/containerops/dockyard/middleware/notifications"
	"github.com/containerops/wrench/setting"
)

func main() {
	if err := setting.SetConfig("conf/containerops.conf"); err != nil {
		fmt.Printf("Read config failed: %v\n", err.Error())
		return
	}

	//if read middleware config failed, register function of middleware will be skipped
	setting.GetConfFromJSON("conf/config.json")

	app := cli.NewApp()

	app.Name = setting.AppName
	app.Usage = setting.Usage
	app.Version = setting.Version
	app.Author = setting.Author
	app.Email = setting.Email

	app.Commands = []cli.Command{
		cmd.CmdWeb,
	}

	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
