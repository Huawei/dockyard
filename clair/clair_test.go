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
	if err := InitClair(); err != nil {
		t.Log("Clair service init failed!")
		return
	} else {
		fmt.Println("Success in init clair service")
	}
	var h History
	h.ID = "123"
	h.Parent = ""
	// Assume we have this layer file in the current directoy
	h.localPath = "123.tar"

	if err := PutLayer(h); err != nil {
		fmt.Println("Cannot put layer")
		return
	}
	if vulns, err := GetVulns(h.ID); err != nil {
		for index := 0; index < len(vulns); index++ {
			fmt.Println(*vulns[index])
		}
	} else {
		fmt.Println("No vul risk!")
	}
	StopClair()
}
