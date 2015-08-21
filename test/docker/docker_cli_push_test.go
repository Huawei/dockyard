package main

import (
	"os/exec"
	"testing"

	"github.com/containerops/dockyard/test"
)

func TestPushInit(t *testing.T) {
	repoBase := "busybox:latest"

	if err := exec.Command(test.DockerBinary, "inspect", repoBase).Run(); err != nil {
		cmd := exec.Command(test.DockerBinary, "pull", repoBase)
		if out, err := test.ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Push testing preparation is failed: [Info]%v, [Error]%v", out, err)
		}
	}
}

func TestPushRepoWithSingleTag(t *testing.T) {
	var cmd *exec.Cmd
	var err error
	var out string

	reponame := "busybox"
	repotag := "latest"
	repoBase := reponame + ":" + repotag

	repoDest := test.Domains + "/" + test.UserName + "/" + repoBase
	cmd = exec.Command(test.DockerBinary, "tag", "-f", repoBase, repoDest)
	if out, err = test.ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Tag %v failed: [Info]%v, [Error]%v", repoBase, out, err)
	}

	//push the same repository with specified tag more than once to cover related code processing branch
	for i := 1; i <= 2; i++ {
		cmd = exec.Command(test.DockerBinary, "push", repoDest)
		if out, err = test.ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Push %v failed: [Info]%v, [Error]%v", repoDest, out, err)
		}
	}

	cmd = exec.Command(test.DockerBinary, "rmi", repoDest)
	out, err = test.ParseCmdCtx(cmd)
}

func TestPushRepoWithMultipleTags(t *testing.T) {
	var cmd *exec.Cmd
	var err error
	var out string

	reponame := "busybox"
	repotags := []string{"latest", "1.0", "2.0"}
	repoBase := reponame + ":" + repotags[0] //pull busybox:latest from docker hub

	repoDest := test.Domains + "/" + test.UserName + "/" + reponame
	for _, v := range repotags {
		tag := repoDest + ":" + v
		cmd = exec.Command(test.DockerBinary, "tag", "-f", repoBase, tag)
		if out, err = test.ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Tag %v failed: [Info]%v, [Error]%v", repoBase, out, err)
		}
	}

	//push the same repository with multiple tags more than once to cover related code processing branch
	for i := 1; i <= 2; i++ {
		cmd = exec.Command(test.DockerBinary, "push", repoDest)
		if out, err = test.ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Push all tags %v failed: [Info]%v, [Error]%v", repoDest, out, err)
		}
	}

	for _, v := range repotags {
		tag := repoDest + ":" + v
		cmd = exec.Command(test.DockerBinary, "rmi", tag)
		out, err = test.ParseCmdCtx(cmd)
	}
}

/*
func Example_random() {
	fmt.Println("mabintest")
	//output:
	//mabintest
}
*/
