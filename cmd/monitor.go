package cmd

import (
	"github.com/codegangsta/cli"
)

var CmdMonitor = cli.Command{
	Name:        "monitor",
	Usage:       "monitor utils for all service",
	Description: "Monitor service health, database status, object storage service and so on.",
	Action:      runMonitor,
	Flags:       []cli.Flag{},
}

func runMonitor(c *cli.Context) {

}
