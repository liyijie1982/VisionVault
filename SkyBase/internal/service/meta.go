package service

import (
	"time"

	"skybase/internal/config"
)

type MetaService struct {
	cfg       config.Config
	startedAt time.Time
	modules   []ModuleSummary
}

type ModuleSummary struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type HealthStatus struct {
	Name      string          `json:"name"`
	Env       string          `json:"env"`
	Status    string          `json:"status"`
	StartedAt time.Time       `json:"startedAt"`
	Now       time.Time       `json:"now"`
	Modules   []ModuleSummary `json:"modules"`
}

func NewMetaService(cfg config.Config) *MetaService {
	return &MetaService{
		cfg:       cfg,
		startedAt: time.Now(),
		modules: []ModuleSummary{
			{Key: "auth", Name: "RBAC", Description: "用户、部门、角色、菜单、登录日志与系统设置", Status: "live"},
			{Key: "storage", Name: "Storage", Description: "本地与 S3 兼容存储、文件浏览与版本包管理", Status: "live"},
			{Key: "agent", Name: "Agent", Description: "Agent 管理、心跳、同步日志、扫描报告与升级分发", Status: "live"},
			{Key: "monitor", Name: "Monitor", Description: "监控面板、告警策略、告警组与通知通道", Status: "live"},
			{Key: "jobs", Name: "Jobs", Description: "文件权限、任务进度、规则配置与运维辅助模块", Status: "partial"},
			{Key: "backup", Name: "Backup", Description: "磁带设备、备份任务与备份日志界面仍为演示态", Status: "demo"},
		},
	}
}

func (s *MetaService) Health() HealthStatus {
	return HealthStatus{
		Name:      s.cfg.App.Name,
		Env:       s.cfg.App.Env,
		Status:    "ok",
		StartedAt: s.startedAt,
		Now:       time.Now(),
		Modules:   s.modules,
	}
}
