package apiserver

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/gorilla/mux"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/oss/apiserver/chunkserver"
	"github.com/containerops/dockyard/oss/apiserver/meta"
	"github.com/containerops/dockyard/oss/apiserver/meta/mysqldriver"
	"github.com/containerops/dockyard/oss/utils"
	"github.com/containerops/dockyard/utils/setting"
)

const (
	headerSourcePath  = "Source-Path"
	headerDestPath    = "Dest-Path"
	headerPath        = "Path"
	headerIndex       = "Fragment-Index"
	headerRange       = "Bytes-Range"
	headerIsLast      = "Is-Last"
	headerVersion     = "Registry-Version"
	LimitCSNormalSize = 2
	SUCCESS           = ""
	VERSION1          = "v1"
	VERSION2          = "v2"
)

var (
	MasterUrl         string
	MasterPort        string
	Ip                string
	Port              int
	HttpsPort         int
	Router            *mux.Router
	Running           bool
	Mu                sync.Mutex
	fids              *chunkserver.Fids                      //ChunkServerGoups
	chunkServerGroups *chunkserver.ChunkServerGroups         //groupId <> []ChunkServer
	connectionPools   *chunkserver.ChunkServerConnectionPool //{"host:port":connectionPool}
	metaDriver        meta.MetaDriver
	LimitNum          int
	ConnPoolCapacity  int
	getFidRetryCount  int32
	MetadbIp          string
	MetadbPort        int
	MetadbUser        string
	MetadbPassword    string
	MetaDatabase      string
)

func InitAPI() error {
	Ip = "0.0.0.0"
	Port = setting.APIPort
	HttpsPort = setting.APIHttpsPort
	fids = chunkserver.NewFids()
	chunkServerGroups = nil
	connectionPools = nil
	getFidRetryCount = 0
	MetaDatabase = "metadb"
	err := GetChunkServerInfo()
	if err != nil {
		return fmt.Errorf("GetChunkServerInfo error: %v  \n", err)
	}

	err = GetFidRange(false)
	if err != nil {
		return fmt.Errorf("GetFidRange error: %v \n", err)
	}

	go GetFidRangeTicker()
	go GetChunkServerInfoTicker()

	err = mysqldriver.InitMeta(MetadbIp, MetadbPort, MetadbUser, MetadbPassword, MetaDatabase)
	if err != nil {
		return fmt.Errorf("Connect metadb error: %v \n", err)
	}

	metaDriver = new(mysqldriver.MysqlDriver)
	return nil
}

func DeleteFile(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	path := ctx.Req.Header.Get(headerPath)

	log.Info("[OSS][deleteFile] path: %s", path)

	var err error

	if err = metaDriver.DeleteFileMetaInfoV1(path); err != nil {
		log.Error("[OSS]delete metainfo error for path: %s, error: %s", path, err)
		return http.StatusInternalServerError, []byte(err.Error())
	}

	log.Info("[OSS][deleteFile] success. path: %s", path)
	return http.StatusNoContent, nil
}

func UploadFile(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	metaInfo, _, err := uploadFileReadParam(&ctx.Req.Header)
	if err != nil {
		log.Error("[OSS][uploadFile] read param error: %v", err)
		return http.StatusBadRequest, []byte(err.Error())
	}
	data, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[OSS]read request body error: %s", err)
		return http.StatusBadRequest, []byte(err.Error())
	}

	statusCode, err := ossupload(data, metaInfo)
	if err != nil {
		log.Error("[OSS][uploadFile] upload error: %v", err)
		return statusCode, []byte(err.Error())
	}

	if err = metaDriver.StoreMetaInfoV1(metaInfo); err != nil {
		log.Error("[OSS]store metaInfo error: %s", err)
		return http.StatusInternalServerError, []byte(err.Error())
	}

	log.Info("[OSS][postFile] success. path: %s, fragmentIndex: %d, bytesRange: %d-%d, isLast: %v", metaInfo.Path, metaInfo.Value.Index, metaInfo.Value.Start, metaInfo.Value.End, metaInfo.Value.IsLast)

	return http.StatusOK, nil
}

func GetFileInfo(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	path := ctx.Req.Header.Get(headerPath)
	log.Info("[OSS][getFileInfo] Path: %s", path)

	result, err := metaDriver.GetFileMetaInfo(path, false)
	if err != nil {
		log.Error("[OSS][getFileInfo] get metainfo error, key: %s, error: %s", path, err)
		return http.StatusInternalServerError, []byte(err.Error())
	}

	if len(result) == 0 {
		log.Info("[OSS][getFileInfo] metainfo not exists, key: %s", path)
		return http.StatusNotFound, []byte(err.Error())
	}

	resultMap := make(map[string]interface{})
	resultMap["fragment-info"] = result
	jsonResult, err := json.Marshal(resultMap)
	if err != nil {
		log.Error("json.Marshal error, key: %s, error: %s", path, err)
		return http.StatusInternalServerError, []byte(err.Error())
	}

	log.Info("[getFileInfo] success, path: %s, result: %s", path, string(jsonResult))
	return http.StatusOK, jsonResult
}

func DownloadFile(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	header := ctx.Req.Header
	path := header.Get(headerPath)
	fragmentIndex := header.Get(headerIndex)
	bytesRange := header.Get(headerRange)
	start, end, err := splitRange(bytesRange)
	if err != nil {
		log.Error("[OSS]splitRange error, bytesRange: %s, error: %s", bytesRange, err)
		return http.StatusBadRequest, []byte(err.Error())
	}

	index, err := strconv.ParseUint(fragmentIndex, 10, 64)
	if err != nil {
		log.Error("[OSS]parser fragmentIndex: %s, error: %s", fragmentIndex, err)
		return http.StatusBadRequest, []byte(err.Error())
	}

	log.Info("[OSS]path: %s, fragmentIndex: %d, bytesRange: %d-%d", path, index, start, end)

	metaInfoValue := &meta.MetaInfoValue{
		Index: index,
		Start: start,
		End:   end,
	}
	metaInfo := &meta.MetaInfo{Path: path, Value: metaInfoValue}
	log.Debug("[OSS]metaInfo: %s", metaInfo)

	chunkServer, err := getOneNormalChunkServer(metaInfo)
	if err != nil {
		if err.Error() == "fragment metainfo not found" {
			return http.StatusNotFound, []byte(err.Error())
		} else {
			return http.StatusInternalServerError, []byte(err.Error())
		}
	}

	connPools := GetConnectionPools()
	conn, err := connPools.GetConn(chunkServer)
	log.Info("downloadFile getconnection success")
	if err != nil {
		log.Error("downloadFile getconnection error: %v", err)
		return http.StatusInternalServerError, []byte(err.Error())
	}

	data, err := chunkServer.GetData(metaInfo.Value, conn.(*chunkserver.PooledConn))
	if err != nil {
		conn.Close()
		connPools.ReleaseConn(conn)
		checkErrorAndConnPool(err, chunkServer, connPools)
		return http.StatusInternalServerError, []byte(err.Error())
	}

	log.Info("[OSS][downloadFile] success. path: %s, fragmentIndex: %d, bytesRange: %d-%d", path, index, start, end)
	connPools.ReleaseConn(conn)
	return http.StatusOK, data
}

func splitRange(bytesRange string) (uint64, uint64, error) {
	var start, end uint64

	fmt.Sscanf(bytesRange, "%d-%d", &start, &end)
	if start >= end {
		return 0, 0, fmt.Errorf("bytesRange error!")
	}

	return start, end, nil
}

func getOneNormalChunkServer(mi *meta.MetaInfo) (*chunkserver.ChunkServer, error) {
	fmt.Printf("[OSS]getOneNormalChunkServer === begin \n")
	fmt.Printf("[OSS]metainfo: %s \n", mi)
	metaInfoValue, err := metaDriver.GetFragmentMetaInfo(mi.Path, mi.Value.Index, mi.Value.Start, mi.Value.End)
	if err != nil {
		fmt.Errorf("[OSS]GetFragmentMetaInfo: %s, error: %s \n", mi, err)
		return nil, err
	}

	if metaInfoValue == nil {
		fmt.Errorf("[OSS]fragment metainfo not found, path: %s, index: %d, start: %d, end: %d \n",
			mi.Path, mi.Value.Index, mi.Value.Start, mi.Value.End)
		return nil, fmt.Errorf("[OSS]fragment metainfo not found \n")
	}

	mi.Value = metaInfoValue
	fmt.Printf("[OSS]getOneNormalChunkServer, metaInfo: %s \n", mi)
	fmt.Printf("[OSS]groupId :%d \n", mi.Value.GroupId)

	groupId := strconv.Itoa(int(mi.Value.GroupId))
	groups := GetChunkServerGroups()
	servers, ok := groups.GroupMap[groupId]
	if !ok {
		fmt.Errorf("[OSS]getOneNormalChunkServer do not exist group: %s \n", groupId)
		return nil, fmt.Errorf("[OSS]do not exist group: %s \n", groupId)
	}

	index := rand.Int() % len(servers)
	server := servers[index]
	if server.Status == chunkserver.RW_STATUS {
		fmt.Printf("[OSS]get an chunkserver: %s \n", server)
		return &server, nil
	}

	for i := 0; i < len(servers); i++ {
		server = servers[i]
		if server.Status == chunkserver.RW_STATUS {
			fmt.Printf("[OSS]get an chunkserver: %s \n", server)
			return &server, nil
		}
	}
	fmt.Errorf("[OSS]can not find an available chunkserver, metainfo: %s \n", mi)
	return nil, fmt.Errorf("[OSS]can not find an available chunkserver")
}

func GetConnectionPools() *chunkserver.ChunkServerConnectionPool {
	Mu.Lock()
	connectionPool := connectionPools
	Mu.Unlock()
	return connectionPool
}

func checkErrorAndConnPool(err error, chunkServer *chunkserver.ChunkServer, connPools *chunkserver.ChunkServerConnectionPool) {
	if "EOF" == err.Error() {
		err := connPools.CheckConnPool(chunkServer)
		if err != nil {
			fmt.Errorf("CheckConnPool error: %v \n", err)
		}
	}
}

func GetChunkServerGroups() *chunkserver.ChunkServerGroups {
	Mu.Lock()
	groups := chunkServerGroups
	Mu.Unlock()
	return groups
}

func uploadFileReadParam(header *http.Header) (*meta.MetaInfo, string, error) {
	path := header.Get(headerPath)
	fragmentIndex := header.Get(headerIndex)
	bytesRange := header.Get(headerRange)
	isLast := header.Get(headerIsLast)
	version := header.Get(headerVersion)

	start, end, err := splitRange(bytesRange)
	if err != nil {
		fmt.Errorf("[OSS]splitRange error: %s \n", err)
		return nil, version, err
	}

	last := false
	if isLast == "true" || isLast == "TRUE" {
		last = true
	}

	index, err := strconv.ParseUint(fragmentIndex, 10, 64)
	if err != nil {
		fmt.Errorf("[OSS]parse fragmentIndex error: %s \n", err)
		return nil, version, err
	}

	fmt.Printf("[OSS][uploadFileReadParam] path: %s, fragmentIndex: %d, bytesRange: %d-%d, isLast: %v \n", path, index, start, end, last)

	metaInfoValue := &meta.MetaInfoValue{
		Index:  index,
		Start:  start,
		End:    end,
		IsLast: last,
	}
	metaInfo := &meta.MetaInfo{Path: path, Value: metaInfoValue}
	return metaInfo, version, nil
}

func ossupload(data []byte, metaInfo *meta.MetaInfo) (int, error) {
	chunkServers, err := selectChunkServerGroupComplex(int64(metaInfo.Value.End - metaInfo.Value.Start))
	if err != nil {
		fmt.Errorf("[OSS][upload] select ChunkServerGroup error: %s \n", err)
		return http.StatusInternalServerError, err
	}

	fmt.Printf("[OSS]ChunkServerGroup: %s \n", chunkServers)

	fileId, err := getFid()
	if err != nil {
		fmt.Errorf("[OSS][upload] get fileId error: %s \n", err)
		return http.StatusInternalServerError, err
	}

	var rangeSize uint64
	rangeSize = metaInfo.Value.End - metaInfo.Value.Start
	if len(data) != int(rangeSize) {
		fmt.Errorf("[OSS]the data length is: %d, rangeSize is: %d \n", len(data), rangeSize)
		return http.StatusBadRequest, fmt.Errorf("length of data: %d != range: %d \n", len(data), rangeSize)
	}

	fmt.Printf("[OSS]begin to upload concurrently \n")

	var normal int = 0
	for i := 0; i < len(chunkServers); i++ {
		if chunkServers[i].Status == chunkserver.RW_STATUS {
			normal++
		}
	}

	ch := make(chan string, normal)
	for i := 0; i < len(chunkServers); i++ {
		if chunkServers[i].Status == chunkserver.RW_STATUS {
			go postFileConcurrence(&chunkServers[i], data, ch, fileId)
		}
	}

	fmt.Printf("[OSS]upload result, num: %d \n", normal)
	err = handlePostResult(ch, normal)
	if err != nil {
		fmt.Errorf("[OSS]upload error: %s \n", err)
		return http.StatusInternalServerError, err
	}

	fmt.Printf("[OSS]upload end \n")
	metaInfo.Value.FileId = fileId
	metaInfo.Value.GroupId = uint16(chunkServers[0].GroupId)

	return http.StatusOK, nil
}

func selectChunkServerGroupComplex(size int64) ([]chunkserver.ChunkServer, error) {
	if size <= 0 {
		fmt.Errorf("[OSS]data size: %d <= 0 \n")
		return nil, fmt.Errorf("data size: %d <= 0 \n", size)
	}

	groups := GetChunkServerGroups()
	var totalNum int = len(groups.GroupMap)
	var selectNum int = totalNum/10 + 3
	minHeap := chunkserver.NewMinHeap(selectNum)

	for groupId, servers := range groups.GroupMap {
		var minMaxFreeSpace int64 = math.MaxInt64
		var normalNum int = 0
		var avilable = true
		var pendingWrites = 0
		var writingCount = 0

		length := len(servers)
		// skip empty group and transfering... group
		if length == 0 || servers[0].GlobalStatus != chunkserver.GLOBAL_NORMAL_STATUS {
			continue
		}

		for index := 0; index < length; index++ {
			server := servers[index]

			if server.Status != chunkserver.ERR_STATUS && server.Status != chunkserver.RW_STATUS {
				avilable = false
				break
			}

			if server.Status == chunkserver.ERR_STATUS {
				continue
			}

			if server.Status == chunkserver.RW_STATUS {
				normalNum += 1
			}

			if server.MaxFreeSpace < minMaxFreeSpace {
				minMaxFreeSpace = server.MaxFreeSpace
			}

			if server.PendingWrites > pendingWrites {
				pendingWrites = server.PendingWrites
			}

			if server.WritingCount > writingCount {
				writingCount = server.WritingCount
			}
		}

		if avilable && minMaxFreeSpace > size && normalNum >= LimitNum {
			minHeap.AddElement(groupId, minMaxFreeSpace, pendingWrites, writingCount)
		}
	}
	if minHeap.GetSize() < selectNum {
		selectNum = minHeap.GetSize()
	}

	if selectNum == 0 {
		fmt.Errorf("[OSS]selectNum == 0, there's not an avaiable chunkserver \n")
		return nil, fmt.Errorf("[OSS]there's not an avaiable chunkserver \n")
	}

	minHeap.BuildMinHeapSecondary()

	fmt.Printf("[OSS]minHeap: %s \n", minHeap)

	index := rand.Int() % selectNum
	fmt.Printf("[OSS]index: %d \n", index)
	resultGroupId, err := minHeap.GetElementGroupId(index)

	if err != nil {
		fmt.Errorf("[OSS]can not find an available chunkserver: %s \n", err)
		return nil, fmt.Errorf("[OSS]can not find an available chunkserver \n")
	}

	fmt.Printf("[OSS]resultGroupId: %s, chunkServers: %v \n", resultGroupId, groups.GroupMap[resultGroupId])
	return groups.GroupMap[resultGroupId], nil
}

func postFileConcurrence(chunkServer *chunkserver.ChunkServer, data []byte, c chan string, fileId uint64) {
	fmt.Printf("[OSS]postFileConcurrence === begin to get connection \n")
	fmt.Printf("[OSS]chunkServer: %v \n", chunkServer)

	connPools := GetConnectionPools()
	if connPools == nil {
		fmt.Errorf("[OSS]connectionPools is nil \n")
		c <- "connectionPools is nil"
		return
	}

	fmt.Printf("[OSS]fid is: %d \n", fileId)
	fmt.Printf("[OSS]connPools: %v, %s \n", connPools, connPools)

	conn, err := connPools.GetConn(chunkServer)
	fmt.Printf("[OSS]connection is: %v \n", conn)

	if err != nil {
		fmt.Errorf("[OSS]can not get connection: %s \n", err.Error())
		c <- err.Error()
		return
	}

	fmt.Printf("[OSS]begin to upload data \n")
	err = chunkServer.PutData(data, conn.(*chunkserver.PooledConn), fileId)
	if err != nil {
		fmt.Errorf("[OSS]upload data failed: %s \n", err)
		conn.Close()
		connectionPools.ReleaseConn(conn)
		c <- err.Error()
		checkErrorAndConnPool(err, chunkServer, connPools)
		return
	}

	fmt.Printf("[OSS]upload data success \n")
	c <- SUCCESS
	fmt.Printf("[OSS]set SUCCESS to chan \n")

	connPools.ReleaseConn(conn)
	fmt.Printf("[OSS]elease connection success \n")
}

func getFid() (uint64, error) {
	fileId, err := fids.GetFid()
	if err != nil {

		var count int32 = 1
		var init int32 = 0
		swapped := atomic.CompareAndSwapInt32(&getFidRetryCount, init, count)
		if !swapped {
			fmt.Errorf("[OSS]another goroutine is trying to get fid, waiting \n")
			filedId, err := fids.GetFidWait()
			if err != nil {
				return 0, err
			}
			return filedId, nil
		}

		fmt.Println("[OSS]=== try to get fid range === begin === \n")
		defer atomic.CompareAndSwapInt32(&getFidRetryCount, count, init)

		err1 := GetFidRange(false)
		fmt.Println("[OSS]=== try to get fid range === end === \n")

		if err1 != nil {
			fmt.Errorf("[OSS]getFid try to get fid failed: %v \n", err1)
			return 0, err1
		}

		fileId, err1 = fids.GetFid()
		if err1 != nil {
			fmt.Errorf("[OSS]getFid error: %v \n", err1)
			return 0, err1
		}
	}

	return fileId, nil
}

func GetFidRange(mergeWait bool) error {
	if !fids.IsShortage() {
		return nil
	}

	byteData, statusCode, err := util.Call("GET", "http://"+MasterUrl+":"+MasterPort, "/cm/v1/chunkmaster/fid", nil, nil)
	if err != nil {
		fmt.Errorf("[OSS]GetChunkServerInfo response code: %d, err: %s \n", statusCode, err)
		return err
	}

	if statusCode != http.StatusOK {
		fmt.Errorf("[OSS]response code: %d \n", statusCode)
		return fmt.Errorf("statusCode error: %d \n", statusCode)
	}

	fmt.Printf("[OSS]GetFidRange data: %s \n \n", string(byteData))

	newFids := chunkserver.NewFids()
	err = json.Unmarshal(byteData, &newFids)
	if err != nil {
		fmt.Errorf("[OSS]GetFidRange json.Unmarshal response data error: %s \n", err)
		return err
	}

	fids.Merge(newFids.Start, newFids.End, mergeWait)
	return nil
}

func GetChunkServerInfo() error {
	byteData, statusCode, err := util.Call("GET", "http://"+MasterUrl+":"+MasterPort, "/cm/v1/chunkmaster/route", nil, nil)
	if err != nil {
		fmt.Errorf("[OSS]GetChunkServerInfo response code: %d, error: %v \n", statusCode, err)
		return err
	}

	if statusCode != http.StatusOK {
		fmt.Errorf("[OSS]response code: %d \n", statusCode)
		return fmt.Errorf("[OSS]statusCode error: %d \n", statusCode)
	}

	infos := make(map[string][]chunkserver.ChunkServer)
	err = json.Unmarshal(byteData, &infos)
	if err != nil {
		fmt.Errorf("[OSS]json.Unmarshal response data error: %s", err)
		return err
	}
	handleChunkServerInfo(infos)
	return nil
}

func handleChunkServerInfo(infos map[string][]chunkserver.ChunkServer) {
	var (
		delServers []*chunkserver.ChunkServer
		addServers []*chunkserver.ChunkServer
	)

	newChunkServerGroups := &chunkserver.ChunkServerGroups{
		GroupMap: infos,
	}
	oldChunkServerGroups := GetChunkServerGroups()

	if oldChunkServerGroups == nil {
		delServers, addServers = serverInfoDiff(infos, nil)
	} else {
		delServers, addServers = serverInfoDiff(infos, oldChunkServerGroups.GroupMap)
	}

	if len(delServers) == 0 && len(addServers) == 0 {
		ReplaceChunkServerGroups(newChunkServerGroups)
		return
	}

	oldConnectionPool := GetConnectionPools()
	newConnectionPool := chunkserver.NewChunkServerConnectionPool()

	if oldConnectionPool != nil {
		fmt.Printf("oldConnectionPool: %v \n", oldConnectionPool)
		for key, connectionPool := range oldConnectionPool.Pools {
			newConnectionPool.AddExistPool(key, connectionPool)
		}
	}

	if len(delServers) != 0 {
		for index := 0; index < len(delServers); index++ {
			fmt.Printf("delete chunkserver: %v \n", delServers[index])
			newConnectionPool.RemovePool(delServers[index])
		}
	}

	if len(addServers) != 0 {
		for index := 0; index < len(addServers); index++ {
			fmt.Printf("add chunkserver: %v \n", addServers[index])
			newConnectionPool.AddPool(addServers[index], ConnPoolCapacity)
		}
	}

	newChunkServerGroups.Print()

	ReplaceConnPoolsAndChunkServerGroups(newConnectionPool, newChunkServerGroups)

	if len(delServers) != 0 && oldConnectionPool != nil {
		for index := 0; index < len(delServers); index++ {
			oldConnectionPool.RemoveAndClosePool(delServers[index])
		}
	}
}

func serverInfoDiff(newInfo, oldInfo map[string][]chunkserver.ChunkServer) (delServers, addServers []*chunkserver.ChunkServer) {
	addServers = infoDiff(newInfo, oldInfo)
	delServers = infoDiff(oldInfo, newInfo)

	return delServers, addServers
}

//diff = info1 - (the intersection info1 and info2  )
func infoDiff(info1, info2 map[string][]chunkserver.ChunkServer) []*chunkserver.ChunkServer {
	diffServers := make([]*chunkserver.ChunkServer, 0)

	for groupId, servers1 := range info1 {
		servers2, ok := info2[groupId]

		if !ok {
			for index := 0; index < len(servers1); index++ {
				diffServers = append(diffServers, &servers1[index])
			}

			continue
		}

		for index1 := 0; index1 < len(servers1); index1++ {
			server1 := servers1[index1]
			found := false

			for index2 := 0; index2 < len(servers2); index2++ {
				server2 := servers2[index2]

				if server1.HostInfoEqual(&server2) {
					found = true
					break
				}
			}

			if !found {
				diffServers = append(diffServers, &server1)
			}
		}
	}

	return diffServers
}

func ReplaceChunkServerGroups(newGroups *chunkserver.ChunkServerGroups) {
	Mu.Lock()
	chunkServerGroups = newGroups
	Mu.Unlock()
}

func ReplaceConnPoolsAndChunkServerGroups(newConnectionPool *chunkserver.ChunkServerConnectionPool, newGroups *chunkserver.ChunkServerGroups) {
	Mu.Lock()
	connectionPools = newConnectionPool
	chunkServerGroups = newGroups
	Mu.Unlock()
}

func GetFidRangeTicker() {
	timer := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-timer.C:
			err := GetFidRange(true)
			if err != nil {
				fmt.Errorf("GetFidRange error: %v \n", err)
			}
		}
	}
}

func GetChunkServerInfoTicker() {
	timer := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-timer.C:
			err := GetChunkServerInfo()
			if err != nil {
				fmt.Errorf("GetChunkServerInfoTicker error: %s \n", err)
			}
		}
	}
}

func handlePostResult(ch chan string, size int) error {
	var result, tempResult string
	var failed = false

	if ch == nil {
		fmt.Errorf("ch is nil  \n")
		return fmt.Errorf("handlePostResult ch is nil ")
	}

	fmt.Printf("len(ch): %d, size: %d \n", len(ch), size)
	for k := 0; k < size; k++ {
		tempResult = <-ch
		if len(tempResult) != 0 {
			result = tempResult
			failed = true
		}
	}

	if failed {
		fmt.Errorf("handlePostResult failed: %s \n", result)
		return fmt.Errorf(result)
	}

	return nil
}
