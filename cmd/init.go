package cmd

import (
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
)

//initlization is
func initlization() {
	setting.LoadServerConfig()
	models.OpenDatabase()
}
