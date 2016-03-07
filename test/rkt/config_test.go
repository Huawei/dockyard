package rkt

import (
	"os/exec"
	"testing"

	"github.com/astaxie/beego/config"
)

var (
	Domains  string
	UserName string
	RktImage string
	PushCmd  string
	FetchCmd string
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

	if image := conf.String("test::image"); image != "" {
		RktImage = image
	} else {
		t.Errorf("Read %s error: nil", image)
	}

	if pushCmd := conf.String("test::pushcmd"); pushCmd != "" {
		PushCmd = pushCmd
	} else {
		PushCmd = "acpush"
	}

	if fetchCmd := conf.String("test::fetchcmd"); fetchCmd != "" {
		FetchCmd = fetchCmd
	} else {
		FetchCmd = "rkt"
	}
}

func ParseCmdCtx(cmd *exec.Cmd) (output string, err error) {
	out, err := cmd.CombinedOutput()
	output = string(out)
	return output, err
}
