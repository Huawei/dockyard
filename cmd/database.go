package cmd

import (
	"github.com/codegangsta/cli"
)

var CmdDatabase = cli.Command{
	Name:        "database",
	Usage:       "database utils for backend database",
	Description: "Dockyard run base SQL database like MySQL, database command provide some utils of migrate, backup, config and so on.",
	Action:      runDatabase,
	Flags:       []cli.Flag{},
}

func runDatabase(c *cli.Context) {

}
