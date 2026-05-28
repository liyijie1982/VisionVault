package service

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"skybase/internal/config"
	"skybase/internal/domain/agent"
	"skybase/internal/domain/storage"
)

var (
	ErrAgentNotFound             = errors.New("agent not found")
	ErrAgentHostSNRequired       = errors.New("agent hostSn is required")
	ErrAgentHostNameRequired     = errors.New("agent hostName is required")
	ErrAgentIPRequired           = errors.New("agent ip is required")
	ErrAgentGroupNotFound        = errors.New("agent group not found")
	ErrAgentStorageNotFound      = errors.New("agent storage not found")
	ErrAgentHostSNDuplicate      = errors.New("agent hostSn already exists")
	ErrAgentIPDuplicate          = errors.New("agent ip already exists")
	ErrAgentDirectoryUnavailable = errors.New("agent directory service unavailable")
	ErrGroupNotFound             = errors.New("group not found")
	ErrGroupNameRequired         = errors.New("group name is required")
	ErrGroupNameExists           = errors.New("group name already exists")
	ErrGroupStorageNotFound      = errors.New("group storage not found")
	ErrGroupWorkWindowsRequired  = errors.New("group work windows are required")
	ErrGroupReferencedByAgent    = errors.New("group is referenced by agents")
	ErrStorageNotFound           = errors.New("storage not found")
	ErrStorageNameRequired       = errors.New("storage name is required")
	ErrStorageNameExists         = errors.New("storage name already exists")
	ErrStorageTypeInvalid        = errors.New("storage type must be local or s3")
	ErrStoragePathRequired       = errors.New("local storage path is required")
	ErrStorageEndpointInvalid    = errors.New("s3 storage endpoint and bucket are required")
	ErrStorageReferenced         = errors.New("storage is referenced by groups, agents, or files")
	ErrFileNotFound              = errors.New("file not found")
	ErrFileNameRequired          = errors.New("file name is required")
	ErrFilePathRequired          = errors.New("file path is required")
	ErrFileStorageNotFound       = errors.New("file storage not found")
)

type AgentService struct {
	mu              sync.RWMutex
	db              *sql.DB
	defaultPolicy   agent.HeartbeatPolicy
	currentVersion  string
	downloadID      string
	downloadFile    string
	downloadName    string
	downloadMD5     string
	lastHeartbeat   *agent.HeartbeatRequest
	lastSyncCommit  map[string]string
	lastScanPayload []byte
	agents          []AgentView
	groups          []GroupView
	storageTargets  []StorageView
	files           []FileView
	syncLogs        []SyncLogView
	scanReports     []ScanReportView
	packageHistory  []VersionView
	nextAgentID     int64
	nextGroupID     int64
	nextStorageID   int64
	nextFileID      int64
	nextLogID       int64
	nextScanID      int64
}

type VersionPayload struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Filename string `json:"filename"`
	MD5      string `json:"md5"`
}

type AgentView struct {
	ID             int64                 `json:"id"`
	HostSN         string                `json:"hostSn"`
	HostName       string                `json:"hostName"`
	IP             string                `json:"ip"`
	GroupID        int64                 `json:"groupId"`
	StorageID      int64                 `json:"storageId"`
	SourcePaths    []string              `json:"sourcePaths"`
	PathPrefix     string                `json:"pathPrefix"`
	Version        string                `json:"version"`
	Status         int                   `json:"status"`
	Tags           []string              `json:"tags"`
	LastAccessTime string                `json:"lastAccessTime"`
	LastCommitTime string                `json:"lastCommitTime"`
	Remark         string                `json:"remark"`
	CPU            float64               `json:"cpu"`
	Mem            float64               `json:"mem"`
	Storage        []agent.StorageMetric `json:"storage"`
}

type WorkWindowView struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type GroupView struct {
	ID                 int64            `json:"id"`
	Name               string           `json:"name"`
	StorageID          int64            `json:"storageId"`
	IPRange            string           `json:"ipRange"`
	PathPrefix         string           `json:"pathPrefix"`
	IntervalTime       int64            `json:"intervalTime"`
	DelTimeDays        int64            `json:"delTimeDays"`
	TransferSpeedLimit int              `json:"transferSpeedLimit"`
	WorkWindows        []WorkWindowView `json:"workWindows"`
	FileFilterID       int64            `json:"fileFilterId"`
	RegexID            int64            `json:"regexId"`
	ImageProcessID     int64            `json:"imageProcessId"`
	AlarmGroupID       int64            `json:"alarmGroupId"`
	LogEnabled         bool             `json:"logEnabled"`
	Status             int              `json:"status"`
	CreatedAt          string           `json:"createdAt"`
	UpdatedAt          string           `json:"updatedAt"`
}

type StorageView struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Type      storage.Type `json:"type"`
	Endpoint  string       `json:"endpoint"`
	AccessKey string       `json:"accessKey"`
	SecretKey string       `json:"secretKey"`
	Bucket    string       `json:"bucket"`
	Region    string       `json:"region"`
	LocalPath string       `json:"localPath"`
	Quota     int64        `json:"quota"`
	Status    int          `json:"status"`
	Remark    string       `json:"remark"`
	CreatedAt string       `json:"createdAt"`
	UpdatedAt string       `json:"updatedAt"`
}

type FileView struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Type       string   `json:"type"`
	Size       string   `json:"size"`
	Tags       []string `json:"tags"`
	ModifiedAt string   `json:"modifiedAt"`
	StorageID  int64    `json:"storageId"`
	Storage    string   `json:"storage"`
}

type SyncLogView struct {
	ID         string `json:"id"`
	AgentIP    string `json:"agentIp"`
	HostName   string `json:"hostName"`
	Path       string `json:"path"`
	StartTime  string `json:"startTime"`
	FileCount  int64  `json:"fileCount"`
	FileSize   string `json:"fileSize"`
	ErrorCount int64  `json:"errorCount"`
	LogPath    string `json:"logPath"`
	CommitTime string `json:"commitTime"`
}

type SyncLogQuery struct {
	Result string
}

type ScanBreakdown struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type ScanDirectoryStat struct {
	Name  string `json:"name"`
	Files int64  `json:"files"`
	Size  string `json:"size"`
}

type ScanReportView struct {
	ID             string              `json:"id"`
	AgentName      string              `json:"agentName"`
	GroupName      string              `json:"groupName"`
	RootPath       string              `json:"rootPath"`
	FileCount      int64               `json:"fileCount"`
	TotalSize      string              `json:"totalSize"`
	FinishedAt     string              `json:"finishedAt"`
	TypeBreakdown  []ScanBreakdown     `json:"typeBreakdown"`
	DirectoryStats []ScanDirectoryStat `json:"directoryStats"`
}

type VersionView struct {
	ID           string `json:"id"`
	Version      string `json:"version"`
	Filename     string `json:"filename"`
	MD5          string `json:"md5"`
	Status       string `json:"status"`
	UpdatedAt    string `json:"updatedAt"`
	ReleaseNotes string `json:"releaseNotes"`
	AgentCount   int64  `json:"agentCount"`
	FilePath     string `json:"-"`
}

type VersionAgentSummary struct {
	Version    string `json:"version"`
	AgentCount int64  `json:"agentCount"`
	IsLatest   bool   `json:"isLatest"`
}

type VersionListView struct {
	Items                []VersionView         `json:"items"`
	AgentVersions        []VersionAgentSummary `json:"agentVersions"`
	TotalAgents          int64                 `json:"totalAgents"`
	PublishedPackages    int64                 `json:"publishedPackages"`
	OnlineAgents         int64                 `json:"onlineAgents"`
	CurrentPackageAgents int64                 `json:"currentPackageAgents"`
}

type AgentMutation struct {
	ID             int64                 `json:"id"`
	HostSN         string                `json:"hostSn"`
	HostName       string                `json:"hostName"`
	IP             string                `json:"ip"`
	GroupID        int64                 `json:"groupId"`
	StorageID      int64                 `json:"storageId"`
	SourcePaths    []string              `json:"sourcePaths"`
	PathPrefix     string                `json:"pathPrefix"`
	Version        string                `json:"version"`
	Status         int                   `json:"status"`
	Tags           []string              `json:"tags"`
	Remark         string                `json:"remark"`
	CPU            float64               `json:"cpu"`
	Mem            float64               `json:"mem"`
	StorageMetrics []agent.StorageMetric `json:"storage"`
}

type GroupMutation struct {
	Name               string           `json:"name"`
	StorageID          int64            `json:"storageId"`
	IPRange            string           `json:"ipRange"`
	PathPrefix         string           `json:"pathPrefix"`
	IntervalTime       int64            `json:"intervalTime"`
	DelTimeDays        int64            `json:"delTimeDays"`
	TransferSpeedLimit int              `json:"transferSpeedLimit"`
	WorkWindows        []WorkWindowView `json:"workWindows"`
	FileFilterID       int64            `json:"fileFilterId"`
	RegexID            int64            `json:"regexId"`
	ImageProcessID     int64            `json:"imageProcessId"`
	AlarmGroupID       int64            `json:"alarmGroupId"`
	LogEnabled         bool             `json:"logEnabled"`
	Status             int              `json:"status"`
}

type StorageMutation struct {
	Name      string       `json:"name"`
	Type      storage.Type `json:"type"`
	Endpoint  string       `json:"endpoint"`
	AccessKey string       `json:"accessKey"`
	SecretKey string       `json:"secretKey"`
	Bucket    string       `json:"bucket"`
	Region    string       `json:"region"`
	LocalPath string       `json:"localPath"`
	Quota     int64        `json:"quota"`
	Status    int          `json:"status"`
	Remark    string       `json:"remark"`
}

type FileMutation struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Type       string   `json:"type"`
	Size       string   `json:"size"`
	Tags       []string `json:"tags"`
	ModifiedAt string   `json:"modifiedAt"`
	StorageID  int64    `json:"storageId"`
}

type AgentDirectoryEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func NewAgentService(cfg config.Config, db *sql.DB) *AgentService {
	return &AgentService{
		db:             db,
		currentVersion: cfg.Agent.Version,
		downloadID:     cfg.Agent.DownloadID,
		downloadFile:   cfg.Agent.DownloadFile,
		downloadName:   filepath.Base(cfg.Agent.DownloadFile),
		defaultPolicy: agent.HeartbeatPolicy{
			Paths:         []agent.TaskPath{},
			Storage:       agent.StorageConfig{},
			DelTime:       0,
			RunTime:       0,
			WorkStartTime: "",
			WorkEndTime:   "",
			PathPrefix:    "",
			Status:        1,
			Tags:          "",
			Filters: agent.SyncFilter{
				Filter:   []string{},
				Type:     "",
				ListType: "",
			},
			MaxWorkers:  1,
			FetchField:  0,
			TagsKeyList: []agent.TagMap{},
			FetchRegex:  "",
			TagsAsPath:  0,
		},
		lastSyncCommit: make(map[string]string),
		storageTargets: []StorageView{
			{
				ID:        1,
				Name:      "Factory Archive",
				Type:      storage.TypeLocal,
				LocalPath: "/srv/archive/factory-a",
				Quota:     12_000_000_000_000,
				Status:    1,
				Remark:    "Primary local archive for line cameras",
				CreatedAt: "2026-04-01 10:00",
				UpdatedAt: "2026-05-10 18:05",
			},
			{
				ID:        2,
				Name:      "Cold Object Vault",
				Type:      storage.TypeS3,
				Endpoint:  "https://s3.internal.example.com",
				AccessKey: "AKIA******",
				SecretKey: "********",
				Bucket:    "visionvault-cold",
				Region:    "ap-southeast-1",
				Quota:     24_000_000_000_000,
				Status:    1,
				Remark:    "Secondary retention target",
				CreatedAt: "2026-03-18 08:20",
				UpdatedAt: "2026-05-09 16:30",
			},
			{
				ID:        3,
				Name:      "QA Staging",
				Type:      storage.TypeLocal,
				LocalPath: "/mnt/staging/qa",
				Quota:     2_000_000_000_000,
				Status:    0,
				Remark:    "Reserved for regression verification",
				CreatedAt: "2026-02-11 11:10",
				UpdatedAt: "2026-05-08 11:10",
			},
		},
		groups: []GroupView{
			{
				ID:                 1,
				Name:               "Assembly Line A",
				StorageID:          1,
				IPRange:            "10.16.1.0/24",
				PathPrefix:         "/assembly-a",
				IntervalTime:       300,
				DelTimeDays:        30,
				TransferSpeedLimit: 6,
				WorkWindows: []WorkWindowView{
					{StartTime: "08:00", EndTime: "12:00"},
					{StartTime: "13:30", EndTime: "20:00"},
				},
				FileFilterID:   3,
				RegexID:        2,
				ImageProcessID: 1,
				AlarmGroupID:   1,
				LogEnabled:     true,
				Status:         1,
				CreatedAt:      "2026-03-01 09:00",
				UpdatedAt:      "2026-05-10 09:10",
			},
			{
				ID:                 2,
				Name:               "Regional Quality Labs",
				StorageID:          2,
				IPRange:            "10.18.12.0/24",
				PathPrefix:         "/quality-labs",
				IntervalTime:       600,
				DelTimeDays:        90,
				TransferSpeedLimit: 4,
				WorkWindows: []WorkWindowView{
					{StartTime: "00:00", EndTime: "23:59"},
				},
				FileFilterID:   2,
				RegexID:        1,
				ImageProcessID: 2,
				AlarmGroupID:   2,
				LogEnabled:     false,
				Status:         1,
				CreatedAt:      "2026-03-19 15:10",
				UpdatedAt:      "2026-05-07 17:20",
			},
		},
		agents: []AgentView{
			{
				ID:             1,
				HostSN:         "SN-AF-001",
				HostName:       "line-a-host-01",
				IP:             "10.16.1.18",
				GroupID:        1,
				StorageID:      1,
				PathPrefix:     "/assembly-a",
				Version:        "1.9.3.3",
				Status:         1,
				Tags:           []string{"assembly", "line-a", "camera"},
				LastAccessTime: "2026-05-11 17:41",
				LastCommitTime: "2026-05-11 17:36",
				Remark:         "Primary collector for station A1",
				CPU:            42,
				Mem:            58,
				Storage: []agent.StorageMetric{
					{Path: "C:", Total: 512, Used: 281, Free: 231},
					{Path: "D:", Total: 2048, Used: 1610, Free: 438},
				},
			},
			{
				ID:             2,
				HostSN:         "SN-AF-014",
				HostName:       "line-a-host-14",
				IP:             "10.16.1.43",
				GroupID:        1,
				StorageID:      1,
				PathPrefix:     "/assembly-a",
				Version:        "1.9.2.8",
				Status:         0,
				Tags:           []string{"assembly", "line-a", "offline"},
				LastAccessTime: "2026-05-11 15:10",
				LastCommitTime: "2026-05-11 14:54",
				Remark:         "Pending network check",
				CPU:            0,
				Mem:            0,
				Storage: []agent.StorageMetric{
					{Path: "C:", Total: 512, Used: 325, Free: 187},
				},
			},
			{
				ID:             3,
				HostSN:         "SN-QL-008",
				HostName:       "lab-node-08",
				IP:             "10.18.12.66",
				GroupID:        2,
				StorageID:      2,
				PathPrefix:     "/quality-labs",
				Version:        "1.9.3.3",
				Status:         1,
				Tags:           []string{"quality", "lab", "night-shift"},
				LastAccessTime: "2026-05-11 17:42",
				LastCommitTime: "2026-05-11 17:35",
				Remark:         "Night shift ingest node",
				CPU:            67,
				Mem:            73,
				Storage: []agent.StorageMetric{
					{Path: "/", Total: 1024, Used: 611, Free: 413},
				},
			},
		},
		files: []FileView{
			{
				ID:         "file-1",
				Name:       "station-a1-20260511-090000.jpg",
				Path:       "/assembly-a/2026/05/11/station-a1-20260511-090000.jpg",
				Type:       "jpg",
				Size:       "18.4 MB",
				Tags:       []string{"assembly", "station-a1"},
				ModifiedAt: "2026-05-11 09:00",
				StorageID:  1,
				Storage:    "Factory Archive",
			},
			{
				ID:         "file-2",
				Name:       "lab-camera-08-20260511-080500.png",
				Path:       "/quality-labs/2026/05/11/lab-camera-08-20260511-080500.png",
				Type:       "png",
				Size:       "5.8 MB",
				Tags:       []string{"lab", "night-shift"},
				ModifiedAt: "2026-05-11 08:05",
				StorageID:  2,
				Storage:    "Cold Object Vault",
			},
			{
				ID:         "file-3",
				Name:       "inspection-report-20260510.zip",
				Path:       "/quality-labs/reports/inspection-report-20260510.zip",
				Type:       "zip",
				Size:       "420 MB",
				Tags:       []string{"report", "archive"},
				ModifiedAt: "2026-05-10 23:10",
				StorageID:  2,
				Storage:    "Cold Object Vault",
			},
		},
		syncLogs: []SyncLogView{},
		scanReports: []ScanReportView{
			{
				ID:         "scan-1",
				AgentName:  "line-a-host-01",
				GroupName:  "Assembly Line A",
				RootPath:   "D:/CameraArchive/Snapshots",
				FileCount:  143221,
				TotalSize:  "3.8 TB",
				FinishedAt: "2026-05-11 17:00",
				TypeBreakdown: []ScanBreakdown{
					{Name: "jpg", Value: 102341},
					{Name: "png", Value: 31220},
					{Name: "zip", Value: 9660},
				},
				DirectoryStats: []ScanDirectoryStat{
					{Name: "2026/05/11", Files: 11220, Size: "281 GB"},
					{Name: "2026/05/10", Files: 10918, Size: "274 GB"},
					{Name: "2026/05/09", Files: 10321, Size: "258 GB"},
				},
			},
			{
				ID:         "scan-2",
				AgentName:  "lab-node-08",
				GroupName:  "Regional Quality Labs",
				RootPath:   "/data/inspection/incoming",
				FileCount:  28420,
				TotalSize:  "790 GB",
				FinishedAt: "2026-05-11 16:20",
				TypeBreakdown: []ScanBreakdown{
					{Name: "png", Value: 18120},
					{Name: "json", Value: 7820},
					{Name: "mp4", Value: 2480},
				},
				DirectoryStats: []ScanDirectoryStat{
					{Name: "batch-212", Files: 8921, Size: "230 GB"},
					{Name: "batch-211", Files: 8053, Size: "217 GB"},
					{Name: "batch-210", Files: 7130, Size: "198 GB"},
				},
			},
		},
		packageHistory: []VersionView{
			{
				ID:           "pkg-20260415",
				Version:      "1.9.2.8",
				Filename:     "SyncAgent-install-std-x86-1.9.2.8.exe",
				MD5:          "4f934c11a328663e4d7030e567145eaf",
				Status:       "Archived",
				UpdatedAt:    "2026-04-15 09:00",
				ReleaseNotes: "Scan report packaging update.",
			},
		},
		nextAgentID:   4,
		nextGroupID:   3,
		nextStorageID: 4,
		nextFileID:    4,
		nextLogID:     1,
		nextScanID:    3,
	}
}

func (s *AgentService) Heartbeat(req agent.HeartbeatRequest) agent.HeartbeatPolicy {
	s.mu.Lock()
	defer s.mu.Unlock()

	reqCopy := req
	s.lastHeartbeat = &reqCopy
	if err := s.upsertAgent(req); err != nil {
		log.Printf("agent heartbeat persistence failed hostSn=%s ip=%s err=%v", req.HostSN, req.IP, err)
	}
	return s.buildHeartbeatPolicyLocked(req)
}

func (s *AgentService) RecordCommit(ctx context.Context, values map[string]string) error {
	s.mu.Lock()
	s.lastSyncCommit = make(map[string]string, len(values))
	for key, value := range values {
		s.lastSyncCommit[key] = value
	}
	s.mu.Unlock()
	return s.recordSyncLog(ctx, values)
}

func (s *AgentService) RecordScanCommit(payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastScanPayload = append([]byte(nil), payload...)
	s.recordScanReport(payload)
}

func (s *AgentService) Version(currVersion string) (*VersionPayload, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.versionPayloadLocked(currVersion)
}

func (s *AgentService) versionPayloadLocked(currVersion string) (*VersionPayload, error) {
	if s.currentVersion == "" || s.downloadFile == "" {
		return nil, nil
	}
	if compareVersion(s.currentVersion, currVersion) <= 0 {
		return nil, nil
	}

	md5Value := s.downloadMD5
	if md5Value == "" {
		var err error
		md5Value, err = fileMD5(s.downloadFile)
		if err != nil {
			return nil, err
		}
	}

	filename := s.downloadName
	if filename == "" {
		filename = filepath.Base(s.downloadFile)
	}

	return &VersionPayload{
		ID:       s.downloadID,
		Version:  s.currentVersion,
		Filename: filename,
		MD5:      md5Value,
	}, nil
}

func (s *AgentService) Download(id string) (string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.downloadFile == "" {
		return "", "", os.ErrNotExist
	}
	if s.downloadID != "" && id != "" && id != s.downloadID {
		return "", "", errors.New("download package not found")
	}

	contentType := mime.TypeByExtension(filepath.Ext(s.downloadFile))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return s.downloadFile, contentType, nil
}

func (s *AgentService) ListAgents() []AgentView {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]AgentView(nil), s.agents...)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	for index := range items {
		items[index] = cloneAgent(items[index])
	}
	return items
}

func (s *AgentService) GetAgent(id int64) (AgentView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index := s.indexAgent(id)
	if index < 0 {
		return AgentView{}, ErrAgentNotFound
	}
	return cloneAgent(s.agents[index]), nil
}

func (s *AgentService) BrowseAgentDirectories(ctx context.Context, id int64, path string) ([]AgentDirectoryEntry, error) {
	s.mu.RLock()
	index := s.indexAgent(id)
	if index < 0 {
		s.mu.RUnlock()
		return nil, ErrAgentNotFound
	}
	agentItem := cloneAgent(s.agents[index])
	s.mu.RUnlock()

	agentIP := strings.TrimSpace(agentItem.IP)
	if agentIP == "" {
		return nil, ErrAgentDirectoryUnavailable
	}

	requestURL, err := url.Parse(fmt.Sprintf("http://%s:8765/file", agentIP))
	if err != nil {
		return nil, err
	}
	query := requestURL.Query()
	query.Set("path", strings.TrimSpace(path))
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrAgentDirectoryUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrAgentDirectoryUnavailable
	}

	var payload struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Code != 0 {
		if strings.TrimSpace(payload.Msg) != "" {
			return nil, errors.New(strings.TrimSpace(payload.Msg))
		}
		return nil, ErrAgentDirectoryUnavailable
	}

	items := make([]AgentDirectoryEntry, 0, len(payload.Data))
	for _, item := range payload.Data {
		trimmedPath := strings.TrimSpace(item.Name)
		if trimmedPath == "" {
			continue
		}
		items = append(items, AgentDirectoryEntry{
			Name: filepath.Base(trimmedPath),
			Path: trimmedPath,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items, nil
}

func (s *AgentService) CreateAgent(req AgentMutation) (AgentView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, err := s.buildAgentView(0, req, "")
	if err != nil {
		return AgentView{}, err
	}
	if s.db != nil {
		item, err = s.insertAgent(context.Background(), item)
		if err != nil {
			return AgentView{}, err
		}
	} else {
		item.ID = s.nextAgentID
		s.nextAgentID++
	}
	s.agents = append(s.agents, item)
	return cloneAgent(item), nil
}

func (s *AgentService) UpdateAgent(id int64, req AgentMutation) (AgentView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	targetID := id
	if targetID == 0 {
		targetID = req.ID
	}
	if req.ID > 0 && targetID != req.ID {
		return AgentView{}, ErrAgentNotFound
	}
	index := s.indexAgent(targetID)

	lastCommitTime := ""
	if index >= 0 {
		lastCommitTime = s.agents[index].LastCommitTime
	}
	item, err := s.buildAgentView(targetID, req, lastCommitTime)
	if err != nil {
		return AgentView{}, err
	}
	if s.db != nil {
		item, err = s.updateAgentRow(context.Background(), targetID, item)
		if err != nil {
			return AgentView{}, err
		}
	} else {
		if index < 0 {
			return AgentView{}, ErrAgentNotFound
		}
		item.ID = targetID
	}
	if index >= 0 {
		s.agents[index] = item
	} else {
		s.agents = append(s.agents, item)
		sort.Slice(s.agents, func(i, j int) bool { return s.agents[i].ID < s.agents[j].ID })
	}
	return cloneAgent(item), nil
}

func (s *AgentService) DeleteAgent(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.indexAgent(id)
	if index < 0 {
		return ErrAgentNotFound
	}
	if s.db != nil {
		if err := s.deleteAgentRow(context.Background(), id); err != nil {
			return err
		}
	}
	s.agents = append(s.agents[:index], s.agents[index+1:]...)
	return nil
}

func (s *AgentService) SyncAgentsFromDB(ctx context.Context) error {
	if s.db == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.listAgentsFromDB(ctx)
	if err != nil {
		return err
	}

	s.agents = items
	var nextID int64 = 1
	for _, item := range items {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	s.nextAgentID = nextID
	return nil
}

func (s *AgentService) ListGroups() []GroupView {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]GroupView(nil), s.groups...)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	for index := range items {
		items[index] = cloneGroup(items[index])
	}
	return items
}

func (s *AgentService) GetGroup(id int64) (GroupView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index := s.indexGroup(id)
	if index < 0 {
		return GroupView{}, ErrGroupNotFound
	}
	return cloneGroup(s.groups[index]), nil
}

func (s *AgentService) CreateGroup(req GroupMutation) (GroupView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, err := s.buildGroupView(req, "")
	if err != nil {
		return GroupView{}, err
	}
	if s.db != nil {
		item, err = s.insertGroup(context.Background(), item)
		if err != nil {
			return GroupView{}, err
		}
	} else {
		item.ID = s.nextGroupID
		s.nextGroupID++
	}
	s.groups = append(s.groups, item)
	return cloneGroup(item), nil
}

func (s *AgentService) UpdateGroup(id int64, req GroupMutation) (GroupView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.indexGroup(id)
	if index < 0 {
		return GroupView{}, ErrGroupNotFound
	}

	item, err := s.buildGroupView(req, s.groups[index].CreatedAt)
	if err != nil {
		return GroupView{}, err
	}
	if s.db != nil {
		item, err = s.updateGroupRow(context.Background(), id, item)
		if err != nil {
			return GroupView{}, err
		}
	} else {
		item.ID = id
	}
	s.groups[index] = item
	s.syncAgentsForGroupLocked(item)
	return cloneGroup(item), nil
}

func (s *AgentService) DeleteGroup(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.indexGroup(id)
	if index < 0 {
		return ErrGroupNotFound
	}
	for _, item := range s.agents {
		if item.GroupID == id {
			return ErrGroupReferencedByAgent
		}
	}
	if s.db != nil {
		if err := s.deleteGroupRow(context.Background(), id); err != nil {
			return err
		}
	}
	s.groups = append(s.groups[:index], s.groups[index+1:]...)
	return nil
}

func (s *AgentService) SyncGroupsFromDB(ctx context.Context) error {
	if s.db == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.listGroupsFromDB(ctx)
	if err != nil {
		return err
	}

	s.groups = items
	var nextID int64 = 1
	for _, item := range items {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	s.nextGroupID = nextID
	return nil
}

func (s *AgentService) ListStorageTargets() []StorageView {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]StorageView(nil), s.storageTargets...)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func (s *AgentService) GetStorageTarget(id int64) (StorageView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index := s.indexStorage(id)
	if index < 0 {
		return StorageView{}, ErrStorageNotFound
	}
	return s.storageTargets[index], nil
}

func (s *AgentService) CreateStorageTarget(req StorageMutation) (StorageView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, err := s.buildStorageView(req, "")
	if err != nil {
		return StorageView{}, err
	}
	if s.db != nil {
		item, err = s.insertStorageTarget(context.Background(), item)
		if err != nil {
			return StorageView{}, err
		}
	} else {
		item.ID = s.nextStorageID
		s.nextStorageID++
	}
	s.storageTargets = append(s.storageTargets, item)
	return item, nil
}

func (s *AgentService) UpdateStorageTarget(id int64, req StorageMutation) (StorageView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.indexStorage(id)
	if index < 0 {
		return StorageView{}, ErrStorageNotFound
	}

	item, err := s.buildStorageView(req, s.storageTargets[index].CreatedAt)
	if err != nil {
		return StorageView{}, err
	}
	if s.db != nil {
		item, err = s.updateStorageTargetRow(context.Background(), id, item)
		if err != nil {
			return StorageView{}, err
		}
	} else {
		item.ID = id
	}
	s.storageTargets[index] = item
	s.syncStorageReferencesLocked(item)
	return item, nil
}

func (s *AgentService) DeleteStorageTarget(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.indexStorage(id)
	if index < 0 {
		return ErrStorageNotFound
	}
	for _, item := range s.groups {
		if item.StorageID == id {
			return ErrStorageReferenced
		}
	}
	for _, item := range s.agents {
		if item.StorageID == id {
			return ErrStorageReferenced
		}
	}
	for _, item := range s.files {
		if item.StorageID == id {
			return ErrStorageReferenced
		}
	}
	if s.db != nil {
		if err := s.deleteStorageTargetRow(context.Background(), id); err != nil {
			return err
		}
	}
	s.storageTargets = append(s.storageTargets[:index], s.storageTargets[index+1:]...)
	return nil
}

func (s *AgentService) SyncStorageTargetsFromDB(ctx context.Context) error {
	if s.db == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items, err := s.listStorageTargetsFromDB(ctx)
	if err != nil {
		return err
	}

	s.storageTargets = items
	var nextID int64 = 1
	for _, item := range items {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	s.nextStorageID = nextID
	return nil
}

func (s *AgentService) ListFiles() []FileView {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]FileView(nil), s.files...)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	for index := range items {
		items[index] = cloneFile(items[index])
	}
	return items
}

func (s *AgentService) GetFile(id string) (FileView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index := s.indexFile(id)
	if index < 0 {
		return FileView{}, ErrFileNotFound
	}
	return cloneFile(s.files[index]), nil
}

func (s *AgentService) CreateFile(req FileMutation) (FileView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, err := s.buildFileView(req)
	if err != nil {
		return FileView{}, err
	}
	item.ID = "file-" + strconv.FormatInt(s.nextFileID, 10)
	s.nextFileID++
	s.files = append([]FileView{item}, s.files...)
	return cloneFile(item), nil
}

func (s *AgentService) UpdateFile(id string, req FileMutation) (FileView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.indexFile(id)
	if index < 0 {
		return FileView{}, ErrFileNotFound
	}

	item, err := s.buildFileView(req)
	if err != nil {
		return FileView{}, err
	}
	item.ID = id
	s.files[index] = item
	return cloneFile(item), nil
}

func (s *AgentService) DeleteFile(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.indexFile(id)
	if index < 0 {
		return ErrFileNotFound
	}
	if path := strings.TrimSpace(s.files[index].Path); path != "" {
		_ = os.Remove(path)
	}
	s.files = append(s.files[:index], s.files[index+1:]...)
	return nil
}

func (s *AgentService) UploadFile(storageID int64, filename string, src io.Reader, tags []string) (FileView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	storageItem := s.findStorage(storageID)
	if storageItem == nil {
		return FileView{}, ErrFileStorageNotFound
	}
	safeName := sanitizeFilename(filename)
	if safeName == "" {
		return FileView{}, ErrFileNameRequired
	}
	targetDir := storageUploadDir(*storageItem)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return FileView{}, err
	}
	targetPath := filepath.Join(targetDir, fmt.Sprintf("%d-%s", time.Now().UnixNano(), safeName))
	file, err := os.Create(targetPath)
	if err != nil {
		return FileView{}, err
	}
	defer file.Close()

	written, err := io.Copy(file, src)
	if err != nil {
		return FileView{}, err
	}
	item := FileView{
		ID:         "file-" + strconv.FormatInt(s.nextFileID, 10),
		Name:       safeName,
		Path:       targetPath,
		Type:       strings.TrimPrefix(strings.ToLower(filepath.Ext(safeName)), "."),
		Size:       humanizeBytes(written),
		Tags:       normalizeStringList(tags),
		ModifiedAt: nowString(),
		StorageID:  storageID,
		Storage:    storageItem.Name,
	}
	s.nextFileID++
	s.files = append([]FileView{item}, s.files...)
	return cloneFile(item), nil
}

func (s *AgentService) DownloadFile(id string) (string, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index := s.indexFile(id)
	if index < 0 {
		return "", "", ErrFileNotFound
	}
	filePath := strings.TrimSpace(s.files[index].Path)
	if filePath == "" {
		return "", "", os.ErrNotExist
	}
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return filePath, contentType, nil
}

func (s *AgentService) ListSyncLogs(ctx context.Context, query SyncLogQuery) ([]SyncLogView, error) {
	if s.db != nil {
		return s.listSyncLogsFromDB(ctx, query)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]SyncLogView(nil), s.syncLogs...)
	return filterSyncLogs(items, query), nil
}

func (s *AgentService) ListScanReports() []ScanReportView {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]ScanReportView(nil), s.scanReports...)
}

func (s *AgentService) ListVersions() (VersionListView, error) {
	if s.db != nil {
		return s.listVersionsFromDB(context.Background())
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]VersionView(nil), s.packageHistory...)
	agentVersions := summarizeAgentVersions(s.agents)
	var totalAgents int64
	for _, item := range s.agents {
		if strings.TrimSpace(item.Version) == "" {
			continue
		}
		totalAgents++
		for index := range items {
			if items[index].Version == item.Version {
				items[index].AgentCount++
				break
			}
		}
	}
	return VersionListView{
		Items:                items,
		AgentVersions:        agentVersions,
		TotalAgents:          totalAgents,
		PublishedPackages:    int64(len(items)),
		OnlineAgents:         countOnlineAgents(items, s.agents),
		CurrentPackageAgents: currentPackageAgentsFromSummary(agentVersions),
	}, nil
}

func (s *AgentService) SyncVersionsFromDB(ctx context.Context) error {
	if s.db == nil {
		return nil
	}

	_, err := s.listVersionsFromDB(ctx)
	return err
}

func (s *AgentService) listVersionsFromDB(ctx context.Context) (VersionListView, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			v.id,
			v.version,
			v.filename,
			v.file_path,
			v.md5,
			v.is_current,
			v.status,
			v.release_note,
			v.updated_at,
			COALESCE(a.agent_count, 0) AS agent_count
		FROM sync_agent_version v
		LEFT JOIN (
			SELECT version, COUNT(1) AS agent_count
			FROM sync_agent
			WHERE deleted_at IS NULL AND TRIM(version) <> ''
			GROUP BY version
		) a ON a.version = v.version
		WHERE v.deleted_at IS NULL
		ORDER BY v.is_current DESC, v.id DESC`)
	if err != nil {
		return VersionListView{}, err
	}
	defer rows.Close()

	items := []VersionView{}
	agentVersions := []VersionAgentSummary{}
	var totalAgents int64
	var publishedPackages int64
	var onlineAgents int64
	currentVersion := ""
	downloadID := ""
	downloadFile := ""
	downloadName := ""
	downloadMD5 := ""
	for rows.Next() {
		var item VersionView
		var numericID int64
		var isCurrent int
		var status int
		var releaseNote sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(&numericID, &item.Version, &item.Filename, &item.FilePath, &item.MD5, &isCurrent, &status, &releaseNote, &updatedAt, &item.AgentCount); err != nil {
			return VersionListView{}, err
		}
		item.ID = versionIDFromVersion(item.Version)
		item.Status = "Uploaded"
		if status == 0 {
			item.Status = "Disabled"
		}
		if isCurrent == 1 {
			item.Status = "Active"
			currentVersion = item.Version
			downloadID = item.ID
			downloadFile = item.FilePath
			downloadName = item.Filename
			downloadMD5 = item.MD5
		}
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		item.ReleaseNotes = releaseNote.String
		totalAgents += item.AgentCount
		publishedPackages++
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return VersionListView{}, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM sync_agent WHERE deleted_at IS NULL`).Scan(&totalAgents); err != nil {
		return VersionListView{}, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM sync_agent WHERE deleted_at IS NULL AND status = 1`).Scan(&onlineAgents); err != nil {
		return VersionListView{}, err
	}
	agentVersions, err = s.listAgentVersionSummariesFromDB(ctx)
	if err != nil {
		return VersionListView{}, err
	}

	currentPackageAgents := int64(0)
	currentPackageAgents = currentPackageAgentsFromSummary(agentVersions)

	s.mu.Lock()
	s.packageHistory = items
	if currentVersion != "" {
		s.currentVersion = currentVersion
		s.downloadID = downloadID
		s.downloadFile = downloadFile
		s.downloadName = downloadName
		s.downloadMD5 = downloadMD5
	}
	s.mu.Unlock()

	return VersionListView{
		Items:                append([]VersionView(nil), items...),
		AgentVersions:        append([]VersionAgentSummary(nil), agentVersions...),
		TotalAgents:          totalAgents,
		PublishedPackages:    publishedPackages,
		OnlineAgents:         onlineAgents,
		CurrentPackageAgents: currentPackageAgents,
	}, nil
}

func (s *AgentService) UploadVersion(version string, releaseNotes string, activate bool, filename string, src io.Reader) (VersionView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	version = strings.TrimSpace(version)
	if version == "" {
		return VersionView{}, errors.New("version is required")
	}
	safeName := sanitizeFilename(filename)
	if safeName == "" {
		return VersionView{}, errors.New("package filename is required")
	}
	targetDir := versionUploadDir(s.downloadFile)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return VersionView{}, err
	}
	targetPath := filepath.Join(targetDir, fmt.Sprintf("%s-%s", version, safeName))
	file, err := os.Create(targetPath)
	if err != nil {
		return VersionView{}, err
	}
	defer file.Close()

	hash := md5.New()
	written, err := io.Copy(io.MultiWriter(file, hash), src)
	if err != nil {
		return VersionView{}, err
	}
	md5Value := hex.EncodeToString(hash.Sum(nil))
	item := VersionView{
		ID:           versionIDFromVersion(version),
		Version:      version,
		Filename:     safeName,
		MD5:          md5Value,
		Status:       "Uploaded",
		UpdatedAt:    nowString(),
		ReleaseNotes: strings.TrimSpace(releaseNotes),
		AgentCount:   0,
		FilePath:     targetPath,
	}
	if activate {
		item.Status = "Active"
		s.currentVersion = version
		s.downloadID = item.ID
		s.downloadFile = targetPath
		s.downloadName = safeName
		s.downloadMD5 = md5Value
	}

	for index := range s.packageHistory {
		if activate && s.packageHistory[index].Status == "Active" {
			s.packageHistory[index].Status = "Archived"
		}
	}
	s.packageHistory = append([]VersionView{item}, s.packageHistory...)

	if s.db != nil {
		if err := s.saveVersionRowLocked(context.Background(), item, written, activate); err != nil {
			return VersionView{}, err
		}
	}
	return item, nil
}

func (s *AgentService) VerifyVersionMD5(id string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.findVersionLocked(id)
	if !ok {
		return nil, ErrFileNotFound
	}
	actualMD5, err := fileMD5(item.FilePath)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":       item.ID,
		"expected": item.MD5,
		"actual":   actualMD5,
		"matched":  strings.EqualFold(item.MD5, actualMD5),
	}, nil
}

func (s *AgentService) saveVersionRowLocked(ctx context.Context, item VersionView, fileSize int64, activate bool) error {
	if s.db == nil {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if activate {
		if _, err := tx.ExecContext(ctx, `UPDATE sync_agent_version SET is_current = 0 WHERE is_current = 1 AND deleted_at IS NULL`); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO sync_agent_version (version, filename, file_path, md5, file_size_bytes, is_current, status, release_note) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.Version, item.Filename, item.FilePath, item.MD5, fileSize, boolToInt(activate), 1, item.ReleaseNotes,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *AgentService) findVersionLocked(id string) (VersionView, bool) {
	for _, item := range s.packageHistory {
		if item.ID == id {
			return item, true
		}
	}
	if payload, err := s.versionPayloadLocked("0.0.0"); err == nil && payload != nil && payload.ID == id {
		return VersionView{
			ID:       payload.ID,
			Version:  payload.Version,
			Filename: payload.Filename,
			MD5:      payload.MD5,
			Status:   "Active",
			FilePath: s.downloadFile,
		}, true
	}
	return VersionView{}, false
}

func (s *AgentService) buildAgentView(id int64, req AgentMutation, lastCommit string) (AgentView, error) {
	if strings.TrimSpace(req.HostSN) == "" {
		return AgentView{}, ErrAgentHostSNRequired
	}
	if strings.TrimSpace(req.HostName) == "" {
		return AgentView{}, ErrAgentHostNameRequired
	}
	if strings.TrimSpace(req.IP) == "" {
		return AgentView{}, ErrAgentIPRequired
	}
	for _, item := range s.agents {
		if item.ID != id && strings.EqualFold(item.HostSN, strings.TrimSpace(req.HostSN)) {
			return AgentView{}, ErrAgentHostSNDuplicate
		}
		if item.ID != id && strings.EqualFold(item.IP, strings.TrimSpace(req.IP)) {
			return AgentView{}, ErrAgentIPDuplicate
		}
	}

	var linkedGroup *GroupView
	if req.GroupID > 0 {
		group := s.findGroup(req.GroupID)
		if group == nil {
			return AgentView{}, ErrAgentGroupNotFound
		}
		linkedGroup = group
	}

	storageID := req.StorageID
	if storageID == 0 && linkedGroup != nil {
		storageID = linkedGroup.StorageID
	}
	if storageID > 0 && s.findStorage(storageID) == nil {
		return AgentView{}, ErrAgentStorageNotFound
	}

	pathPrefix := strings.TrimSpace(req.PathPrefix)
	if pathPrefix == "" && linkedGroup != nil {
		pathPrefix = linkedGroup.PathPrefix
	}

	lastAccess := nowString()
	if id > 0 {
		if existing := s.findAgent(id); existing != nil && existing.LastAccessTime != "" {
			lastAccess = existing.LastAccessTime
		}
	}

	return AgentView{
		ID:             id,
		HostSN:         strings.TrimSpace(req.HostSN),
		HostName:       strings.TrimSpace(req.HostName),
		IP:             strings.TrimSpace(req.IP),
		GroupID:        req.GroupID,
		StorageID:      storageID,
		SourcePaths:    normalizeStringList(req.SourcePaths),
		PathPrefix:     pathPrefix,
		Version:        strings.TrimSpace(req.Version),
		Status:         req.Status,
		Tags:           normalizeStringList(req.Tags),
		LastAccessTime: lastAccess,
		LastCommitTime: lastCommit,
		Remark:         strings.TrimSpace(req.Remark),
		CPU:            req.CPU,
		Mem:            req.Mem,
		Storage:        cloneStorageMetrics(req.StorageMetrics),
	}, nil
}

func (s *AgentService) buildGroupView(req GroupMutation, createdAt string) (GroupView, error) {
	if strings.TrimSpace(req.Name) == "" {
		return GroupView{}, ErrGroupNameRequired
	}
	if s.findStorage(req.StorageID) == nil {
		return GroupView{}, ErrGroupStorageNotFound
	}
	workWindows := normalizeWorkWindows(req.WorkWindows)
	if len(workWindows) == 0 {
		return GroupView{}, ErrGroupWorkWindowsRequired
	}
	if createdAt == "" {
		createdAt = nowString()
	}
	return GroupView{
		Name:               strings.TrimSpace(req.Name),
		StorageID:          req.StorageID,
		IPRange:            strings.TrimSpace(req.IPRange),
		PathPrefix:         strings.TrimSpace(req.PathPrefix),
		IntervalTime:       req.IntervalTime,
		DelTimeDays:        req.DelTimeDays,
		TransferSpeedLimit: req.TransferSpeedLimit,
		WorkWindows:        workWindows,
		FileFilterID:       req.FileFilterID,
		RegexID:            req.RegexID,
		ImageProcessID:     req.ImageProcessID,
		AlarmGroupID:       req.AlarmGroupID,
		LogEnabled:         req.LogEnabled,
		Status:             req.Status,
		CreatedAt:          createdAt,
		UpdatedAt:          nowString(),
	}, nil
}

func (s *AgentService) buildStorageView(req StorageMutation, createdAt string) (StorageView, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return StorageView{}, ErrStorageNameRequired
	}
	if req.Type != storage.TypeLocal && req.Type != storage.TypeS3 {
		return StorageView{}, ErrStorageTypeInvalid
	}
	if req.Type == storage.TypeLocal && strings.TrimSpace(req.LocalPath) == "" {
		return StorageView{}, ErrStoragePathRequired
	}
	if req.Type == storage.TypeS3 && (strings.TrimSpace(req.Endpoint) == "" || strings.TrimSpace(req.Bucket) == "") {
		return StorageView{}, ErrStorageEndpointInvalid
	}
	if createdAt == "" {
		createdAt = nowString()
	}

	item := StorageView{
		Name:      name,
		Type:      req.Type,
		Endpoint:  strings.TrimSpace(req.Endpoint),
		AccessKey: strings.TrimSpace(req.AccessKey),
		SecretKey: strings.TrimSpace(req.SecretKey),
		Bucket:    strings.TrimSpace(req.Bucket),
		Region:    strings.TrimSpace(req.Region),
		LocalPath: strings.TrimSpace(req.LocalPath),
		Quota:     req.Quota,
		Status:    req.Status,
		Remark:    strings.TrimSpace(req.Remark),
		CreatedAt: createdAt,
		UpdatedAt: nowString(),
	}
	if item.Type == storage.TypeLocal {
		item.Endpoint = ""
		item.AccessKey = ""
		item.SecretKey = ""
		item.Bucket = ""
		item.Region = ""
	}
	if item.Type == storage.TypeS3 {
		item.LocalPath = ""
	}
	return item, nil
}

func (s *AgentService) listAgentsFromDB(ctx context.Context) ([]AgentView, error) {
	const query = `
		SELECT a.id, a.host_sn, a.host_name, a.ip, a.group_id, a.storage_id, a.source_paths, a.storage_metrics, a.path_prefix, a.version, a.status,
		       a.tags, a.last_access_time, a.last_commit_time, a.remark,
		       COALESCE(m.cpu_usage, 0), COALESCE(m.mem_usage, 0)
		FROM sync_agent a
		LEFT JOIN sync_agent_monitor m ON m.agent_id = a.id
		WHERE a.deleted_at IS NULL
		ORDER BY a.id ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AgentView, 0)
	for rows.Next() {
		item, scanErr := scanAgent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *AgentService) getAgentByIDDB(ctx context.Context, id int64) (AgentView, error) {
	const query = `
		SELECT a.id, a.host_sn, a.host_name, a.ip, a.group_id, a.storage_id, a.source_paths, a.storage_metrics, a.path_prefix, a.version, a.status,
		       a.tags, a.last_access_time, a.last_commit_time, a.remark,
		       COALESCE(m.cpu_usage, 0), COALESCE(m.mem_usage, 0)
		FROM sync_agent a
		LEFT JOIN sync_agent_monitor m ON m.agent_id = a.id
		WHERE a.id = ? AND a.deleted_at IS NULL
	`

	row := s.db.QueryRowContext(ctx, query, id)
	item, err := scanAgent(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AgentView{}, ErrAgentNotFound
		}
		return AgentView{}, err
	}
	return item, nil
}

func (s *AgentService) insertAgent(ctx context.Context, item AgentView) (AgentView, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AgentView{}, err
	}
	defer tx.Rollback()

	tagsJSON, err := json.Marshal(item.Tags)
	if err != nil {
		return AgentView{}, err
	}
	sourcePathsJSON, err := json.Marshal(item.SourcePaths)
	if err != nil {
		return AgentView{}, err
	}
	storageMetricsJSON, err := json.Marshal(item.Storage)
	if err != nil {
		return AgentView{}, err
	}

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO sync_agent (host_sn, host_name, ip, group_id, storage_id, source_paths, storage_metrics, path_prefix, version, status, tags, last_access_time, last_commit_time, remark) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.HostSN,
		item.HostName,
		item.IP,
		item.GroupID,
		item.StorageID,
		sourcePathsJSON,
		storageMetricsJSON,
		item.PathPrefix,
		item.Version,
		item.Status,
		tagsJSON,
		nullTimeValue(item.LastAccessTime),
		nullTimeValue(item.LastCommitTime),
		item.Remark,
	)
	if err != nil {
		return AgentView{}, mapAgentWriteError(err)
	}

	agentID, err := result.LastInsertId()
	if err != nil {
		return AgentView{}, err
	}
	if err := s.upsertAgentMonitorTx(ctx, tx, agentID, item); err != nil {
		return AgentView{}, err
	}
	if err := tx.Commit(); err != nil {
		return AgentView{}, err
	}
	return s.getAgentByIDDB(ctx, agentID)
}

func (s *AgentService) updateAgentRow(ctx context.Context, id int64, item AgentView) (AgentView, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AgentView{}, err
	}
	defer tx.Rollback()

	tagsJSON, err := json.Marshal(item.Tags)
	if err != nil {
		return AgentView{}, err
	}
	sourcePathsJSON, err := json.Marshal(item.SourcePaths)
	if err != nil {
		return AgentView{}, err
	}
	storageMetricsJSON, err := json.Marshal(item.Storage)
	if err != nil {
		return AgentView{}, err
	}

	result, err := tx.ExecContext(
		ctx,
		`UPDATE sync_agent SET host_sn = ?, host_name = ?, ip = ?, group_id = ?, storage_id = ?, source_paths = ?, storage_metrics = ?, path_prefix = ?, version = ?, status = ?, tags = ?, last_access_time = ?, last_commit_time = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		item.HostSN,
		item.HostName,
		item.IP,
		item.GroupID,
		item.StorageID,
		sourcePathsJSON,
		storageMetricsJSON,
		item.PathPrefix,
		item.Version,
		item.Status,
		tagsJSON,
		nullTimeValue(item.LastAccessTime),
		nullTimeValue(item.LastCommitTime),
		item.Remark,
		id,
	)
	if err != nil {
		return AgentView{}, mapAgentWriteError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return AgentView{}, err
	}
	if affected == 0 {
		return AgentView{}, ErrAgentNotFound
	}
	if err := s.upsertAgentMonitorTx(ctx, tx, id, item); err != nil {
		return AgentView{}, err
	}
	if err := tx.Commit(); err != nil {
		return AgentView{}, err
	}
	return s.getAgentByIDDB(ctx, id)
}

func (s *AgentService) deleteAgentRow(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM sync_agent_monitor WHERE agent_id = ?`, id); err != nil {
		return err
	}
	result, err := tx.ExecContext(
		ctx,
		`UPDATE sync_agent SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrAgentNotFound
	}
	return tx.Commit()
}

func (s *AgentService) upsertAgentMonitorTx(ctx context.Context, tx *sql.Tx, agentID int64, item AgentView) error {
	storageJSON, err := json.Marshal(item.Storage)
	if err != nil {
		return err
	}

	heartbeatAt := item.LastAccessTime
	if strings.TrimSpace(heartbeatAt) == "" {
		heartbeatAt = nowString()
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO sync_agent_monitor (agent_id, cpu_usage, mem_usage, disk_usage_json, heartbeat_at) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE cpu_usage = VALUES(cpu_usage), mem_usage = VALUES(mem_usage), disk_usage_json = VALUES(disk_usage_json), heartbeat_at = VALUES(heartbeat_at)`,
		agentID,
		item.CPU,
		item.Mem,
		storageJSON,
		nullTimeValue(heartbeatAt),
	)
	return err
}

func (s *AgentService) EnsureSchema(ctx context.Context) error {
	if s.db == nil {
		return nil
	}

	var count int
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT COUNT(1) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sync_agent' AND COLUMN_NAME = 'source_paths'`,
	).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := s.db.ExecContext(ctx, `ALTER TABLE sync_agent ADD COLUMN source_paths JSON NULL AFTER storage_id`); err != nil {
			return err
		}
	}
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT COUNT(1) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sync_agent' AND COLUMN_NAME = 'storage_metrics'`,
	).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := s.db.ExecContext(ctx, `ALTER TABLE sync_agent ADD COLUMN storage_metrics JSON NULL AFTER source_paths`); err != nil {
			return err
		}
	}
	if _, err := s.db.ExecContext(
		ctx,
		`UPDATE sync_agent a
		LEFT JOIN sync_agent_monitor m ON m.agent_id = a.id
		SET a.storage_metrics = m.disk_usage_json
		WHERE a.deleted_at IS NULL
		  AND (a.storage_metrics IS NULL OR JSON_LENGTH(a.storage_metrics) = 0)
		  AND m.disk_usage_json IS NOT NULL`,
	); err != nil {
		return err
	}
	return nil
}

func (s *AgentService) seedAgentsIfEmpty(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM sync_agent WHERE deleted_at IS NULL`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	for _, item := range s.agents {
		if _, err := s.insertAgent(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func scanAgent(scanner interface {
	Scan(dest ...any) error
}) (AgentView, error) {
	var item AgentView
	var tagsRaw []byte
	var sourcePathsRaw []byte
	var storageMetricsRaw []byte
	var lastAccess sql.NullTime
	var lastCommit sql.NullTime
	if err := scanner.Scan(
		&item.ID,
		&item.HostSN,
		&item.HostName,
		&item.IP,
		&item.GroupID,
		&item.StorageID,
		&sourcePathsRaw,
		&storageMetricsRaw,
		&item.PathPrefix,
		&item.Version,
		&item.Status,
		&tagsRaw,
		&lastAccess,
		&lastCommit,
		&item.Remark,
		&item.CPU,
		&item.Mem,
	); err != nil {
		return AgentView{}, err
	}
	if len(tagsRaw) > 0 {
		if err := json.Unmarshal(tagsRaw, &item.Tags); err != nil {
			return AgentView{}, err
		}
	}
	if len(sourcePathsRaw) > 0 {
		if err := json.Unmarshal(sourcePathsRaw, &item.SourcePaths); err != nil {
			return AgentView{}, err
		}
	}
	if len(storageMetricsRaw) > 0 {
		if err := json.Unmarshal(storageMetricsRaw, &item.Storage); err != nil {
			return AgentView{}, err
		}
	}
	if lastAccess.Valid {
		item.LastAccessTime = lastAccess.Time.Format(mysqlTimeFormat)
	}
	if lastCommit.Valid {
		item.LastCommitTime = lastCommit.Time.Format(mysqlTimeFormat)
	}
	item.Tags = normalizeStringList(item.Tags)
	item.SourcePaths = normalizeStringList(item.SourcePaths)
	item.Storage = cloneStorageMetrics(item.Storage)
	return item, nil
}

func mapAgentWriteError(err error) error {
	switch mysqlDuplicateField(err) {
	case "uk_sync_agent_host_sn":
		return ErrAgentHostSNDuplicate
	case "uk_sync_agent_ip":
		return ErrAgentIPDuplicate
	default:
		return err
	}
}

func (s *AgentService) listGroupsFromDB(ctx context.Context) ([]GroupView, error) {
	const query = `
		SELECT id, name, storage_id, ip_range, path_prefix, interval_time, del_time_days, max_workers,
		       work_windows, file_filter_id, regex_id, image_process_id, alarm_group_id,
		       log_enabled, status, created_at, updated_at
		FROM sync_agent_group
		WHERE deleted_at IS NULL
		ORDER BY id ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]GroupView, 0)
	for rows.Next() {
		item, scanErr := scanGroup(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *AgentService) getGroupByIDDB(ctx context.Context, id int64) (GroupView, error) {
	const query = `
		SELECT id, name, storage_id, ip_range, path_prefix, interval_time, del_time_days, max_workers,
		       work_windows, file_filter_id, regex_id, image_process_id, alarm_group_id,
		       log_enabled, status, created_at, updated_at
		FROM sync_agent_group
		WHERE id = ? AND deleted_at IS NULL
	`

	row := s.db.QueryRowContext(ctx, query, id)
	item, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GroupView{}, ErrGroupNotFound
		}
		return GroupView{}, err
	}
	return item, nil
}

func (s *AgentService) insertGroup(ctx context.Context, item GroupView) (GroupView, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return GroupView{}, err
	}
	defer tx.Rollback()

	workWindowsJSON, err := encodeWorkWindows(item.WorkWindows)
	if err != nil {
		return GroupView{}, err
	}
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO sync_agent_group (name, storage_id, ip_range, path_prefix, interval_time, del_time_days, max_workers, work_windows, file_filter_id, regex_id, image_process_id, alarm_group_id, log_enabled, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.Name,
		item.StorageID,
		item.IPRange,
		item.PathPrefix,
		item.IntervalTime,
		item.DelTimeDays,
		item.TransferSpeedLimit,
		workWindowsJSON,
		item.FileFilterID,
		item.RegexID,
		item.ImageProcessID,
		item.AlarmGroupID,
		item.LogEnabled,
		item.Status,
	)
	if err != nil {
		return GroupView{}, mapGroupWriteError(err)
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		return GroupView{}, err
	}
	if err := tx.Commit(); err != nil {
		return GroupView{}, err
	}
	return s.getGroupByIDDB(ctx, groupID)
}

func (s *AgentService) updateGroupRow(ctx context.Context, id int64, item GroupView) (GroupView, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return GroupView{}, err
	}
	defer tx.Rollback()

	workWindowsJSON, err := encodeWorkWindows(item.WorkWindows)
	if err != nil {
		return GroupView{}, err
	}
	result, err := tx.ExecContext(
		ctx,
		`UPDATE sync_agent_group SET name = ?, storage_id = ?, ip_range = ?, path_prefix = ?, interval_time = ?, del_time_days = ?, max_workers = ?, work_windows = ?, file_filter_id = ?, regex_id = ?, image_process_id = ?, alarm_group_id = ?, log_enabled = ?, status = ? WHERE id = ? AND deleted_at IS NULL`,
		item.Name,
		item.StorageID,
		item.IPRange,
		item.PathPrefix,
		item.IntervalTime,
		item.DelTimeDays,
		item.TransferSpeedLimit,
		workWindowsJSON,
		item.FileFilterID,
		item.RegexID,
		item.ImageProcessID,
		item.AlarmGroupID,
		item.LogEnabled,
		item.Status,
		id,
	)
	if err != nil {
		return GroupView{}, mapGroupWriteError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return GroupView{}, err
	}
	if affected == 0 {
		return GroupView{}, ErrGroupNotFound
	}
	if err := tx.Commit(); err != nil {
		return GroupView{}, err
	}
	return s.getGroupByIDDB(ctx, id)
}

func (s *AgentService) deleteGroupRow(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(
		ctx,
		`UPDATE sync_agent_group SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrGroupNotFound
	}
	return tx.Commit()
}

func (s *AgentService) seedGroupsIfEmpty(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM sync_agent_group WHERE deleted_at IS NULL`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	for _, item := range s.groups {
		if _, err := s.insertGroup(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func scanGroup(scanner interface {
	Scan(dest ...any) error
}) (GroupView, error) {
	var item GroupView
	var workWindowsRaw []byte
	var createdAt time.Time
	var updatedAt time.Time
	if err := scanner.Scan(
		&item.ID,
		&item.Name,
		&item.StorageID,
		&item.IPRange,
		&item.PathPrefix,
		&item.IntervalTime,
		&item.DelTimeDays,
		&item.TransferSpeedLimit,
		&workWindowsRaw,
		&item.FileFilterID,
		&item.RegexID,
		&item.ImageProcessID,
		&item.AlarmGroupID,
		&item.LogEnabled,
		&item.Status,
		&createdAt,
		&updatedAt,
	); err != nil {
		return GroupView{}, err
	}
	item.WorkWindows = decodeWorkWindows(workWindowsRaw)
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func mapGroupWriteError(err error) error {
	switch mysqlDuplicateField(err) {
	case "uk_sync_agent_group_name":
		return ErrGroupNameExists
	default:
		return err
	}
}

func (s *AgentService) listStorageTargetsFromDB(ctx context.Context) ([]StorageView, error) {
	const query = `
		SELECT id, name, type, endpoint, access_key, secret_key, bucket, region, local_path, quota_bytes, status, remark, created_at, updated_at
		FROM storage
		WHERE deleted_at IS NULL
		ORDER BY id ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StorageView, 0)
	for rows.Next() {
		item, scanErr := scanStorageTarget(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *AgentService) getStorageTargetByIDDB(ctx context.Context, id int64) (StorageView, error) {
	const query = `
		SELECT id, name, type, endpoint, access_key, secret_key, bucket, region, local_path, quota_bytes, status, remark, created_at, updated_at
		FROM storage
		WHERE id = ? AND deleted_at IS NULL
	`

	row := s.db.QueryRowContext(ctx, query, id)
	item, err := scanStorageTarget(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StorageView{}, ErrStorageNotFound
		}
		return StorageView{}, err
	}
	return item, nil
}

func (s *AgentService) insertStorageTarget(ctx context.Context, item StorageView) (StorageView, error) {
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO storage (name, type, endpoint, access_key, secret_key, bucket, region, local_path, quota_bytes, status, remark) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.Name,
		item.Type,
		item.Endpoint,
		item.AccessKey,
		item.SecretKey,
		item.Bucket,
		item.Region,
		item.LocalPath,
		item.Quota,
		item.Status,
		item.Remark,
	)
	if err != nil {
		return StorageView{}, mapStorageWriteError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return StorageView{}, err
	}
	return s.getStorageTargetByIDDB(ctx, id)
}

func (s *AgentService) updateStorageTargetRow(ctx context.Context, id int64, item StorageView) (StorageView, error) {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE storage SET name = ?, type = ?, endpoint = ?, access_key = ?, secret_key = ?, bucket = ?, region = ?, local_path = ?, quota_bytes = ?, status = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		item.Name,
		item.Type,
		item.Endpoint,
		item.AccessKey,
		item.SecretKey,
		item.Bucket,
		item.Region,
		item.LocalPath,
		item.Quota,
		item.Status,
		item.Remark,
		id,
	)
	if err != nil {
		return StorageView{}, mapStorageWriteError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return StorageView{}, err
	}
	if affected == 0 {
		return StorageView{}, ErrStorageNotFound
	}
	return s.getStorageTargetByIDDB(ctx, id)
}

func (s *AgentService) deleteStorageTargetRow(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE storage SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrStorageNotFound
	}
	return nil
}

func (s *AgentService) seedStorageTargetsIfEmpty(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM storage WHERE deleted_at IS NULL`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	for _, item := range s.storageTargets {
		if _, err := s.insertStorageTarget(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func scanStorageTarget(scanner interface {
	Scan(dest ...any) error
}) (StorageView, error) {
	var item StorageView
	var createdAt time.Time
	var updatedAt time.Time
	if err := scanner.Scan(
		&item.ID,
		&item.Name,
		&item.Type,
		&item.Endpoint,
		&item.AccessKey,
		&item.SecretKey,
		&item.Bucket,
		&item.Region,
		&item.LocalPath,
		&item.Quota,
		&item.Status,
		&item.Remark,
		&createdAt,
		&updatedAt,
	); err != nil {
		return StorageView{}, err
	}
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func mapStorageWriteError(err error) error {
	switch mysqlDuplicateField(err) {
	case "uk_storage_name":
		return ErrStorageNameExists
	default:
		return err
	}
}

func (s *AgentService) buildFileView(req FileMutation) (FileView, error) {
	if strings.TrimSpace(req.Name) == "" {
		return FileView{}, ErrFileNameRequired
	}
	if strings.TrimSpace(req.Path) == "" {
		return FileView{}, ErrFilePathRequired
	}
	storageItem := s.findStorage(req.StorageID)
	if storageItem == nil {
		return FileView{}, ErrFileStorageNotFound
	}
	modifiedAt := strings.TrimSpace(req.ModifiedAt)
	if modifiedAt == "" {
		modifiedAt = nowString()
	}
	return FileView{
		Name:       strings.TrimSpace(req.Name),
		Path:       strings.TrimSpace(req.Path),
		Type:       strings.TrimSpace(req.Type),
		Size:       firstNonEmpty(strings.TrimSpace(req.Size), "Unknown"),
		Tags:       normalizeStringList(req.Tags),
		ModifiedAt: modifiedAt,
		StorageID:  req.StorageID,
		Storage:    storageItem.Name,
	}, nil
}

func (s *AgentService) upsertAgent(req agent.HeartbeatRequest) error {
	now := nowString()
	for index := range s.agents {
		item := &s.agents[index]
		if item.HostSN == req.HostSN || (req.IP != "" && item.IP == req.IP) {
			item.IP = firstNonEmpty(req.IP, item.IP)
			item.HostSN = firstNonEmpty(req.HostSN, item.HostSN)
			item.Version = firstNonEmpty(req.Version, item.Version)
			item.Status = 1
			item.CPU = req.CPU
			item.Mem = req.Mem
			item.Storage = append([]agent.StorageMetric(nil), req.StorageInfo...)
			item.LastAccessTime = now
			if err := s.persistHeartbeatAgentLocked(item); err != nil {
				return err
			}
			return nil
		}
	}

	groupID, storageID, pathPrefix := s.matchGroupByIP(req.IP)
	hostSN := firstNonEmpty(req.HostSN, "unknown-host")
	item := AgentView{
		ID:             s.nextAgentID,
		HostSN:         hostSN,
		HostName:       hostSN,
		IP:             req.IP,
		GroupID:        groupID,
		StorageID:      storageID,
		PathPrefix:     pathPrefix,
		Version:        req.Version,
		Status:         1,
		Tags:           []string{"auto-registered"},
		LastAccessTime: now,
		LastCommitTime: "",
		Remark:         "Auto-registered from heartbeat",
		CPU:            req.CPU,
		Mem:            req.Mem,
		Storage:        append([]agent.StorageMetric(nil), req.StorageInfo...),
	}
	if err := s.persistHeartbeatAgentLocked(&item); err != nil {
		return err
	}
	s.agents = append(s.agents, item)
	if item.ID >= s.nextAgentID {
		s.nextAgentID = item.ID + 1
	} else {
		s.nextAgentID++
	}
	return nil
}

func (s *AgentService) persistHeartbeatAgentLocked(item *AgentView) error {
	if s.db == nil {
		return nil
	}

	ctx := context.Background()
	if existingID, err := s.findAgentIDByIdentityDB(ctx, item.HostSN, item.IP); err != nil {
		return err
	} else if existingID > 0 {
		persisted, err := s.updateAgentRow(ctx, existingID, AgentView{
			ID:             existingID,
			HostSN:         item.HostSN,
			HostName:       item.HostName,
			IP:             item.IP,
			GroupID:        item.GroupID,
			StorageID:      item.StorageID,
			SourcePaths:    item.SourcePaths,
			PathPrefix:     item.PathPrefix,
			Version:        item.Version,
			Status:         item.Status,
			Tags:           item.Tags,
			LastAccessTime: item.LastAccessTime,
			LastCommitTime: item.LastCommitTime,
			Remark:         item.Remark,
			CPU:            item.CPU,
			Mem:            item.Mem,
			Storage:        item.Storage,
		})
		if err != nil {
			return err
		}
		*item = persisted
		return nil
	}

	persisted, err := s.insertAgent(ctx, *item)
	if err != nil {
		return err
	}
	*item = persisted
	return nil
}

func (s *AgentService) findAgentIDByIdentityDB(ctx context.Context, hostSN, ip string) (int64, error) {
	hostSN = strings.TrimSpace(hostSN)
	ip = strings.TrimSpace(ip)

	var row *sql.Row
	switch {
	case hostSN != "" && ip != "":
		row = s.db.QueryRowContext(ctx, `SELECT id FROM sync_agent WHERE deleted_at IS NULL AND (host_sn = ? OR ip = ?) ORDER BY id ASC LIMIT 1`, hostSN, ip)
	case hostSN != "":
		row = s.db.QueryRowContext(ctx, `SELECT id FROM sync_agent WHERE deleted_at IS NULL AND host_sn = ? ORDER BY id ASC LIMIT 1`, hostSN)
	case ip != "":
		row = s.db.QueryRowContext(ctx, `SELECT id FROM sync_agent WHERE deleted_at IS NULL AND ip = ? ORDER BY id ASC LIMIT 1`, ip)
	default:
		return 0, nil
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}

func (s *AgentService) matchGroupByIP(ip string) (int64, int64, string) {
	for _, group := range s.groups {
		prefix := strings.TrimSuffix(group.IPRange, "0/24")
		if prefix != "" && strings.HasPrefix(ip, prefix) {
			return group.ID, group.StorageID, group.PathPrefix
		}
	}
	return 0, 0, ""
}

func (s *AgentService) buildHeartbeatPolicyLocked(req agent.HeartbeatRequest) agent.HeartbeatPolicy {
	policy := s.defaultPolicy

	agentItem := s.findAgentForHeartbeatLocked(req)
	group := s.findGroupForHeartbeatLocked(req)
	if group == nil {
		return policy
	}

	storageID := group.StorageID
	if agentItem != nil && agentItem.StorageID > 0 {
		storageID = agentItem.StorageID
	}
	storageItem := s.findStorage(storageID)
	if storageItem != nil {
		policy.Storage = agent.StorageConfig{
			Type:      string(storageItem.Type),
			Name:      storageItem.Name,
			Endpoint:  storageItem.Endpoint,
			AK:        storageItem.AccessKey,
			SK:        storageItem.SecretKey,
			Bucket:    storageItem.Bucket,
			LocalPath: storageItem.LocalPath,
		}
	}
	sourcePaths := []string{}
	if agentItem != nil && len(agentItem.SourcePaths) > 0 {
		sourcePaths = normalizeStringList(agentItem.SourcePaths)
	}
	if len(sourcePaths) == 0 && strings.TrimSpace(group.PathPrefix) != "" {
		sourcePaths = []string{strings.TrimSpace(group.PathPrefix)}
	}
	if len(sourcePaths) > 0 {
		policy.Paths = make([]agent.TaskPath, 0, len(sourcePaths))
		for _, sourcePath := range sourcePaths {
			policy.Paths = append(policy.Paths, agent.TaskPath{Path: sourcePath, Type: agent.PathTypeSync})
		}
	}
	policy.DelTime = group.DelTimeDays
	policy.RunTime = group.IntervalTime
	policy.PathPrefix = group.PathPrefix
	if agentItem != nil && strings.TrimSpace(agentItem.PathPrefix) != "" {
		policy.PathPrefix = strings.TrimSpace(agentItem.PathPrefix)
	}
	policy.Status = group.Status
	if agentItem != nil {
		policy.Status = agentItem.Status
	}
	if group.TransferSpeedLimit > 0 {
		policy.MaxWorkers = group.TransferSpeedLimit
	}
	if len(group.WorkWindows) > 0 {
		policy.WorkStartTime = group.WorkWindows[0].StartTime
		policy.WorkEndTime = group.WorkWindows[0].EndTime
	}

	if filter, ok := s.lookupFileFilterLocked(group.FileFilterID); ok {
		policy.Filters = filter
	}
	if regexValue, tagMappings, tagsAsPath, fetchField, ok := s.lookupRegexLocked(group.RegexID); ok {
		policy.FetchRegex = regexValue
		policy.TagsKeyList = tagMappings
		policy.TagsAsPath = tagsAsPath
		policy.FetchField = fetchField
	}

	return policy
}

func (s *AgentService) findAgentForHeartbeatLocked(req agent.HeartbeatRequest) *AgentView {
	for index := range s.agents {
		item := &s.agents[index]
		if item.HostSN == req.HostSN || (req.IP != "" && item.IP == req.IP) {
			return item
		}
	}
	return nil
}

func (s *AgentService) findGroupForHeartbeatLocked(req agent.HeartbeatRequest) *GroupView {
	if item := s.findAgentForHeartbeatLocked(req); item != nil {
		return s.findGroup(item.GroupID)
	}
	groupID, _, _ := s.matchGroupByIP(req.IP)
	return s.findGroup(groupID)
}

func (s *AgentService) lookupFileFilterLocked(id int64) (agent.SyncFilter, bool) {
	if id <= 0 || s.db == nil {
		return agent.SyncFilter{}, false
	}
	var filter agent.SyncFilter
	var patternsRaw []byte
	if err := s.db.QueryRowContext(context.Background(), `SELECT filter_scope, list_type, patterns FROM file_filter WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&filter.Type, &filter.ListType, &patternsRaw); err != nil {
		return agent.SyncFilter{}, false
	}
	_ = json.Unmarshal(patternsRaw, &filter.Filter)
	filter.Filter = normalizeStringList(filter.Filter)
	return filter, true
}

func (s *AgentService) lookupRegexLocked(id int64) (string, []agent.TagMap, int, int, bool) {
	if id <= 0 || s.db == nil {
		return "", nil, 0, 0, false
	}
	var sourceField string
	var regexpValue string
	var asPath int
	if err := s.db.QueryRowContext(context.Background(), `SELECT source_field, regexp, as_path FROM sync_regex WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&sourceField, &regexpValue, &asPath); err != nil {
		return "", nil, 0, 0, false
	}
	rows, err := s.db.QueryContext(context.Background(), `SELECT capture_index, tag_key FROM sync_regex_mapping WHERE regex_id = ? ORDER BY capture_index ASC, tag_key ASC`, id)
	if err != nil {
		return regexpValue, nil, asPath, fetchFieldFromSource(sourceField), true
	}
	defer rows.Close()
	items := []agent.TagMap{}
	for rows.Next() {
		var item agent.TagMap
		if err := rows.Scan(&item.Index, &item.TagKey); err != nil {
			continue
		}
		item.TagKey = strings.TrimSpace(item.TagKey)
		if item.TagKey == "" {
			continue
		}
		items = append(items, item)
	}
	return regexpValue, items, asPath, fetchFieldFromSource(sourceField), true
}

func fetchFieldFromSource(source string) int {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "filename":
		return 1
	default:
		return 0
	}
}

func (s *AgentService) recordSyncLog(ctx context.Context, values map[string]string) error {
	fileCount, _ := strconv.ParseInt(values["fileCount"], 10, 64)
	fileSizeBytes, _ := strconv.ParseInt(values["fileSize"], 10, 64)
	if fileCount == 0 && fileSizeBytes == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowString()
	errorCount, _ := strconv.ParseInt(values["errCount"], 10, 64)
	agentID, hostName := s.lookupAgentIdentityByIP(values["ip"])
	status := "success"
	message := ""
	if errorCount > 0 {
		status = "partial_success"
		message = fmt.Sprintf("%d files failed during sync", errorCount)
	}

	entryID := "log-" + strconv.FormatInt(s.nextLogID, 10)
	if s.db != nil {
		result, err := s.db.ExecContext(
			ctx,
			`INSERT INTO sync_agent_logs (agent_id, group_id, task_type, source_path, task_start_time, file_count, file_size_bytes, error_count, tar_list_path, status, message) VALUES (?, ?, 'sync', ?, ?, ?, ?, ?, ?, ?, ?)`,
			agentID,
			s.lookupGroupIDByAgentID(agentID),
			firstNonEmpty(values["path"], values["tarListPath"]),
			nullTimeValue(values["taskStartTime"]),
			fileCount,
			fileSizeBytes,
			errorCount,
			values["tarListPath"],
			status,
			message,
		)
		if err != nil {
			return err
		}
		if dbID, err := result.LastInsertId(); err == nil && dbID > 0 {
			entryID = "log-" + strconv.FormatInt(dbID, 10)
		}
	}

	entry := SyncLogView{
		ID:         entryID,
		AgentIP:    values["ip"],
		HostName:   hostName,
		Path:       firstNonEmpty(values["path"], values["tarListPath"]),
		StartTime:  firstNonEmpty(values["taskStartTime"], now),
		FileCount:  fileCount,
		FileSize:   humanizeBytes(fileSizeBytes),
		ErrorCount: errorCount,
		LogPath:    values["tarListPath"],
		CommitTime: now,
	}
	s.nextLogID++
	s.syncLogs = append([]SyncLogView{entry}, s.syncLogs...)

	for index := range s.agents {
		item := &s.agents[index]
		if item.IP == values["ip"] {
			item.LastCommitTime = now
			if item.Status == 0 {
				item.Status = 1
			}
			return nil
		}
	}

	return nil
}

func (s *AgentService) listSyncLogsFromDB(ctx context.Context, query SyncLogQuery) ([]SyncLogView, error) {
	sqlText := `SELECT l.id, COALESCE(a.ip, ''), COALESCE(a.host_name, ''), l.source_path, l.task_start_time, l.file_count, l.file_size_bytes, l.error_count, l.tar_list_path, l.created_at
		FROM sync_agent_logs l
		LEFT JOIN sync_agent a ON a.id = l.agent_id
		WHERE l.task_type = 'sync'`
	args := make([]any, 0, 1)
	switch normalizeSyncLogResult(query.Result) {
	case "failed":
		sqlText += ` AND l.error_count > 0`
	case "success":
		sqlText += ` AND l.error_count = 0`
	}
	sqlText += ` ORDER BY l.created_at DESC, l.id DESC`

	rows, err := s.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SyncLogView, 0)
	for rows.Next() {
		var (
			numericID     int64
			agentIP       string
			hostName      string
			sourcePath    string
			taskStartTime sql.NullTime
			fileCount     int64
			fileSizeBytes int64
			errorCount    int64
			tarListPath   string
			createdAt     time.Time
		)
		if err := rows.Scan(&numericID, &agentIP, &hostName, &sourcePath, &taskStartTime, &fileCount, &fileSizeBytes, &errorCount, &tarListPath, &createdAt); err != nil {
			return nil, err
		}

		startTime := ""
		if taskStartTime.Valid {
			startTime = taskStartTime.Time.Format(mysqlTimeFormat)
		}

		items = append(items, SyncLogView{
			ID:         "log-" + strconv.FormatInt(numericID, 10),
			AgentIP:    agentIP,
			HostName:   hostName,
			Path:       sourcePath,
			StartTime:  startTime,
			FileCount:  fileCount,
			FileSize:   humanizeBytes(fileSizeBytes),
			ErrorCount: errorCount,
			LogPath:    tarListPath,
			CommitTime: createdAt.Format(mysqlTimeFormat),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func normalizeSyncLogResult(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "failed":
		return "failed"
	case "success":
		return "success"
	default:
		return "all"
	}
}

func filterSyncLogs(items []SyncLogView, query SyncLogQuery) []SyncLogView {
	result := normalizeSyncLogResult(query.Result)
	if result == "all" {
		return items
	}

	filtered := make([]SyncLogView, 0, len(items))
	for _, item := range items {
		if result == "failed" && item.ErrorCount > 0 {
			filtered = append(filtered, item)
			continue
		}
		if result == "success" && item.ErrorCount == 0 {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (s *AgentService) lookupAgentIdentityByIP(ip string) (int64, string) {
	for _, item := range s.agents {
		if item.IP == ip {
			return item.ID, item.HostName
		}
	}
	return 0, ""
}

func (s *AgentService) lookupGroupIDByAgentID(agentID int64) int64 {
	if agentID == 0 {
		return 0
	}
	for _, item := range s.agents {
		if item.ID == agentID {
			return item.GroupID
		}
	}
	return 0
}

func (s *AgentService) recordScanReport(payload []byte) {
	type scanCommitPayload struct {
		AgentName      string              `json:"agentName"`
		HostName       string              `json:"hostName"`
		IP             string              `json:"ip"`
		GroupName      string              `json:"groupName"`
		RootPath       string              `json:"rootPath"`
		Path           string              `json:"path"`
		FileCount      int64               `json:"fileCount"`
		TotalSize      string              `json:"totalSize"`
		TypeBreakdown  []ScanBreakdown     `json:"typeBreakdown"`
		DirectoryStats []ScanDirectoryStat `json:"directoryStats"`
	}

	var body scanCommitPayload
	if err := json.Unmarshal(payload, &body); err != nil {
		return
	}

	rootPath := firstNonEmpty(body.RootPath, body.Path)
	if rootPath == "" {
		return
	}

	report := ScanReportView{
		ID:             "scan-" + strconv.FormatInt(s.nextScanID, 10),
		AgentName:      firstNonEmpty(body.AgentName, body.HostName, body.IP, "unknown-agent"),
		GroupName:      firstNonEmpty(body.GroupName, s.lookupGroupNameByIP(body.IP), "Unknown"),
		RootPath:       rootPath,
		FileCount:      body.FileCount,
		TotalSize:      firstNonEmpty(body.TotalSize, "Unknown"),
		FinishedAt:     nowString(),
		TypeBreakdown:  append([]ScanBreakdown(nil), body.TypeBreakdown...),
		DirectoryStats: append([]ScanDirectoryStat(nil), body.DirectoryStats...),
	}
	s.nextScanID++
	s.scanReports = append([]ScanReportView{report}, s.scanReports...)
}

func (s *AgentService) lookupGroupNameByIP(ip string) string {
	groupID, _, _ := s.matchGroupByIP(ip)
	for _, group := range s.groups {
		if group.ID == groupID {
			return group.Name
		}
	}
	return ""
}

func (s *AgentService) syncAgentsForGroupLocked(group GroupView) {
	for index := range s.agents {
		if s.agents[index].GroupID == group.ID {
			s.agents[index].StorageID = group.StorageID
			if strings.TrimSpace(s.agents[index].PathPrefix) == "" || s.agents[index].PathPrefix == group.PathPrefix {
				s.agents[index].PathPrefix = group.PathPrefix
			}
		}
	}
}

func (s *AgentService) syncStorageReferencesLocked(target StorageView) {
	for index := range s.files {
		if s.files[index].StorageID == target.ID {
			s.files[index].Storage = target.Name
		}
	}
}

func (s *AgentService) indexAgent(id int64) int {
	for index, item := range s.agents {
		if item.ID == id {
			return index
		}
	}
	return -1
}

func (s *AgentService) indexGroup(id int64) int {
	for index, item := range s.groups {
		if item.ID == id {
			return index
		}
	}
	return -1
}

func (s *AgentService) indexStorage(id int64) int {
	for index, item := range s.storageTargets {
		if item.ID == id {
			return index
		}
	}
	return -1
}

func (s *AgentService) indexFile(id string) int {
	for index, item := range s.files {
		if item.ID == id {
			return index
		}
	}
	return -1
}

func (s *AgentService) findAgent(id int64) *AgentView {
	index := s.indexAgent(id)
	if index < 0 {
		return nil
	}
	return &s.agents[index]
}

func (s *AgentService) findGroup(id int64) *GroupView {
	index := s.indexGroup(id)
	if index < 0 {
		return nil
	}
	return &s.groups[index]
}

func (s *AgentService) findStorage(id int64) *StorageView {
	index := s.indexStorage(id)
	if index < 0 {
		return nil
	}
	return &s.storageTargets[index]
}

func cloneAgent(item AgentView) AgentView {
	item.Tags = append([]string(nil), item.Tags...)
	item.SourcePaths = append([]string(nil), item.SourcePaths...)
	item.Storage = cloneStorageMetrics(item.Storage)
	return item
}

func cloneGroup(item GroupView) GroupView {
	item.WorkWindows = append([]WorkWindowView(nil), item.WorkWindows...)
	return item
}

func cloneFile(item FileView) FileView {
	item.Tags = append([]string(nil), item.Tags...)
	return item
}

func cloneStorageMetrics(items []agent.StorageMetric) []agent.StorageMetric {
	return append([]agent.StorageMetric(nil), items...)
}

func normalizeStringList(values []string) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func normalizeWorkWindows(values []WorkWindowView) []WorkWindowView {
	items := make([]WorkWindowView, 0, len(values))
	for _, value := range values {
		startTime := strings.TrimSpace(value.StartTime)
		endTime := strings.TrimSpace(value.EndTime)
		if startTime == "" || endTime == "" {
			continue
		}
		items = append(items, WorkWindowView{StartTime: startTime, EndTime: endTime})
	}
	return items
}

func encodeWorkWindows(values []WorkWindowView) ([]byte, error) {
	items := normalizeWorkWindows(values)
	if len(items) == 0 {
		return json.Marshal([]WorkWindowView{})
	}
	return json.Marshal(items)
}

func decodeWorkWindows(workWindowsRaw []byte) []WorkWindowView {
	if len(workWindowsRaw) == 0 {
		return nil
	}
	var items []WorkWindowView
	if err := json.Unmarshal(workWindowsRaw, &items); err != nil {
		return nil
	}
	return normalizeWorkWindows(items)
}

func nowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func nullTimeValue(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if unixMillis, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		switch len(trimmed) {
		case 13:
			return time.UnixMilli(unixMillis)
		case 10:
			return time.Unix(unixMillis, 0)
		}
	}
	parsed, err := time.ParseInLocation(mysqlTimeFormat, trimmed, time.Local)
	if err != nil {
		return trimmed
	}
	return parsed
}

func ParseStringID(raw, label string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("%s id is required", label)
	}
	return raw, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func compareVersion(left, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	maxLen := len(leftParts)
	if len(rightParts) > maxLen {
		maxLen = len(rightParts)
	}

	for i := 0; i < maxLen; i++ {
		lv := versionPart(leftParts, i)
		rv := versionPart(rightParts, i)
		if lv > rv {
			return 1
		}
		if lv < rv {
			return -1
		}
	}
	return 0
}

func sanitizeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "..", "")
	return name
}

func storageUploadDir(target StorageView) string {
	if strings.TrimSpace(target.LocalPath) != "" {
		return target.LocalPath
	}
	return filepath.Join(os.TempDir(), "visionvault", fmt.Sprintf("storage-%d", target.ID))
}

func versionUploadDir(currentDownloadFile string) string {
	if strings.TrimSpace(currentDownloadFile) != "" {
		return filepath.Dir(currentDownloadFile)
	}
	return filepath.Join(os.TempDir(), "visionvault", "agent-packages")
}

func humanizeBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(size)
	for _, unit := range units {
		value /= 1024
		if value < 1024 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", value, unit)
		}
	}
	return fmt.Sprintf("%d B", size)
}

func versionIDFromVersion(version string) string {
	replacer := strings.NewReplacer(".", "-", " ", "-", "/", "-")
	return "pkg-" + replacer.Replace(strings.TrimSpace(version))
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func versionPart(parts []string, index int) int {
	if index >= len(parts) {
		return 0
	}
	value, err := strconv.Atoi(parts[index])
	if err != nil {
		return 0
	}
	return value
}

func (s *AgentService) listAgentVersionSummariesFromDB(ctx context.Context) ([]VersionAgentSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT version, COUNT(1) AS agent_count
		FROM sync_agent
		WHERE deleted_at IS NULL AND TRIM(version) <> ''
		GROUP BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []VersionAgentSummary{}
	for rows.Next() {
		var item VersionAgentSummary
		if err := rows.Scan(&item.Version, &item.AgentCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortVersionSummaries(items)
	return items, nil
}

func summarizeAgentVersions(agents []AgentView) []VersionAgentSummary {
	counts := make(map[string]int64)
	for _, item := range agents {
		version := strings.TrimSpace(item.Version)
		if version == "" {
			continue
		}
		counts[version]++
	}
	items := make([]VersionAgentSummary, 0, len(counts))
	for version, count := range counts {
		items = append(items, VersionAgentSummary{
			Version:    version,
			AgentCount: count,
		})
	}
	sortVersionSummaries(items)
	return items
}

func sortVersionSummaries(items []VersionAgentSummary) {
	sort.Slice(items, func(i, j int) bool {
		return compareVersion(items[i].Version, items[j].Version) > 0
	})
	if len(items) > 0 {
		items[0].IsLatest = true
	}
}

func countOnlineAgents(_ []VersionView, agents []AgentView) int64 {
	var count int64
	for _, item := range agents {
		if item.Status == 1 {
			count++
		}
	}
	return count
}

func currentPackageAgents(items []VersionView) int64 {
	if len(items) == 0 {
		return 0
	}
	return items[0].AgentCount
}

func currentPackageAgentsFromSummary(items []VersionAgentSummary) int64 {
	if len(items) == 0 {
		return 0
	}
	return items[0].AgentCount
}

func fileMD5(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
