package rkt

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPullInit(t *testing.T) {
	if err := checkTestingEnv(); err != nil {
		t.Fatalf("Failed to prepare the testing enviornment: %v", err)
	}

	if err := pushImage(); err != nil {
		//	t.Fatalf("Failed to push the testing image to Dockyard: %v", err)
	}

	if err := fetchImage(); err != nil {
		t.Fatalf("Failed to fetch the testing image fro Dockyard: %v", err)
	}
}

func pushImage() error {
	imageName := strings.TrimSuffix(filepath.Base(RktImage), ".aci")
	cmd := exec.Command(PushCmd, RktImage, RktImage+".asc", Domains+"/"+UserName+"/"+imageName)
	if out, err := ParseCmdCtx(cmd); err != nil {
		fmt.Println(out)
		return err
	}
	return nil
}

func fetchImage() error {
	tempDir, err := ioutil.TempDir("", "rkt-test-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	imageName := strings.TrimSuffix(filepath.Base(RktImage), ".aci")
	cmd := exec.Command(FetchCmd, "fetch", Domains+"/"+UserName+"/"+imageName, "--dir="+tempDir)
	fmt.Println(cmd)
	if out, err := ParseCmdCtx(cmd); err != nil {
		fmt.Println(out)
		return err
	}
	return nil
}

func checkTestingEnv() error {
	if _, err := exec.LookPath(FetchCmd); err != nil {
		return err
	}

	if _, err := exec.LookPath(PushCmd); err != nil {
		return err
	}

	if _, err := os.Stat(RktImage); err != nil {
		return err
	}

	if _, err := os.Stat(RktImage + ".asc"); err != nil {
		return err
	}

	return nil
}
