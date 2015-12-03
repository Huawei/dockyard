package upyun

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
)

var (
	u              *UpYun
	invalidAccount *UpYun
	root           string
	txtpath        string
	imgpath        string
)

const (
	TXTPATH = "../tests/test.txt"
	IMGPATH = "../tests/test.png"
)

func init() {
	BUCKET := os.Getenv("UPYUN_BUCKET")
	USERNAME := os.Getenv("UPYUN_USERNAME")
	PASSWD := os.Getenv("UPYUN_PASSWORD")

	if BUCKET == "" || USERNAME == "" || PASSWD == "" {
		panic("Incomplete file bucket infomation in environment variable")
	}

	u = NewUpYun(BUCKET, USERNAME, PASSWD)
	invalidAccount = NewUpYun("bucket", "username", "passwd")
	root = "GoSDKTest"

	err := u.Mkdir(root)
	if err != nil {
		panic(err)
	}

	// Upload test.txt to root dir
	txtfi, err := os.Open(TXTPATH)
	if err != nil {
		panic(err)
	}

	txtpath = path.Join(root, "test.txt")
	_, err = u.Put(txtpath, txtfi, false, "")
	if err != nil {
		panic(err)
	}

	txtfi.Close()

	// Upload text.png to root dir
	imgfi, err := os.Open(IMGPATH)
	if err != nil {
		panic(err)
	}

	imgpath = path.Join(root, "test.png")
	_, err = u.Put(imgpath, imgfi, false, "")
	if err != nil {
		panic(err)
	}

	imgfi.Close()

}

func assert(condition bool, log string, t *testing.T) {
	if !condition {
		t.Error(log)
	}
}

func TestUpyun(t *testing.T) {
	// Use it the right way

	testUsage(t, u)
	testGetList(t, u)
	testGetInfo(t, u)
	testMkdir(t, u)
	testGetFile(t, u)
	testPutFile(t, u)
	testDelete(t, u)
	// Use it the wrong way to make it fail

	testAuthFail(t, invalidAccount)
}

// -----------------------------------------------------------------------------------

func testUsage(t *testing.T, client *UpYun) {
	used, err := client.Usage()
	assert(err == nil, "Usage: Get usage error", t)
	assert(used > 0, "Usage: smaller than zero", t)

	fmt.Println(used)
}

func testGetList(t *testing.T, client *UpYun) {
	infoList, err := client.GetList(root)
	assert(err == nil, "Get list error", t)

	for _, info := range infoList {
		if info.Name == "test.txt" {
			assert(info.Type == "N", "GetList: wrong file type", t)
			assert(info.Size == 10, "GetList: wrong file size", t)
		}
	}
}

func testGetInfo(t *testing.T, client *UpYun) {
	fileInfo, err := client.GetInfo(txtpath)

	assert(err == nil, "GetInfo error", t)
	assert(fileInfo.Type == "file", "GetInfo: wrong type", t)
	assert(fileInfo.Size == 10, "GetInfo: wrong size", t)
}

func testMkdir(t *testing.T, client *UpYun) {
	fileInfo, err := client.GetInfo(root)
	assert(err == nil, "Mkdir: dir not exist", t)
	assert(fileInfo.Type == "folder", "Mkdir: wrong dir type", t)
}

func testGetFile(t *testing.T, client *UpYun) {
	txtfo, err := os.Create("get.txt")
	if err != nil {
		panic(err)
	}

	imgfo, err := os.Create("get.png")
	if err != nil {
		panic(err)
	}

	defer func() {
		err := txtfo.Close()
		if err != nil {
			panic(err)
		}
		os.Remove("get.txt")

		err = imgfo.Close()
		if err != nil {
			panic(err)
		}
		os.Remove("get.png")
	}()

	err = client.Get(txtpath, txtfo)
	assert(err == nil, "Get: get txt error", t)

	err = client.Get(imgpath, imgfo)
	assert(err == nil, "Get: get img error", t)

	stat, _ := txtfo.Stat()
	assert(stat.Size() == 10, "Get: get txt size error", t)

	stat, _ = imgfo.Stat()
	assert(stat.Size() == 13001, "Get: get img size error", t)
}

func testPutFile(t *testing.T, client *UpYun) {
	txtinfo, err := client.GetInfo(txtpath)
	assert(err == nil, "Put: put txt error", t)
	assert(txtinfo.Size == 10, "Put: put txt size error", t)
	assert(txtinfo.Type == "file", "Put: put txt type error", t)

	imgfi, err := os.Open(IMGPATH)
	if err != nil {
		panic(err)
	}
	_, err = client.Put(imgpath, imgfi, true, "")
	imgfi.Close()

	imginfo, err := client.GetInfo(imgpath)
	assert(err == nil, "Put: put img error", t)
	assert(imginfo.Size == 13001, "Put: put img size error", t)
	assert(imginfo.Type == "file", "Put: put img type error", t)
}

func testDelete(t *testing.T, client *UpYun) {
	err := client.Delete(txtpath)
	assert(err == nil, "Delete: delete txt error", t)

	err = client.Delete(imgpath)
	assert(err == nil, "Delete: delete img error", t)

	err = client.Delete(root)
	assert(err == nil, "Delete: delete folder error", t)
}

// -----------------------------------------------------------------------------------

func testAuthFail(t *testing.T, client *UpYun) {
	if _, err := client.Usage(); err != nil {
		assert(strings.Contains(err.Error(), "401 Unauthorized"), "testAuthFail error", t)
	}
}
