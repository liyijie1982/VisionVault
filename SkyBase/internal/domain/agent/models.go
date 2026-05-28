package agent

import "time"

type Agent struct {
	ID             int64     `json:"id"`
	HostSN         string    `json:"hostSn"`
	HostName       string    `json:"hostName"`
	IP             string    `json:"ip"`
	GroupID        int64     `json:"groupId"`
	StorageID      int64     `json:"storageId"`
	SourcePaths    []string  `json:"sourcePaths"`
	PathPrefix     string    `json:"pathPrefix"`
	Version        string    `json:"version"`
	Status         int       `json:"status"`
	Tags           []string  `json:"tags"`
	LastAccessTime time.Time `json:"lastAccessTime"`
	LastCommitTime time.Time `json:"lastCommitTime"`
	Remark         string    `json:"remark"`
}

type PathType string

const (
	PathTypeSync PathType = "sync"
	PathTypeScan PathType = "scan"
)

type WorkWindow struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type Group struct {
	ID                 int64        `json:"id"`
	Name               string       `json:"name"`
	StorageID          int64        `json:"storageId"`
	IPRange            string       `json:"ipRange"`
	PathPrefix         string       `json:"pathPrefix"`
	IntervalTime       int64        `json:"intervalTime"`
	DelTimeDays        int64        `json:"delTimeDays"`
	TransferSpeedLimit int          `json:"transferSpeedLimit"`
	WorkWindows        []WorkWindow `json:"workWindows"`
	FileFilterID       int64        `json:"fileFilterId"`
	RegexID            int64        `json:"regexId"`
	ImageProcessID     int64        `json:"imageProcessId"`
	AlarmGroupID       int64        `json:"alarmGroupId"`
	LogEnabled         bool         `json:"logEnabled"`
	Status             int          `json:"status"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
}

type HeartbeatRequest struct {
	HostSN      string          `json:"hostSn"`
	IP          string          `json:"ip"`
	Version     string          `json:"version"`
	CPU         float64         `json:"cpu"`
	Mem         float64         `json:"mem"`
	StorageInfo []StorageMetric `json:"storage"`
}

type StorageMetric struct {
	Path  string  `json:"path"`
	Total float64 `json:"total"`
	Used  float64 `json:"used"`
	Free  float64 `json:"free"`
}

type TaskPath struct {
	Path string   `json:"path"`
	Type PathType `json:"type"`
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
	Type     string   `json:"type"`
	ListType string   `json:"listType"`
}

type TagMap struct {
	Index  int    `json:"index"`
	TagKey string `json:"tagKey"`
}

type HeartbeatPolicy struct {
	Paths         []TaskPath    `json:"paths"`
	Storage       StorageConfig `json:"storage"`
	DelTime       int64         `json:"delTime"`
	RunTime       int64         `json:"runTime"`
	WorkStartTime string        `json:"workStartTime"`
	WorkEndTime   string        `json:"workEndTime"`
	PathPrefix    string        `json:"pathPrefix"`
	Status        int           `json:"status"`
	Tags          string        `json:"tags"`
	Filters       SyncFilter    `json:"filters"`
	MaxWorkers    int           `json:"maxWorkers"`
	FetchField    int           `json:"fetchField"`
	TagsKeyList   []TagMap      `json:"tagsKeyList"`
	FetchRegex    string        `json:"fetchRegex"`
	TagsAsPath    int           `json:"tagsAsPath"`
}
