package test

import (
	"fmt"
	"os/exec"
	"testing"

	//"github.com/astaxie/beego/config"

	//"github.com/containerops/dockyard/test/docker"
	//"github.com/containerops/wrench/utils"
)

type Handler func(t *testing.T)

var testsuite map[string]Handler = map[string]Handler{}

func TestEntry(t *testing.T) {
	var namearray []string = []string{"docker", "rkt"}

	for _, name := range namearray {
		if hanlde, existed := testsuite[name]; existed {
			hanlde(t)
		}
	}
}

func Register(name string, handler Handler) error {
	if _, existed := testsuite[name]; existed {
		return fmt.Errorf("%v has already been registered", name)
	}
	testsuite[name] = handler

	return nil
}

func ParseCmdCtx(cmd *exec.Cmd) (output string, err error) {
	out, err := cmd.CombinedOutput()
	output = string(out)
	return output, err
}
