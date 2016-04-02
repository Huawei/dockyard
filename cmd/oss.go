package cmd

import (
	"github.com/codegangsta/cli"
)

var CmdOSS = cli.Command{
	Name:        "oss",
	Usage:       "start object storage service",
	Description: "Provide a build-in object storage service.",
	Action:      runOSS,
	Flags:       []cli.Flag{},
}

func runOSS(c *cli.Context) {

}
