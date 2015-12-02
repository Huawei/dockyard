package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/test"
)

func TestPullInit(t *testing.T) {
	var cmd *exec.Cmd
	var err error
	var out string

	reponame := "busybox"
	repotags := []string{"latest", "1.0", "2.0"}
	repoBase := reponame + ":" + repotags[0]

	if err = exec.Command(test.DockerBinary, "inspect", repoBase).Run(); err != nil {
		cmd = exec.Command(test.DockerBinary, "pull", repoBase)
		if out, err = test.ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Pull testing preparation is failed: [Info]%v, [Error]%v", out, err)
		}
	}

	repoDest := test.Domains + "/" + test.UserName + "/" + reponame
	for _, v := range repotags {
		tag := repoDest + ":" + v
		cmd = exec.Command(test.DockerBinary, "tag", "-f", repoBase, tag)
		if out, err = test.ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Tag %v failed: [Info]%v, [Error]%v", repoBase, out, err)
		}
		cmd = exec.Command(test.DockerBinary, "push", tag)
		if out, err = test.ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Push all tags %v failed: [Info]%v, [Error]%v", tag, out, err)
		}
	}
}

func getCurrentVersion() (string, string, error) {
	cmd := exec.Command(test.DockerBinary, "-v")
	out, err := test.ParseCmdCtx(cmd)
	if err != nil {
		return "", "", fmt.Errorf("Get docker version failed")
	}

	curVer := strings.Split(strings.Split(out, ",")[0], " ")[2]
	val := test.Compare(curVer, "1.6.0")
	if val < 0 {
		return curVer, "V1", nil
	} else {
		return curVer, "V2", nil
	}
}

func chkAndDelAllRepoTags(repoBase string, repotags []string) error {
	for _, v := range repotags {
		repotag := repoBase + ":" + v
		if err := exec.Command(test.DockerBinary, "inspect", repotag).Run(); err != nil {
			return fmt.Errorf(err.Error())
		}

		cmd := exec.Command(test.DockerBinary, "rmi", repotag)
		if _, err := test.ParseCmdCtx(cmd); err != nil {
			return fmt.Errorf(err.Error())
		}
	}
	return nil
}

func TestPullRepoSingleTag(t *testing.T) {
	var cmd *exec.Cmd
	var err error
	var out string

	reponame := "busybox"
	repotags := []string{"latest"}
	repoDest := test.Domains + "/" + test.UserName + "/" + reponame

	for i := 1; i <= 2; i++ {
		cmd = exec.Command(test.DockerBinary, "pull", repoDest+":"+repotags[0])
		if out, err = test.ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Pull %v failed: [Info]%v, [Error]%v", repoDest+":"+repotags[0], out, err)
		}
	}

	if err := chkAndDelAllRepoTags(repoDest, repotags); err != nil {
		t.Fatalf("Pull %v failed, it is not found in location. [Error]%v", repoDest, err.Error())
	}
}

//if there are multiple tags in registry,if protocol V2 pulling is not specified tag, default to get latest tag,different from protocol V1
func TestPullV2RepoWithoutTags(t *testing.T) {
	var cmd *exec.Cmd
	var err error
	var out string

	reponame := "busybox"
	repoDest := test.Domains + "/" + test.UserName + "/" + reponame
	repotags := []string{"latest", "1.0", "2.0"}

	_, registry, err := getCurrentVersion()
	if err != nil {
		t.Fatal(err.Error())
	}

	if registry == "V2" {
		for _, v := range repotags {
			tag := repoDest + ":" + v
			if err = exec.Command(test.DockerBinary, "inspect", tag).Run(); err == nil {
				cmd = exec.Command(test.DockerBinary, "rmi", tag)
				if out, err = test.ParseCmdCtx(cmd); err != nil {
					t.Fatalf("Pull testing preparation is failed: [Info]%v, [Error]%v", out, err)
				}
			}
		}

		for i := 1; i <= 2; i++ {
			cmd = exec.Command(test.DockerBinary, "pull", repoDest)
			if out, err = test.ParseCmdCtx(cmd); err != nil {
				t.Fatalf("Pull %v failed: [Info]%v, [Error]%v", repoDest, out, err)
			}
		}

		for _, v := range repotags {
			if v != "latest" {
				tag := repoDest + ":" + v
				if err = exec.Command(test.DockerBinary, "inspect", tag).Run(); err == nil {
					t.Fatalf("Not expect result")
				}
			}
		}

		if err := chkAndDelAllRepoTags(repoDest, []string{"latest"}); err != nil {
			t.Fatalf("Pull %v failed, it is not found in location. [Error]%v", repoDest, err.Error())
		}
	}
}

func TestPullRepoMultipleTags(t *testing.T) {
	var cmd *exec.Cmd
	var err error
	var out string

	reponame := "busybox"
	repotags := []string{"latest", "1.0", "2.0"}
	repoDest := test.Domains + "/" + test.UserName + "/" + reponame

	curVer, registry, err := getCurrentVersion()
	if err != nil {
		t.Fatal(err.Error())
	}

	if !strings.Contains(curVer, "1.6") {
		//Pull Repository Multiple Tags with option "-a",protocol V1 and V2 are all the same except Docker 1.6.x
		for i := 1; i <= 2; i++ {
			cmd = exec.Command(test.DockerBinary, "pull", "-a", repoDest)
			if out, err = test.ParseCmdCtx(cmd); err != nil {
				t.Fatalf("Pull all tags of %v failed: [Info]%v, [Error]%v", repoDest, out, err)
			}
		}
		if err := chkAndDelAllRepoTags(repoDest, repotags); err != nil {
			t.Fatalf("Pull all tags of %v failed, it is not found in location. [Error]%v", repoDest, err.Error())
		}

		if registry == "V1" {
			//Docker daemon support to pull Repository Multiple Tags without option "-a"
			for i := 1; i <= 2; i++ {
				cmd = exec.Command(test.DockerBinary, "pull", repoDest)
				if out, err = test.ParseCmdCtx(cmd); err != nil {
					t.Fatalf("Pull all tags of %v failed: [Info]%v, [Error]%v", repoDest, out, err)
				}
			}
			if err := chkAndDelAllRepoTags(repoDest, repotags); err != nil {
				t.Fatalf("Pull all tags of %v failed, it is not found in location. [Error]%v", repoDest, err.Error())
			}

			//if there are multiple tags in registry,pull specified tag will get all tags of repository
			for i := 1; i <= 2; i++ {
				cmd = exec.Command(test.DockerBinary, "pull", repoDest+":"+repotags[0])
				if out, err = test.ParseCmdCtx(cmd); err != nil {
					t.Fatalf("Pull all tags of %v failed: [Info]%v, [Error]%v", repoDest, out, err)
				}
			}
			if err := chkAndDelAllRepoTags(repoDest, repotags); err != nil {
				t.Fatalf("Pull all tags of %v failed, it is not found in location. [Error]%v", repoDest, err.Error())
			}
		}
	} else {
		//There is a bug about "pull -a" in Docker 1.6.x,it will be change from protocol V2 to V1 if using "pull -a"
	}
}

func TestPullNonExistentRepo(t *testing.T) {
	repoDest := test.Domains + "/" + test.UserName + "/" + "nonexistentrepo"
	cmd := exec.Command(test.DockerBinary, "pull", repoDest)
	if out, err := test.ParseCmdCtx(cmd); err == nil {
		t.Fatalf("Pull %v failed: [Info]%v, [Error]%v", repoDest, out, err)
	}
}

/*
func Example_random() {
	fmt.Println("mabintest")
	//output:
	//mabintest
}
*/
