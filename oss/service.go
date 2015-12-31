package oss

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/containerops/dockyard/handler"
	"github.com/containerops/dockyard/oss/chunkmaster/api"
	"github.com/containerops/dockyard/oss/chunkmaster/metadata"
	"github.com/containerops/dockyard/oss/logs"
	"github.com/containerops/dockyard/oss/utils"
	"github.com/containerops/wrench/setting"
)

var _instance *oss

type oss struct {
	cm chunkmaster
	cs []metadata.Chunkserver
}

type chunkmaster struct {
	serverHost       string
	serverPort       int
	metaHost         string
	metaPort         string
	user             string
	passwd           string
	db               string
	limitCSNum       int
	connPoolCapacity int
}

func Instance() *oss {
	if _instance == nil {
		_instance = new(oss)
	}
	return _instance
}

func (this *oss) StartOSS() error {
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
		fmt.Println(err.Error())
	}
	if err = this.Registerservers(); err != nil {
		fmt.Println(err.Error())
	}
	if err = this.Startservers(); err != nil {
		fmt.Println(err.Error())
	}
	if err = handler.InitAPI(); err != nil {
		fmt.Println(err.Error())
	}
	return nil
}

func (this *oss) Loadconfig() error {
	// load chunkmaster configs
	this.cm.metaHost = setting.MetaHost
	this.cm.metaPort = setting.MetaPort
	this.cm.user = setting.DbUser
	this.cm.passwd = setting.DbPasswd
	this.cm.db = setting.Db
	this.cm.serverHost = setting.MasterHost
	this.cm.serverPort = setting.MasterPort
	this.cm.limitCSNum = setting.LimitCSNum
	this.cm.connPoolCapacity = setting.ConnPoolCapacity
	// Load chunkserver configs and convert chunkserver string to  to objs
	servers := setting.Servers
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
		chunkserver.DataDir = fmt.Sprintf("%v/server_%v_%v", setting.DataPath, chunkserver.Ip, chunkserver.Port)
		this.cs = append(this.cs, chunkserver)
	}
	return nil
}

func (this *oss) Initdb() error {
	return nil
}

func (this *oss) Startmaster() error {
	api.InitAll(this.cm.metaHost, this.cm.metaPort, this.cm.user, this.cm.passwd, this.cm.db)
	if err := api.LoadChunkserverInfo(); err != nil {
		return fmt.Errorf("loadChunkserverInfo error: %v", err)
	}
	go api.MonitorTicker(5, 30)

	router := initRouter()
	http.Handle("/cm/", router)
	log.Infof("listen %s:%d", this.cm.serverHost, this.cm.serverPort)
	go func() {
		if err := http.ListenAndServe(this.cm.serverHost+":"+strconv.Itoa(this.cm.serverPort), nil); err != nil {
			log.Fatalf("listen error: %v", err)
		}
	}()
	runtime.Gosched()
	return nil
}

func (this *oss) Registerservers() error {
	if err := api.BatchAddChunkserver(&this.cs); err != nil {
		return fmt.Errorf("Registerservers err %v", err)
	}
	return nil
}

func (this *oss) Startservers() error {
	binpath := "./oss/chunkserver/spy_server"
	errlogfolder := setting.ErrLogPath
	// check if chunkserver binary exsist,if not ,create it
	_, err := os.Stat(binpath)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("Cannot find chunkserver excution file")
	}
	// check if errlog folder exsist , if not ,create it
	_, err = os.Stat(errlogfolder)
	if err != nil || os.IsNotExist(err) {
		os.MkdirAll(errlogfolder, 0777)
	}
	for i := 0; i < len(this.cs); i++ {
		go func() {
			var stdout, stderr bytes.Buffer
			curcs := this.cs[i]
			_, err := os.Stat(curcs.DataDir)
			if err != nil || os.IsNotExist(err) {
				os.MkdirAll(curcs.DataDir, 0777)
			}
			port := fmt.Sprintf("%v", curcs.Port)
			masterport := fmt.Sprintf("%v", this.cm.serverPort)
			errlogpath := fmt.Sprintf("%v/errlog_%v_%v.log", errlogfolder, curcs.Ip, curcs.Port)
			groupid := fmt.Sprintf("%v", curcs.GroupId)
			cmd := exec.Command("./oss/chunkserver/spy_server", "--ip", curcs.Ip, "--port", port, "--master_ip", this.cm.serverHost, "--master_port", masterport, "--group_id", groupid, "--chunks", "2", "--data_dir", curcs.DataDir, "--error_log", errlogpath)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println(err.Error())
			}
		}()
		runtime.Gosched()
	}
	return nil
}

func initRouter() *mux.Router {
	router := mux.NewRouter()
	log.Debugf("initRouter")
	for method, routes := range api.RouteMap {
		for route, fct := range routes {
			localRoute := route
			localMethod := method
			log.Debugf("route: %s, method: %v", route, method)
			router.Path(localRoute).Methods(localMethod).HandlerFunc(fct)
		}
	}
	router.NotFoundHandler = http.HandlerFunc(util.NotFoundHandle)
	return router
}
