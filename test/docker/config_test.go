package docker

import (
	"os/exec"
	"testing"

	"github.com/astaxie/beego/config"
)

var (
	Domains      string
	UserName     string
	DockerBinary string
)

func TestGetDockerConf(t *testing.T) {
	path := "../testsuite.conf"
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		t.Errorf("Read %s error: %v", path, err.Error())
	}

	if domains := conf.String("test::domains"); domains != "" {
		Domains = domains
	} else {
		t.Errorf("Read %s error: nil", domains)
	}

	if username := conf.String("test::username"); username != "" {
		UserName = username
	} else {
		t.Errorf("Read %s error: nil", username)
	}

	if client := conf.String("test::client"); client != "" {
		DockerBinary = client
	} else {
		t.Errorf("Read %s error: nil", client)
	}
}

func ParseCmdCtx(cmd *exec.Cmd) (output string, err error) {
	out, err := cmd.CombinedOutput()
	output = string(out)
	return output, err
}
