package clair

import (
	"fmt"
	"testing"
)

func DemoConfig() ClairConfig {
	var conf ClairConfig
	conf.KeepDB = false
	conf.DBPath = "/tmp/clair-db-test"
	conf.LogLevel = DefaultClairLogLevel
	conf.Duration = DefaultClairUpdateDuration
	conf.VulnPriority = DefaultClairVulnPriority
	return conf

}

func Test_ClairService(t *testing.T) {
	if err := ClairServiceInit(); err != nil {
		t.Log("Clair service init failed!")
		return
	} else {
		fmt.Println("Success in init clair service")
	}
	id := "123"
	parentID := ""
	// Assume we have this layer file in the current directoy
	Path := "123.tar"

	if vulns, err := ClairGetVulns(id, parentID, Path); err != nil {
		for index := 0; index < len(vulns); index++ {
			fmt.Println(*vulns[index])
		}
	} else {
		fmt.Println("No vul risk!")
	}
	ClairServiceStop()
}
