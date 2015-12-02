package test

import (
	"os/exec"
)

func ParseCmdCtx(cmd *exec.Cmd) (output string, err error) {
	out, err := cmd.CombinedOutput()
	output = string(out)
	return output, err
}

func Compare(a, b string) int {
	if a == b {
		return 0
	}
	if a < b {
		return -1
	}
	return +1
}
