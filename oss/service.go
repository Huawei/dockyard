package oss

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/containerops/dockyard/oss/chunkmaster/api"
	"github.com/containerops/dockyard/oss/chunkmaster/metadata"
	// "github.com/containerops/wrench/setting"
)

type OSS struct {
	cm chunkmaster
	cs []metadata.Chunkserver
}

type chunkmaster struct {
	serverHost string
	serverPort int
	metaHost   string
	metaPort   string
	user       string
	passwd     string
	db         string
}

var OSSOBJ *OSS = new(OSS)

func (this *OSS) StartOSS() error {
	fmt.Println("enter initoss")
	var (
		err error
	)
	if err = this.Loadconfig(); err != nil {
		return err
	}
	if err = this.Initdb(); err != nil {
		return err
	}
	if err = this.Startmaster(); err != nil {
		return err
	}
	if err = this.Registerservers(); err != nil {
		return err
	}
	if err = this.Startservers(); err != nil {
		return err
	}
	return nil
}

func (this *OSS) Loadconfig() error {
	// load chunkmaster configs
	// TODO load configs from config files
	this.cm.metaHost = "10.229.40.121"
	this.cm.metaPort = "3306"
	this.cm.user = "root"
	this.cm.passwd = "wang"
	this.cm.db = "speedy1"
	// load chunkserver configs and conver to objs
	// TODO serverslist should come from config file
	servers := "1_10.229.40.121:7657;1_10.229.40.121:7658;1_10.229.40.121:7659"
	// servers string convert to Chunkservers
	for _, server := range strings.Split(servers, ";") {
		if isMatch, _ := regexp.MatchString("^\\d_((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\:\\d{0,5}$", server); !isMatch {
			return fmt.Errorf("chunkserver config format error : %s", server)
		}
		groupid := strings.Split(server, "_")[0]
		ip := strings.Split(strings.Split(server, "_")[1], ":")[0]
		port := strings.Split(strings.Split(server, "_")[1], ":")[1]
		chunkserver := metadata.Chunkserver{}
		chunkserver.Ip = ip
		groupiduint, _ := strconv.ParseUint(groupid, 10, 16)
		chunkserver.GroupId = uint16(groupiduint)
		portint, _ := strconv.Atoi(port)
		chunkserver.Port = portint
		this.cs = append(this.cs, chunkserver)
		fmt.Println(chunkserver)
	}
	return nil
}

func (this *OSS) Initdb() error {

	return nil
}

func (this *OSS) Startmaster() error {
	api.InitAll(this.cm.metaHost, this.cm.metaPort, this.cm.user, this.cm.passwd, this.cm.db)
	if err := api.LoadChunkserverInfo(); err != nil {
		return fmt.Errorf("loadChunkserverInfo error: %v", err)
	}
	go api.MonitorTicker(5, 30)
	return nil
}

func (this *OSS) Registerservers() error {
	// NOTE: change the name of func oss/chunkmaster/api/batchAddChunkserver to BatchAddChunkserver
	// TODO : check if the server ip  address and prot
	if err := api.BatchAddChunkserver(&this.cs); err != nil {
		return fmt.Errorf("Registerservers err %v", err)
	}
	return nil
}

func (this *OSS) Startservers() error {

	return nil
}
