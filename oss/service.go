package oss

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/mux"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/oss/apiserver"
	"github.com/containerops/dockyard/oss/chunkmaster/api"
	"github.com/containerops/dockyard/oss/chunkmaster/metadata"
	log "github.com/containerops/dockyard/oss/logs"
	"github.com/containerops/dockyard/oss/utils"
	"github.com/containerops/dockyard/setting"
)

var (
	_instance *oss
)

type oss struct {
	ServerMode string
	Nodenum    int
	cm         chunkmaster
	nodes      []node
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

type node struct {
	groupid    int
	ip         string
	port       int
	listenmode string
	datadir    string
	errlogdir  string
	chunknum   int
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
	if err = this.LoadChunkMasterConfig(); err != nil {
		fmt.Println(err.Error())
	}
	if err = this.LoadChunkServerConfig(); err != nil {
		fmt.Println(err.Error())
	}
	if err = this.Initdb(); err != nil {
		fmt.Println(err.Error())
	}
	if err = this.Startmaster(); err != nil {
		fmt.Println(err.Error())
		return err
	}
	if err = this.Registerservers(); err != nil {
		fmt.Println(err.Error())
	}
	switch this.ServerMode {
	case "allinone":
		if err = this.StartServersAllinone(); err != nil {
			fmt.Println(err.Error())
			return err
		}
	case "distribute":
		go func() {
			if err = this.StartServerDistribute(); err != nil {
				fmt.Println(err.Error())
			}
		}()
	}
	if err = apiserver.InitAPI(); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (this *oss) LoadChunkMasterConfig() error {
	var err error
	confpath := "oss/chunkmaster.conf"
	conf, err := config.NewConfig("ini", confpath)
	if err != nil {
		return fmt.Errorf("Read OSS config file %s error: %v", confpath, err.Error())
	}
	// load chunkmaster configs
	if servermode := conf.String("servermode"); servermode != "" {
		this.ServerMode = servermode
	} else {
		this.ServerMode = "allinone"
	}
	if masterhost := conf.String("masterhost"); masterhost != "" {
		this.cm.serverHost = masterhost
		apiserver.MasterUrl = masterhost
	} else {
		err = fmt.Errorf("masterhost value is null")
	}
	this.cm.serverPort, err = conf.Int("masterport")
	apiserver.MasterPort = strconv.Itoa(this.cm.serverPort)

	if metahost := conf.String("metahost"); metahost != "" {
		this.cm.metaHost = metahost
		apiserver.MetadbIp = metahost
	} else {
		err = fmt.Errorf("metaHost  value is null")
	}
	if metaport := conf.String("metaport"); metaport != "" {
		this.cm.metaPort = metaport
		apiserver.MetadbPort, err = strconv.Atoi(metaport)
	} else {
		err = fmt.Errorf("metaport  value is null")
	}

	if dbuser := conf.String("dbuser"); dbuser != "" {
		this.cm.user = dbuser
		apiserver.MetadbUser = dbuser
	} else {
		err = fmt.Errorf("dbuser value is null")
	}
	if dbpasswd := conf.String("dbpasswd"); dbpasswd != "" {
		this.cm.passwd = dbpasswd
		apiserver.MetadbPassword = dbpasswd
	} else {
		err = fmt.Errorf("dbpasswd value is null")
	}
	if db := conf.String("db"); db != "" {
		this.cm.db = db
	} else {
		err = fmt.Errorf("db value is null")
	}
	this.cm.limitCSNum, err = conf.Int("limitcsnum")
	apiserver.LimitNum = this.cm.limitCSNum
	this.cm.connPoolCapacity, err = conf.Int("connpoolcapacity")
	apiserver.ConnPoolCapacity = this.cm.connPoolCapacity

	return nil
}

func (this *oss) LoadChunkServerConfig() error {
	var err error
	confpath := "oss/chunkserver.conf"
	conf, err := config.NewConfig("ini", confpath)
	this.Nodenum, _ = conf.Int("nodenum")
	for i := 0; i < this.Nodenum; i++ {
		nodename := fmt.Sprintf("node%v", i+1)
		nodetmp := new(node)
		nodetmp.groupid, _ = conf.Int(nodename + "::" + "groupid")
		if ip := conf.String(nodename + "::" + "ip"); ip != "" {
			nodetmp.ip = ip
		} else {
			err = fmt.Errorf(nodename + " ip value is null")
		}
		nodetmp.port, _ = conf.Int(nodename + "::" + "port")
		if listenmode := conf.String(nodename + "::" + "listenmode"); listenmode != "" {
			nodetmp.listenmode = listenmode
		} else {
			nodetmp.listenmode = "https"
		}
		if datadir := conf.String(nodename + "::" + "datadir"); datadir != "" {
			nodetmp.datadir = datadir
		} else {
			err = fmt.Errorf(nodename + " datadir value is null")
		}
		if errlogdir := conf.String(nodename + "::" + "errlogdir"); errlogdir != "" {
			nodetmp.errlogdir = errlogdir
		} else {
			err = fmt.Errorf(nodename + " errlogdir value is null")
		}
		nodetmp.chunknum, _ = conf.Int(nodename + "::" + "chunknum")

		fmt.Println(nodetmp)
		this.nodes = append(this.nodes, *nodetmp)
	}
	return err
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
	for _, nodetmp := range this.nodes {
		regserver := node2metaserver(nodetmp)
		exist, err := api.IsChunkServerExsist(&regserver)
		if err != nil {
			fmt.Errorf("check ChunkServer is Exist error", err)
		}
		if exist {
			fmt.Printf("[OSS]chunkserver [%v:%v] is already exsist, will NOT register\n", regserver.Ip, regserver.Port)
		} else {
			if err := api.AddChunkserver(&regserver); err != nil {
				return fmt.Errorf("[OSS]chunkserver [%v:%v] register failed, error info:%v \n", regserver.Ip, regserver.Port, err)
			}
		}
	}
	return nil
}

func (this *oss) StartServersAllinone() error {
	binpath := "./oss/chunkserver/spy_server"
	errlogfolder := "./errlog"
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
	for _, node2start := range this.nodes {
		go func() {
			var stdout, stderr bytes.Buffer
			// check if data folder exsist , if not ,create it
			_, err := os.Stat(node2start.datadir)
			if err != nil || os.IsNotExist(err) {
				os.MkdirAll(node2start.datadir, 0777)
			}
			port := fmt.Sprintf("%v", node2start.port)
			masterport := fmt.Sprintf("%v", this.cm.serverPort)
			errlogpath := fmt.Sprintf("%v/errlog_%v_%v.log", errlogfolder, node2start.ip, node2start.port)
			groupid := fmt.Sprintf("%v", node2start.groupid)
			cmd := exec.Command("./oss/chunkserver/spy_server", "--ip", node2start.ip, "--port", port, "--master_ip", this.cm.serverHost, "--master_port", masterport, "--group_id", groupid, "--chunks", "2", "--data_dir", node2start.datadir, "--error_log", errlogpath)
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

func (this *oss) StartServerDistribute() error {
	for _, node2start := range this.nodes {
		header := make(map[string][]string, 0)
		header["ip"] = []string{node2start.ip}
		portstr := strconv.Itoa(node2start.port)
		header["port"] = []string{portstr}
		chunknumstr := strconv.Itoa(node2start.chunknum)
		header["chunknum"] = []string{chunknumstr}
		header["datadir"] = []string{node2start.datadir}
		header["errlogdir"] = []string{node2start.errlogdir}
		header["masterip"] = []string{this.cm.serverHost}
		groupidstr := strconv.Itoa(node2start.groupid)
		header["groupid"] = []string{groupidstr}
		masterportstr := strconv.Itoa(this.cm.serverPort)
		header["masterport"] = []string{masterportstr}
		header["masterlistenmode"] = []string{setting.ListenMode}
		switch node2start.listenmode {
		case "http":
			result, statusCode, err := util.Call("POST", "http://"+node2start.ip+":80", "/oss/chunkserver", nil, header)
			if err != nil {
				return fmt.Errorf("[OSS] Sent remote server start request error: %v", err)
			}
			if statusCode != http.StatusOK {
				return fmt.Errorf("[OSS] Start remote server failed STATUS CODE:%v ,result:%s", statusCode, result)
			}
		default:
			result, statusCode, err := util.CallHttps("POST", "https://"+node2start.ip+":443", "/oss/chunkserver", nil, header)
			if err != nil {
				return fmt.Errorf("[OSS] Sent remote server start request error: %v", err)
			}
			if statusCode != http.StatusOK {
				return fmt.Errorf("[OSS] Start remote server failed STATUS CODE:%v ,result:%s", statusCode, result)
			}
		}

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

func node2metaserver(nodecon node) metadata.Chunkserver {
	metaserver := new(metadata.Chunkserver)
	metaserver.Ip = nodecon.ip
	metaserver.Port = nodecon.port
	metaserver.DataDir = nodecon.datadir
	metaserver.GroupId = uint16(nodecon.groupid)
	return *metaserver
}

func StartLocalServer(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	var stdout, stderr bytes.Buffer
	header := ctx.Req.Header
	ip := header.Get("ip")
	port := header.Get("port")
	masterip := header.Get("masterip")
	masterport := header.Get("masterport")
	chunknum := header.Get("chunknum")
	groupid := header.Get("groupid")
	datadir := header.Get("datadir")
	errlogdir := header.Get("errlogdir")
	masterlistenmode := header.Get("masterlistenmode")
	// check if chunkserver binary exsist,if not ,create it
	_, err := os.Stat("./oss/chunkserver/spy_server")
	if err != nil && os.IsNotExist(err) {
		return http.StatusInternalServerError, []byte("Cannot find chunkserver excution file")
	}
	// check if errlog folder exsist , if not ,create it
	_, err = os.Stat(errlogdir)
	if err != nil || os.IsNotExist(err) {
		os.MkdirAll(errlogdir, 0777)
	}
	// check if data folder exsist , if not ,create it
	_, err = os.Stat(datadir)
	if err != nil || os.IsNotExist(err) {
		os.MkdirAll(datadir, 0777)
	}
	// excecute chunkserver start script in a new goroutine, if failed, send https request to notify master node
	go func() {
		cmd := exec.Command("./oss/chunkserver/install.sh", "-i", ip, "-p", port, "-m", masterip, "-n", masterport, "-c", chunknum, "-g", groupid, "-d", datadir, "-e", errlogdir)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			log.Alert("[OSS]Start local ChunkServer error, error INFO: %v", err)
			log.Alert("[OSS]Start local ChunkServer error, STDOUT: %s", stdout.Bytes())
			log.Alert("[OSS]Start local ChunkServer error, STDERR: %s", stderr.Bytes())
			header2send := make(map[string][]string, 0)
			header2send["nodeip"] = []string{ip}
			header2send["nodeport"] = []string{port}
			header2send["groupid"] = []string{groupid}
			header2send["stderr"] = []string{string(stderr.Bytes())}
			header2send["stdout"] = []string{string(stdout.Bytes())}
			switch masterlistenmode {
			case "http":
				result, statusCode, err := util.Call("PUT", "http://"+masterip+":80", "/oss/chunkserver/info", nil, header2send)
				if err != nil {
					log.Error("[OSS] Sent chunkserver info back to chunkmaster error: %v", err)
				}
				if statusCode != http.StatusOK {
					log.Error("[OSS] Sent chunkserver info back to chunkmaster failed STATUS CODE:%v ,result:%s", statusCode, result)
				}
			default:
				result, statusCode, err := util.CallHttps("PUT", "https://"+masterip+":443", "/oss/chunkserver/info", nil, header2send)
				if err != nil {
					log.Error("[OSS] Sent chunkserver info back to chunkmaster error: %v", err)
				}
				if statusCode != http.StatusOK {
					log.Error("[OSS] Sent chunkserver info back to chunkmaster failed STATUS CODE:%v ,result:%s", statusCode, result)
				}
			}
		}
	}()
	return http.StatusOK, []byte("")
}

func ReceiveChunkserverInfo(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	header := ctx.Req.Header
	nodeip := header.Get("nodeip")
	nodeport := header.Get("nodeport")
	groupid := header.Get("groupid")
	stderr := header.Get("stderr")
	stdout := header.Get("stdout")
	log.Info("[OSS]Chunkserver %s:%s groupid:%s start error", nodeip, nodeport, groupid)
	log.Info("[OSS]STDOUT:%s", stdout)
	log.Info("[OSS]STDERR:%s", stderr)
	return http.StatusOK, []byte("information recieced")
}
