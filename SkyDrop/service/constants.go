package service

import (
	"SyncAgent/utils"
	"encoding/json"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var instance *Constants
var mutex sync.Mutex

type PathType struct {
	Path string `json:"path"`
	Type string `json:"type"`
}
type StorageConfig struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint"`
	AK        string `json:"ak"`
	SK        string `json:"sk"`
	Bucket    string `json:"bucket"`
	LocalPath string `json:"localPath"`
}
type SyncFilter struct {
	Filter   []string `json:"filter"`
	Type     string   `json:"type"`     //过滤类型：extension-后缀过滤, path-路径过滤
	ListType string   `json:"listType"` //名单类型：blacklist-黑名单, whitelist-白名单
}
type TagMap struct {
	Index  int    `json:"index"`
	TagKey string `json:"tagKey"`
}
type AgentTask struct {
	Paths         []PathType    `json:"paths"`
	Storage       StorageConfig `json:"storage"`
	DelTime       int64         `json:"delTime"`
	RunTime       int64         `json:"runTime"`
	WorkStartTime string        `json:"workStartTime"` //工作开始时间，| 分割多组
	WorkEndTime   string        `json:"workEndTime"`   //工作结束时间，| 分割多组
	PathPrefix    string        `json:"pathPrefix"`    //目标路径
	Status        int           `json:"status"`        //运行状态
	Tags          string        `json:"tags"`
	Filters       SyncFilter    `json:"filters"`
	MaxWorkers    int           `json:"maxWorkers"`
	FetchField    int           `json:"fetchField"`
	TagsKeyList   []TagMap      `json:"tagsKeyList"`
	FetchRegex    string        `json:"fetchRegex"`
	TagsAsPath    int           `json:"tagsAsPath"`
}
type Constants struct {
	Level     int    //最大统计深度
	Console   string //console访问地址
	TmpPath   string //临时文件存放路径
	HostSn    string //节点唯一标识
	IP        string //节点IP地址
	IPRange   string
	Version   string //Agent版本
	Workspace string

	PathMTime        map[string]map[string]int64 //源路径和最后更新时间
	PathFinishedTime map[string]map[string]int64 //源路径和上次完成时间

	TaskParams AgentTask
}

type HBResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data AgentTask `json:"data"`
}

func NewConstants() *Constants {
	if instance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if instance == nil {
			instance = new(Constants)
		}
	}
	return instance
}

func (args *Constants) RefreshCurrentIP() {
	args.IP = utils.GetOneLocalIp(args.IPRange)
}

func (args *Constants) Init(ipRange, console, version, workspace string) {
	args.IP = utils.GetOneLocalIp(ipRange)
	args.IPRange = ipRange
	args.HostSn = utils.GetLocalHostName()
	args.Console = console
	args.Version = version
	args.PathMTime = utils.ReadConfig(workspace)
	args.Workspace = workspace
}

func (args *Constants) IsPause() bool {
	if &args.TaskParams != nil {
		return args.TaskParams.Status == 2
	}
	return true
}

func (args *Constants) FromBytes(byteData []byte) {
	// 使用 json.Unmarshal 将 JSON 字符串解析到 person 对象中
	var response HBResponse
	err := json.Unmarshal(byteData, &response)
	if err != nil {
		log.Fatal("Error parsing JSON: ", err)
	} else {
		args.TaskParams = response.Data
	}
}
func (args *Constants) PrintArgs() {
	log.Println("HostSn: " + args.HostSn)
	log.Println("IP: " + args.IP)
	log.Println("Console: " + args.Console)
	log.Println("Workspace: " + args.Workspace)
}

func (args *Constants) AddPathFinishedTime(pathType PathType) {
	if args.PathFinishedTime == nil {
		args.PathFinishedTime = make(map[string]map[string]int64)
	}
	if args.PathFinishedTime[pathType.Type] == nil {
		args.PathFinishedTime[pathType.Type] = make(map[string]int64)
	}
	args.PathFinishedTime[pathType.Type][pathType.Path] = time.Now().Unix()
}
func (args *Constants) AddPathMTime(pathType PathType, mTime int64) {
	if args.PathMTime == nil {
		args.PathMTime = make(map[string]map[string]int64)
	}
	if args.PathMTime[pathType.Type] == nil {
		args.PathMTime[pathType.Type] = make(map[string]int64)
	}
	args.PathMTime[pathType.Type][pathType.Path] = mTime
	utils.WriteConfig(args.Workspace, args.PathMTime)
}

func (args *Constants) IsPathWorkTime(pathType PathType) bool {
	value, ok := args.PathFinishedTime[pathType.Type]
	if ok {
		upFinishTime, ok1 := value[pathType.Path]
		if !ok1 {
			return true
		} else {
			return time.Now().Unix()-upFinishTime > args.TaskParams.RunTime
		}
	}
	return true
}

func (args *Constants) LogFilePath(srcRootPath string) string {
	return filepath.Join(utils.GetExeDir(), "tmp", "archive_"+NewConstants().IP+"_"+filepath.Base(srcRootPath)+"_"+utils.GetMillisecond()+".log")
}

func (args *Constants) HasFilter(path string) bool {

	if &NewConstants().TaskParams == nil || &NewConstants().TaskParams.Filters == nil { //没有设置过滤
		return false //返回false表示不过滤要继续上传
	}

	filterType := NewConstants().TaskParams.Filters.Type
	filtersArray := NewConstants().TaskParams.Filters.Filter
	listType := NewConstants().TaskParams.Filters.ListType

	if filterType == "extension" { //extension-后缀过滤
		return fileExtensionFilter(listType, filtersArray, path)
	} else if filterType == "path" { //path-路径过滤
		return filePathFilter(listType, filtersArray, path)
	}
	return false
}

// 过滤验证，只要文件后缀在白名单中包含就要上传 blacklist-黑名单, whitelist-白名单
// filtersArray: 过滤列表
// path: 文件路径
// return: 是否过滤，true表示过滤（不上传），false表示不过滤（上传）
func fileExtensionFilter(filterType string, filtersArray []string, path string) bool {
	for _, _suffix := range filtersArray {
		if strings.HasSuffix(strings.ToLower(path), _suffix) {
			return filterType == "blacklist" //返回false表示不过滤要继续上传
		}
	}
	return filterType == "whitelist"
}
func filePathFilter(filterType string, filtersArray []string, path string) bool {
	for _, _value := range filtersArray {
		if strings.Contains(strings.ToLower(path), _value) {
			return filterType == "blacklist" //返回false表示不过滤要继续上传
		}
	}
	return filterType == "whitelist"
}
