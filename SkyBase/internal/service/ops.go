package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type OpsService struct {
	db *sql.DB
}

func NewOpsService(db *sql.DB) *OpsService {
	return &OpsService{db: db}
}

type FileFilterRecord struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	FilterScope string   `json:"filterScope"`
	ListType    string   `json:"listType"`
	Patterns    []string `json:"patterns"`
	Status      int      `json:"status"`
	Remark      string   `json:"remark"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

type FileFilterMutation struct {
	Name        string   `json:"name"`
	FilterScope string   `json:"filterScope"`
	ListType    string   `json:"listType"`
	Patterns    []string `json:"patterns"`
	Status      int      `json:"status"`
	Remark      string   `json:"remark"`
}

type RegexTagMapping struct {
	CaptureIndex int    `json:"captureIndex"`
	TagKey       string `json:"tagKey"`
}

type RegexRuleRecord struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	SourceField string            `json:"sourceField"`
	Regexp      string            `json:"regexp"`
	AsPath      int               `json:"asPath"`
	Status      int               `json:"status"`
	Mappings    []RegexTagMapping `json:"mappings"`
	Remark      string            `json:"remark"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

type RegexRuleMutation struct {
	Name        string            `json:"name"`
	SourceField string            `json:"sourceField"`
	Regexp      string            `json:"regexp"`
	AsPath      int               `json:"asPath"`
	Status      int               `json:"status"`
	Mappings    []RegexTagMapping `json:"mappings"`
	Remark      string            `json:"remark"`
}

type ImageProcessorRecord struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	ProcessorType string `json:"processorType"`
	ConfigJSON    string `json:"configJson"`
	Status        int    `json:"status"`
	Remark        string `json:"remark"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type ImageProcessorMutation struct {
	Name          string `json:"name"`
	ProcessorType string `json:"processorType"`
	ConfigJSON    string `json:"configJson"`
	Status        int    `json:"status"`
	Remark        string `json:"remark"`
}

type AlertGroupRecord struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Receivers []string `json:"receivers"`
	Status    int      `json:"status"`
	Remark    string   `json:"remark"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

type AlertGroupMutation struct {
	Name      string   `json:"name"`
	Receivers []string `json:"receivers"`
	Status    int      `json:"status"`
	Remark    string   `json:"remark"`
}

type MessageChannelRecord struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ChannelType string `json:"channelType"`
	ConfigJSON  string `json:"configJson"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type MessageChannelMutation struct {
	Name        string `json:"name"`
	ChannelType string `json:"channelType"`
	ConfigJSON  string `json:"configJson"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
}

type AlertPolicyRecord struct {
	ID                      int64   `json:"id"`
	Name                    string  `json:"name"`
	CPUThreshold            float64 `json:"cpuThreshold"`
	MemThreshold            float64 `json:"memThreshold"`
	DiskThreshold           float64 `json:"diskThreshold"`
	CPUConsecutiveTimes     int     `json:"cpuConsecutiveTimes"`
	MemConsecutiveTimes     int     `json:"memConsecutiveTimes"`
	HeartbeatTimeoutSeconds int     `json:"heartbeatTimeoutSeconds"`
	SendFrequencySeconds    int     `json:"sendFrequencySeconds"`
	Status                  int     `json:"status"`
	Remark                  string  `json:"remark"`
	CreatedAt               string  `json:"createdAt"`
	UpdatedAt               string  `json:"updatedAt"`
}

type AlertPolicyMutation struct {
	Name                    string  `json:"name"`
	CPUThreshold            float64 `json:"cpuThreshold"`
	MemThreshold            float64 `json:"memThreshold"`
	DiskThreshold           float64 `json:"diskThreshold"`
	CPUConsecutiveTimes     int     `json:"cpuConsecutiveTimes"`
	MemConsecutiveTimes     int     `json:"memConsecutiveTimes"`
	HeartbeatTimeoutSeconds int     `json:"heartbeatTimeoutSeconds"`
	SendFrequencySeconds    int     `json:"sendFrequencySeconds"`
	Status                  int     `json:"status"`
	Remark                  string  `json:"remark"`
}

type FileLogRecord struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"userId"`
	Username      string `json:"username"`
	StorageID     int64  `json:"storageId"`
	StorageName   string `json:"storageName"`
	FilePath      string `json:"filePath"`
	OperationType string `json:"operationType"`
	ResultStatus  string `json:"resultStatus"`
	ClientIP      string `json:"clientIp"`
	Message       string `json:"message"`
	CreatedAt     string `json:"createdAt"`
}

type TaskProgressRecord struct {
	ID           int64  `json:"id"`
	TaskType     string `json:"taskType"`
	Status       string `json:"status"`
	TotalCount   int64  `json:"totalCount"`
	SuccessCount int64  `json:"successCount"`
	FailedCount  int64  `json:"failedCount"`
	PayloadJSON  string `json:"payloadJson"`
	ResultJSON   string `json:"resultJson"`
	StartedAt    string `json:"startedAt"`
	FinishedAt   string `json:"finishedAt"`
	CreatedAt    string `json:"createdAt"`
	Remark       string `json:"remark"`
}

type TaskProgressMutation struct {
	TaskType     string `json:"taskType"`
	Status       string `json:"status"`
	TotalCount   int64  `json:"totalCount"`
	SuccessCount int64  `json:"successCount"`
	FailedCount  int64  `json:"failedCount"`
	PayloadJSON  string `json:"payloadJson"`
	ResultJSON   string `json:"resultJson"`
	StartedAt    string `json:"startedAt"`
	FinishedAt   string `json:"finishedAt"`
	Remark       string `json:"remark"`
}

type AlertLogRecord struct {
	ID               int64  `json:"id"`
	AgentID          int64  `json:"agentId"`
	GroupID          int64  `json:"groupId"`
	RuleID           int64  `json:"ruleId"`
	MessageChannelID int64  `json:"messageChannelId"`
	AlertLevel       string `json:"alertLevel"`
	AlertTitle       string `json:"alertTitle"`
	AlertBody        string `json:"alertBody"`
	SendStatus       string `json:"sendStatus"`
	FailureReason    string `json:"failureReason"`
	CreatedAt        string `json:"createdAt"`
}

type SystemConfigRecord struct {
	ID          int64  `json:"id"`
	ConfigGroup string `json:"configGroup"`
	ConfigKey   string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	ValueType   string `json:"valueType"`
	IsEncrypted int    `json:"isEncrypted"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type SystemConfigMutation struct {
	ConfigGroup string `json:"configGroup"`
	ConfigKey   string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	ValueType   string `json:"valueType"`
	IsEncrypted int    `json:"isEncrypted"`
	Status      int    `json:"status"`
	Remark      string `json:"remark"`
}

type FilePermissionRecord struct {
	ID                  int64  `json:"id"`
	UserID              int64  `json:"userId"`
	Username            string `json:"username"`
	StorageID           int64  `json:"storageId"`
	StorageName         string `json:"storageName"`
	CanView             int    `json:"canView"`
	CanUpload           int    `json:"canUpload"`
	CanDownload         int    `json:"canDownload"`
	CanDelete           int    `json:"canDelete"`
	CanBatchDownload    int    `json:"canBatchDownload"`
	CanDownloadToServer int    `json:"canDownloadToServer"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
}

type FilePermissionMutation struct {
	UserID              int64 `json:"userId"`
	StorageID           int64 `json:"storageId"`
	CanView             int   `json:"canView"`
	CanUpload           int   `json:"canUpload"`
	CanDownload         int   `json:"canDownload"`
	CanDelete           int   `json:"canDelete"`
	CanBatchDownload    int   `json:"canBatchDownload"`
	CanDownloadToServer int   `json:"canDownloadToServer"`
}

type LicenseRecord struct {
	ID            int64  `json:"id"`
	LicenseCode   string `json:"licenseCode"`
	SerialNumber  string `json:"serialNumber"`
	MaxAgentCount int    `json:"maxAgentCount"`
	IssuedAt      string `json:"issuedAt"`
	ExpiredAt     string `json:"expiredAt"`
	TrialDays     int    `json:"trialDays"`
	Status        int    `json:"status"`
	Remark        string `json:"remark"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type LicenseMutation struct {
	LicenseCode   string `json:"licenseCode"`
	SerialNumber  string `json:"serialNumber"`
	MaxAgentCount int    `json:"maxAgentCount"`
	IssuedAt      string `json:"issuedAt"`
	ExpiredAt     string `json:"expiredAt"`
	TrialDays     int    `json:"trialDays"`
	Status        int    `json:"status"`
	Remark        string `json:"remark"`
}

func (s *OpsService) ListFileFilters(ctx context.Context) ([]FileFilterRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, filter_scope, list_type, patterns, status, remark, created_at, updated_at FROM file_filter WHERE deleted_at IS NULL ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []FileFilterRecord{}
	for rows.Next() {
		var item FileFilterRecord
		var patternsRaw []byte
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.FilterScope, &item.ListType, &patternsRaw, &item.Status, &item.Remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.Patterns = decodeStringArray(patternsRaw)
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateFileFilter(ctx context.Context, req FileFilterMutation) (FileFilterRecord, error) {
	payload, err := normalizeFileFilterMutation(req)
	if err != nil {
		return FileFilterRecord{}, err
	}
	patternsRaw, _ := json.Marshal(payload.Patterns)
	result, err := s.db.ExecContext(ctx, `INSERT INTO file_filter (name, filter_scope, list_type, patterns, status, remark) VALUES (?, ?, ?, ?, ?, ?)`,
		payload.Name, payload.FilterScope, payload.ListType, patternsRaw, payload.Status, payload.Remark,
	)
	if err != nil {
		return FileFilterRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.getFileFilter(ctx, id)
}

func (s *OpsService) UpdateFileFilter(ctx context.Context, id int64, req FileFilterMutation) (FileFilterRecord, error) {
	payload, err := normalizeFileFilterMutation(req)
	if err != nil {
		return FileFilterRecord{}, err
	}
	patternsRaw, _ := json.Marshal(payload.Patterns)
	if _, err := s.db.ExecContext(ctx, `UPDATE file_filter SET name = ?, filter_scope = ?, list_type = ?, patterns = ?, status = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		payload.Name, payload.FilterScope, payload.ListType, patternsRaw, payload.Status, payload.Remark, id,
	); err != nil {
		return FileFilterRecord{}, err
	}
	return s.getFileFilter(ctx, id)
}

func (s *OpsService) DeleteFileFilter(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE file_filter SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, id)
	return err
}

func (s *OpsService) getFileFilter(ctx context.Context, id int64) (FileFilterRecord, error) {
	var item FileFilterRecord
	var patternsRaw []byte
	var createdAt time.Time
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id, name, filter_scope, list_type, patterns, status, remark, created_at, updated_at FROM file_filter WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&item.ID, &item.Name, &item.FilterScope, &item.ListType, &patternsRaw, &item.Status, &item.Remark, &createdAt, &updatedAt)
	if err != nil {
		return FileFilterRecord{}, err
	}
	item.Patterns = decodeStringArray(patternsRaw)
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func (s *OpsService) ListRegexRules(ctx context.Context) ([]RegexRuleRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, source_field, regexp, as_path, status, remark, created_at, updated_at FROM sync_regex WHERE deleted_at IS NULL ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []RegexRuleRecord{}
	for rows.Next() {
		var item RegexRuleRecord
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.SourceField, &item.Regexp, &item.AsPath, &item.Status, &item.Remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		mappings, err := s.listRegexMappings(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		item.Mappings = mappings
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateRegexRule(ctx context.Context, req RegexRuleMutation) (RegexRuleRecord, error) {
	payload, err := normalizeRegexRuleMutation(req)
	if err != nil {
		return RegexRuleRecord{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RegexRuleRecord{}, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `INSERT INTO sync_regex (name, source_field, regexp, as_path, status, remark) VALUES (?, ?, ?, ?, ?, ?)`,
		payload.Name, payload.SourceField, payload.Regexp, payload.AsPath, payload.Status, payload.Remark,
	)
	if err != nil {
		return RegexRuleRecord{}, err
	}
	id, _ := result.LastInsertId()
	if err := s.replaceRegexMappingsTx(ctx, tx, id, payload.Mappings); err != nil {
		return RegexRuleRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return RegexRuleRecord{}, err
	}
	return s.getRegexRule(ctx, id)
}

func (s *OpsService) UpdateRegexRule(ctx context.Context, id int64, req RegexRuleMutation) (RegexRuleRecord, error) {
	payload, err := normalizeRegexRuleMutation(req)
	if err != nil {
		return RegexRuleRecord{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RegexRuleRecord{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE sync_regex SET name = ?, source_field = ?, regexp = ?, as_path = ?, status = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		payload.Name, payload.SourceField, payload.Regexp, payload.AsPath, payload.Status, payload.Remark, id,
	); err != nil {
		return RegexRuleRecord{}, err
	}
	if err := s.replaceRegexMappingsTx(ctx, tx, id, payload.Mappings); err != nil {
		return RegexRuleRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return RegexRuleRecord{}, err
	}
	return s.getRegexRule(ctx, id)
}

func (s *OpsService) DeleteRegexRule(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM sync_regex_mapping WHERE regex_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sync_regex SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OpsService) getRegexRule(ctx context.Context, id int64) (RegexRuleRecord, error) {
	var item RegexRuleRecord
	var createdAt time.Time
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id, name, source_field, regexp, as_path, status, remark, created_at, updated_at FROM sync_regex WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&item.ID, &item.Name, &item.SourceField, &item.Regexp, &item.AsPath, &item.Status, &item.Remark, &createdAt, &updatedAt)
	if err != nil {
		return RegexRuleRecord{}, err
	}
	mappings, err := s.listRegexMappings(ctx, id)
	if err != nil {
		return RegexRuleRecord{}, err
	}
	item.Mappings = mappings
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func (s *OpsService) listRegexMappings(ctx context.Context, regexID int64) ([]RegexTagMapping, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT capture_index, tag_key FROM sync_regex_mapping WHERE regex_id = ? ORDER BY capture_index ASC, tag_key ASC`, regexID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []RegexTagMapping{}
	for rows.Next() {
		var item RegexTagMapping
		if err := rows.Scan(&item.CaptureIndex, &item.TagKey); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) replaceRegexMappingsTx(ctx context.Context, tx *sql.Tx, regexID int64, mappings []RegexTagMapping) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM sync_regex_mapping WHERE regex_id = ?`, regexID); err != nil {
		return err
	}
	for _, item := range mappings {
		if _, err := tx.ExecContext(ctx, `INSERT INTO sync_regex_mapping (regex_id, capture_index, tag_key) VALUES (?, ?, ?)`, regexID, item.CaptureIndex, strings.TrimSpace(item.TagKey)); err != nil {
			return err
		}
	}
	return nil
}

func (s *OpsService) ListImageProcessors(ctx context.Context) ([]ImageProcessorRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, processor_type, config_json, status, remark, created_at, updated_at FROM sync_image_process WHERE deleted_at IS NULL ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ImageProcessorRecord{}
	for rows.Next() {
		var item ImageProcessorRecord
		var configRaw []byte
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.ProcessorType, &configRaw, &item.Status, &item.Remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.ConfigJSON = normalizeJSONObject(configRaw)
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateImageProcessor(ctx context.Context, req ImageProcessorMutation) (ImageProcessorRecord, error) {
	payload, err := normalizeImageProcessorMutation(req)
	if err != nil {
		return ImageProcessorRecord{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO sync_image_process (name, processor_type, config_json, status, remark) VALUES (?, ?, ?, ?, ?)`,
		payload.Name, payload.ProcessorType, []byte(payload.ConfigJSON), payload.Status, payload.Remark,
	)
	if err != nil {
		return ImageProcessorRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.getImageProcessor(ctx, id)
}

func (s *OpsService) UpdateImageProcessor(ctx context.Context, id int64, req ImageProcessorMutation) (ImageProcessorRecord, error) {
	payload, err := normalizeImageProcessorMutation(req)
	if err != nil {
		return ImageProcessorRecord{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE sync_image_process SET name = ?, processor_type = ?, config_json = ?, status = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		payload.Name, payload.ProcessorType, []byte(payload.ConfigJSON), payload.Status, payload.Remark, id,
	); err != nil {
		return ImageProcessorRecord{}, err
	}
	return s.getImageProcessor(ctx, id)
}

func (s *OpsService) DeleteImageProcessor(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sync_image_process SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, id)
	return err
}

func (s *OpsService) getImageProcessor(ctx context.Context, id int64) (ImageProcessorRecord, error) {
	var item ImageProcessorRecord
	var configRaw []byte
	var createdAt time.Time
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id, name, processor_type, config_json, status, remark, created_at, updated_at FROM sync_image_process WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&item.ID, &item.Name, &item.ProcessorType, &configRaw, &item.Status, &item.Remark, &createdAt, &updatedAt)
	if err != nil {
		return ImageProcessorRecord{}, err
	}
	item.ConfigJSON = normalizeJSONObject(configRaw)
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func (s *OpsService) ListAlertGroups(ctx context.Context) ([]AlertGroupRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, receivers, status, remark, created_at, updated_at FROM sync_alarm_group WHERE deleted_at IS NULL ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []AlertGroupRecord{}
	for rows.Next() {
		var item AlertGroupRecord
		var receiversRaw []byte
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &receiversRaw, &item.Status, &item.Remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.Receivers = decodeStringArray(receiversRaw)
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateAlertGroup(ctx context.Context, req AlertGroupMutation) (AlertGroupRecord, error) {
	payload, err := normalizeAlertGroupMutation(req)
	if err != nil {
		return AlertGroupRecord{}, err
	}
	receiversRaw, _ := json.Marshal(payload.Receivers)
	result, err := s.db.ExecContext(ctx, `INSERT INTO sync_alarm_group (name, receivers, status, remark) VALUES (?, ?, ?, ?)`,
		payload.Name, receiversRaw, payload.Status, payload.Remark,
	)
	if err != nil {
		return AlertGroupRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.getAlertGroup(ctx, id)
}

func (s *OpsService) UpdateAlertGroup(ctx context.Context, id int64, req AlertGroupMutation) (AlertGroupRecord, error) {
	payload, err := normalizeAlertGroupMutation(req)
	if err != nil {
		return AlertGroupRecord{}, err
	}
	receiversRaw, _ := json.Marshal(payload.Receivers)
	if _, err := s.db.ExecContext(ctx, `UPDATE sync_alarm_group SET name = ?, receivers = ?, status = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		payload.Name, receiversRaw, payload.Status, payload.Remark, id,
	); err != nil {
		return AlertGroupRecord{}, err
	}
	return s.getAlertGroup(ctx, id)
}

func (s *OpsService) DeleteAlertGroup(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sync_alarm_group SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, id)
	return err
}

func (s *OpsService) getAlertGroup(ctx context.Context, id int64) (AlertGroupRecord, error) {
	var item AlertGroupRecord
	var receiversRaw []byte
	var createdAt time.Time
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id, name, receivers, status, remark, created_at, updated_at FROM sync_alarm_group WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&item.ID, &item.Name, &receiversRaw, &item.Status, &item.Remark, &createdAt, &updatedAt)
	if err != nil {
		return AlertGroupRecord{}, err
	}
	item.Receivers = decodeStringArray(receiversRaw)
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func (s *OpsService) ListMessageChannels(ctx context.Context) ([]MessageChannelRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, channel_type, config_json, status, remark, created_at, updated_at FROM sync_alarm_message WHERE deleted_at IS NULL ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []MessageChannelRecord{}
	for rows.Next() {
		var item MessageChannelRecord
		var configRaw []byte
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.ChannelType, &configRaw, &item.Status, &item.Remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.ConfigJSON = normalizeJSONObject(configRaw)
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateMessageChannel(ctx context.Context, req MessageChannelMutation) (MessageChannelRecord, error) {
	payload, err := normalizeMessageChannelMutation(req)
	if err != nil {
		return MessageChannelRecord{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO sync_alarm_message (name, channel_type, config_json, status, remark) VALUES (?, ?, ?, ?, ?)`,
		payload.Name, payload.ChannelType, []byte(payload.ConfigJSON), payload.Status, payload.Remark,
	)
	if err != nil {
		return MessageChannelRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.getMessageChannel(ctx, id)
}

func (s *OpsService) UpdateMessageChannel(ctx context.Context, id int64, req MessageChannelMutation) (MessageChannelRecord, error) {
	payload, err := normalizeMessageChannelMutation(req)
	if err != nil {
		return MessageChannelRecord{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE sync_alarm_message SET name = ?, channel_type = ?, config_json = ?, status = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		payload.Name, payload.ChannelType, []byte(payload.ConfigJSON), payload.Status, payload.Remark, id,
	); err != nil {
		return MessageChannelRecord{}, err
	}
	return s.getMessageChannel(ctx, id)
}

func (s *OpsService) DeleteMessageChannel(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sync_alarm_message SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, id)
	return err
}

func (s *OpsService) getMessageChannel(ctx context.Context, id int64) (MessageChannelRecord, error) {
	var item MessageChannelRecord
	var configRaw []byte
	var createdAt time.Time
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id, name, channel_type, config_json, status, remark, created_at, updated_at FROM sync_alarm_message WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&item.ID, &item.Name, &item.ChannelType, &configRaw, &item.Status, &item.Remark, &createdAt, &updatedAt)
	if err != nil {
		return MessageChannelRecord{}, err
	}
	item.ConfigJSON = normalizeJSONObject(configRaw)
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func (s *OpsService) ListAlertPolicies(ctx context.Context) ([]AlertPolicyRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, cpu_threshold, mem_threshold, disk_threshold, cpu_consecutive_times, mem_consecutive_times, heartbeat_timeout_seconds, send_frequency_seconds, status, remark, created_at, updated_at FROM sync_alarm_rule WHERE deleted_at IS NULL ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []AlertPolicyRecord{}
	for rows.Next() {
		var item AlertPolicyRecord
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.Name, &item.CPUThreshold, &item.MemThreshold, &item.DiskThreshold, &item.CPUConsecutiveTimes, &item.MemConsecutiveTimes, &item.HeartbeatTimeoutSeconds, &item.SendFrequencySeconds, &item.Status, &item.Remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateAlertPolicy(ctx context.Context, req AlertPolicyMutation) (AlertPolicyRecord, error) {
	payload, err := normalizeAlertPolicyMutation(req)
	if err != nil {
		return AlertPolicyRecord{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO sync_alarm_rule (name, cpu_threshold, mem_threshold, disk_threshold, cpu_consecutive_times, mem_consecutive_times, heartbeat_timeout_seconds, send_frequency_seconds, status, remark) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		payload.Name, payload.CPUThreshold, payload.MemThreshold, payload.DiskThreshold, payload.CPUConsecutiveTimes, payload.MemConsecutiveTimes, payload.HeartbeatTimeoutSeconds, payload.SendFrequencySeconds, payload.Status, payload.Remark,
	)
	if err != nil {
		return AlertPolicyRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.getAlertPolicy(ctx, id)
}

func (s *OpsService) UpdateAlertPolicy(ctx context.Context, id int64, req AlertPolicyMutation) (AlertPolicyRecord, error) {
	payload, err := normalizeAlertPolicyMutation(req)
	if err != nil {
		return AlertPolicyRecord{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE sync_alarm_rule SET name = ?, cpu_threshold = ?, mem_threshold = ?, disk_threshold = ?, cpu_consecutive_times = ?, mem_consecutive_times = ?, heartbeat_timeout_seconds = ?, send_frequency_seconds = ?, status = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		payload.Name, payload.CPUThreshold, payload.MemThreshold, payload.DiskThreshold, payload.CPUConsecutiveTimes, payload.MemConsecutiveTimes, payload.HeartbeatTimeoutSeconds, payload.SendFrequencySeconds, payload.Status, payload.Remark, id,
	); err != nil {
		return AlertPolicyRecord{}, err
	}
	return s.getAlertPolicy(ctx, id)
}

func (s *OpsService) DeleteAlertPolicy(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sync_alarm_rule SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, id)
	return err
}

func (s *OpsService) getAlertPolicy(ctx context.Context, id int64) (AlertPolicyRecord, error) {
	var item AlertPolicyRecord
	var createdAt time.Time
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id, name, cpu_threshold, mem_threshold, disk_threshold, cpu_consecutive_times, mem_consecutive_times, heartbeat_timeout_seconds, send_frequency_seconds, status, remark, created_at, updated_at FROM sync_alarm_rule WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&item.ID, &item.Name, &item.CPUThreshold, &item.MemThreshold, &item.DiskThreshold, &item.CPUConsecutiveTimes, &item.MemConsecutiveTimes, &item.HeartbeatTimeoutSeconds, &item.SendFrequencySeconds, &item.Status, &item.Remark, &createdAt, &updatedAt)
	if err != nil {
		return AlertPolicyRecord{}, err
	}
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func (s *OpsService) ListFileLogs(ctx context.Context) ([]FileLogRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, user_id, username, storage_id, storage_name, file_path, operation_type, result_status, client_ip, message, created_at FROM sync_file_log ORDER BY id DESC LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []FileLogRecord{}
	for rows.Next() {
		var item FileLogRecord
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.UserID, &item.Username, &item.StorageID, &item.StorageName, &item.FilePath, &item.OperationType, &item.ResultStatus, &item.ClientIP, &item.Message, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) LogFileOperation(ctx context.Context, item FileLogRecord) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO sync_file_log (user_id, username, storage_id, storage_name, file_path, operation_type, result_status, client_ip, message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.UserID, item.Username, item.StorageID, item.StorageName, item.FilePath, item.OperationType, firstNonEmpty(item.ResultStatus, "success"), item.ClientIP, item.Message,
	)
	return err
}

func (s *OpsService) ListTaskProgress(ctx context.Context) ([]TaskProgressRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, task_type, status, total_count, success_count, failed_count, payload_json, result_json, started_at, finished_at, created_at, remark FROM sync_file_progress ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []TaskProgressRecord{}
	for rows.Next() {
		var item TaskProgressRecord
		var payloadRaw []byte
		var resultRaw []byte
		var startedAt sql.NullTime
		var finishedAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.TaskType, &item.Status, &item.TotalCount, &item.SuccessCount, &item.FailedCount, &payloadRaw, &resultRaw, &startedAt, &finishedAt, &createdAt, &item.Remark); err != nil {
			return nil, err
		}
		item.PayloadJSON = normalizeJSONObject(payloadRaw)
		item.ResultJSON = normalizeJSONObject(resultRaw)
		if startedAt.Valid {
			item.StartedAt = startedAt.Time.Format(mysqlTimeFormat)
		}
		if finishedAt.Valid {
			item.FinishedAt = finishedAt.Time.Format(mysqlTimeFormat)
		}
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateTaskProgress(ctx context.Context, req TaskProgressMutation) (TaskProgressRecord, error) {
	payload, err := normalizeTaskProgressMutation(req)
	if err != nil {
		return TaskProgressRecord{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO sync_file_progress (task_type, status, total_count, success_count, failed_count, payload_json, result_json, started_at, finished_at, remark) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		payload.TaskType, payload.Status, payload.TotalCount, payload.SuccessCount, payload.FailedCount, []byte(payload.PayloadJSON), nullableJSON(payload.ResultJSON), nullTimeValue(payload.StartedAt), nullTimeValue(payload.FinishedAt), payload.Remark,
	)
	if err != nil {
		return TaskProgressRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.getTaskProgress(ctx, id)
}

func (s *OpsService) UpdateTaskProgress(ctx context.Context, id int64, req TaskProgressMutation) (TaskProgressRecord, error) {
	payload, err := normalizeTaskProgressMutation(req)
	if err != nil {
		return TaskProgressRecord{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE sync_file_progress SET task_type = ?, status = ?, total_count = ?, success_count = ?, failed_count = ?, payload_json = ?, result_json = ?, started_at = ?, finished_at = ?, remark = ? WHERE id = ?`,
		payload.TaskType, payload.Status, payload.TotalCount, payload.SuccessCount, payload.FailedCount, []byte(payload.PayloadJSON), nullableJSON(payload.ResultJSON), nullTimeValue(payload.StartedAt), nullTimeValue(payload.FinishedAt), payload.Remark, id,
	); err != nil {
		return TaskProgressRecord{}, err
	}
	return s.getTaskProgress(ctx, id)
}

func (s *OpsService) DeleteTaskProgress(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sync_file_progress WHERE id = ?`, id)
	return err
}

func (s *OpsService) getTaskProgress(ctx context.Context, id int64) (TaskProgressRecord, error) {
	var item TaskProgressRecord
	var payloadRaw []byte
	var resultRaw []byte
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	var createdAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id, task_type, status, total_count, success_count, failed_count, payload_json, result_json, started_at, finished_at, created_at, remark FROM sync_file_progress WHERE id = ?`, id).
		Scan(&item.ID, &item.TaskType, &item.Status, &item.TotalCount, &item.SuccessCount, &item.FailedCount, &payloadRaw, &resultRaw, &startedAt, &finishedAt, &createdAt, &item.Remark)
	if err != nil {
		return TaskProgressRecord{}, err
	}
	item.PayloadJSON = normalizeJSONObject(payloadRaw)
	item.ResultJSON = normalizeJSONObject(resultRaw)
	if startedAt.Valid {
		item.StartedAt = startedAt.Time.Format(mysqlTimeFormat)
	}
	if finishedAt.Valid {
		item.FinishedAt = finishedAt.Time.Format(mysqlTimeFormat)
	}
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	return item, nil
}

func (s *OpsService) ListAlertLogs(ctx context.Context) ([]AlertLogRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, agent_id, group_id, rule_id, message_channel_id, alert_level, alert_title, alert_body, send_status, failure_reason, created_at FROM sync_alarm_logs ORDER BY id DESC LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []AlertLogRecord{}
	for rows.Next() {
		var item AlertLogRecord
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.AgentID, &item.GroupID, &item.RuleID, &item.MessageChannelID, &item.AlertLevel, &item.AlertTitle, &item.AlertBody, &item.SendStatus, &item.FailureReason, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) ListSystemConfigs(ctx context.Context) ([]SystemConfigRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, config_group, config_key, config_value, value_type, is_encrypted, status, remark, created_at, updated_at FROM sys_config WHERE deleted_at IS NULL ORDER BY config_group ASC, config_key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SystemConfigRecord{}
	for rows.Next() {
		var item SystemConfigRecord
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.ConfigGroup, &item.ConfigKey, &item.ConfigValue, &item.ValueType, &item.IsEncrypted, &item.Status, &item.Remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateSystemConfig(ctx context.Context, req SystemConfigMutation) (SystemConfigRecord, error) {
	payload, err := normalizeSystemConfigMutation(req)
	if err != nil {
		return SystemConfigRecord{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO sys_config (config_group, config_key, config_value, value_type, is_encrypted, status, remark) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		payload.ConfigGroup, payload.ConfigKey, payload.ConfigValue, payload.ValueType, payload.IsEncrypted, payload.Status, payload.Remark,
	)
	if err != nil {
		return SystemConfigRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.getSystemConfig(ctx, id)
}

func (s *OpsService) UpdateSystemConfig(ctx context.Context, id int64, req SystemConfigMutation) (SystemConfigRecord, error) {
	payload, err := normalizeSystemConfigMutation(req)
	if err != nil {
		return SystemConfigRecord{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE sys_config SET config_group = ?, config_key = ?, config_value = ?, value_type = ?, is_encrypted = ?, status = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		payload.ConfigGroup, payload.ConfigKey, payload.ConfigValue, payload.ValueType, payload.IsEncrypted, payload.Status, payload.Remark, id,
	); err != nil {
		return SystemConfigRecord{}, err
	}
	return s.getSystemConfig(ctx, id)
}

func (s *OpsService) DeleteSystemConfig(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sys_config SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, id)
	return err
}

func (s *OpsService) getSystemConfig(ctx context.Context, id int64) (SystemConfigRecord, error) {
	var item SystemConfigRecord
	var createdAt time.Time
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id, config_group, config_key, config_value, value_type, is_encrypted, status, remark, created_at, updated_at FROM sys_config WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&item.ID, &item.ConfigGroup, &item.ConfigKey, &item.ConfigValue, &item.ValueType, &item.IsEncrypted, &item.Status, &item.Remark, &createdAt, &updatedAt)
	if err != nil {
		return SystemConfigRecord{}, err
	}
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func (s *OpsService) ListFilePermissions(ctx context.Context) ([]FilePermissionRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.user_id, u.username, p.storage_id, st.name,
		       p.can_view, p.can_upload, p.can_download, p.can_delete, p.can_batch_download, p.can_download_to_server,
		       p.created_at, p.updated_at
		FROM sync_user_auth p
		LEFT JOIN sys_user u ON u.id = p.user_id
		LEFT JOIN storage st ON st.id = p.storage_id
		ORDER BY p.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []FilePermissionRecord{}
	for rows.Next() {
		var item FilePermissionRecord
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.UserID, &item.Username, &item.StorageID, &item.StorageName, &item.CanView, &item.CanUpload, &item.CanDownload, &item.CanDelete, &item.CanBatchDownload, &item.CanDownloadToServer, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateFilePermission(ctx context.Context, req FilePermissionMutation) (FilePermissionRecord, error) {
	payload, err := normalizeFilePermissionMutation(req)
	if err != nil {
		return FilePermissionRecord{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO sync_user_auth (user_id, storage_id, can_view, can_upload, can_download, can_delete, can_batch_download, can_download_to_server) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		payload.UserID, payload.StorageID, payload.CanView, payload.CanUpload, payload.CanDownload, payload.CanDelete, payload.CanBatchDownload, payload.CanDownloadToServer,
	)
	if err != nil {
		return FilePermissionRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.getFilePermission(ctx, id)
}

func (s *OpsService) UpdateFilePermission(ctx context.Context, id int64, req FilePermissionMutation) (FilePermissionRecord, error) {
	payload, err := normalizeFilePermissionMutation(req)
	if err != nil {
		return FilePermissionRecord{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE sync_user_auth SET user_id = ?, storage_id = ?, can_view = ?, can_upload = ?, can_download = ?, can_delete = ?, can_batch_download = ?, can_download_to_server = ? WHERE id = ?`,
		payload.UserID, payload.StorageID, payload.CanView, payload.CanUpload, payload.CanDownload, payload.CanDelete, payload.CanBatchDownload, payload.CanDownloadToServer, id,
	); err != nil {
		return FilePermissionRecord{}, err
	}
	return s.getFilePermission(ctx, id)
}

func (s *OpsService) DeleteFilePermission(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sync_user_auth WHERE id = ?`, id)
	return err
}

func (s *OpsService) getFilePermission(ctx context.Context, id int64) (FilePermissionRecord, error) {
	var item FilePermissionRecord
	var createdAt time.Time
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT p.id, p.user_id, u.username, p.storage_id, st.name,
		       p.can_view, p.can_upload, p.can_download, p.can_delete, p.can_batch_download, p.can_download_to_server,
		       p.created_at, p.updated_at
		FROM sync_user_auth p
		LEFT JOIN sys_user u ON u.id = p.user_id
		LEFT JOIN storage st ON st.id = p.storage_id
		WHERE p.id = ?
	`, id).Scan(&item.ID, &item.UserID, &item.Username, &item.StorageID, &item.StorageName, &item.CanView, &item.CanUpload, &item.CanDownload, &item.CanDelete, &item.CanBatchDownload, &item.CanDownloadToServer, &createdAt, &updatedAt)
	if err != nil {
		return FilePermissionRecord{}, err
	}
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func (s *OpsService) ListLicenses(ctx context.Context) ([]LicenseRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, license_code, serial_number, max_agent_count, issued_at, expired_at, trial_days, status, remark, created_at, updated_at FROM sync_license ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []LicenseRecord{}
	for rows.Next() {
		var item LicenseRecord
		var issuedAt sql.NullTime
		var expiredAt sql.NullTime
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.LicenseCode, &item.SerialNumber, &item.MaxAgentCount, &issuedAt, &expiredAt, &item.TrialDays, &item.Status, &item.Remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if issuedAt.Valid {
			item.IssuedAt = issuedAt.Time.Format(mysqlTimeFormat)
		}
		if expiredAt.Valid {
			item.ExpiredAt = expiredAt.Time.Format(mysqlTimeFormat)
		}
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *OpsService) CreateLicense(ctx context.Context, req LicenseMutation) (LicenseRecord, error) {
	payload, err := normalizeLicenseMutation(req)
	if err != nil {
		return LicenseRecord{}, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO sync_license (license_code, serial_number, max_agent_count, issued_at, expired_at, trial_days, status, remark) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		payload.LicenseCode, payload.SerialNumber, payload.MaxAgentCount, nullTimeValue(payload.IssuedAt), nullTimeValue(payload.ExpiredAt), payload.TrialDays, payload.Status, payload.Remark,
	)
	if err != nil {
		return LicenseRecord{}, err
	}
	id, _ := result.LastInsertId()
	return s.getLicense(ctx, id)
}

func (s *OpsService) UpdateLicense(ctx context.Context, id int64, req LicenseMutation) (LicenseRecord, error) {
	payload, err := normalizeLicenseMutation(req)
	if err != nil {
		return LicenseRecord{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE sync_license SET license_code = ?, serial_number = ?, max_agent_count = ?, issued_at = ?, expired_at = ?, trial_days = ?, status = ?, remark = ? WHERE id = ?`,
		payload.LicenseCode, payload.SerialNumber, payload.MaxAgentCount, nullTimeValue(payload.IssuedAt), nullTimeValue(payload.ExpiredAt), payload.TrialDays, payload.Status, payload.Remark, id,
	); err != nil {
		return LicenseRecord{}, err
	}
	return s.getLicense(ctx, id)
}

func (s *OpsService) DeleteLicense(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sync_license WHERE id = ?`, id)
	return err
}

func (s *OpsService) getLicense(ctx context.Context, id int64) (LicenseRecord, error) {
	var item LicenseRecord
	var issuedAt sql.NullTime
	var expiredAt sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id, license_code, serial_number, max_agent_count, issued_at, expired_at, trial_days, status, remark, created_at, updated_at FROM sync_license WHERE id = ?`, id).
		Scan(&item.ID, &item.LicenseCode, &item.SerialNumber, &item.MaxAgentCount, &issuedAt, &expiredAt, &item.TrialDays, &item.Status, &item.Remark, &createdAt, &updatedAt)
	if err != nil {
		return LicenseRecord{}, err
	}
	if issuedAt.Valid {
		item.IssuedAt = issuedAt.Time.Format(mysqlTimeFormat)
	}
	if expiredAt.Valid {
		item.ExpiredAt = expiredAt.Time.Format(mysqlTimeFormat)
	}
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func decodeStringArray(raw []byte) []string {
	items := []string{}
	if len(raw) == 0 {
		return items
	}
	_ = json.Unmarshal(raw, &items)
	return normalizeStringList(items)
}

func normalizeJSONObject(raw []byte) string {
	if len(raw) == 0 {
		return "{}"
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return "{}"
	}
	return trimmed
}

func nullableJSON(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "{}" {
		return nil
	}
	return []byte(trimmed)
}

func normalizeFileFilterMutation(req FileFilterMutation) (FileFilterMutation, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.FilterScope = strings.TrimSpace(req.FilterScope)
	req.ListType = strings.TrimSpace(req.ListType)
	req.Remark = strings.TrimSpace(req.Remark)
	req.Patterns = normalizeStringList(req.Patterns)
	if req.Name == "" {
		return req, errors.New("filter name is required")
	}
	if req.FilterScope == "" {
		req.FilterScope = "extension"
	}
	if req.ListType == "" {
		req.ListType = "whitelist"
	}
	return req, nil
}

func normalizeRegexRuleMutation(req RegexRuleMutation) (RegexRuleMutation, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.SourceField = strings.TrimSpace(req.SourceField)
	req.Regexp = strings.TrimSpace(req.Regexp)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Name == "" {
		return req, errors.New("regex rule name is required")
	}
	if req.SourceField == "" {
		req.SourceField = "path"
	}
	normalized := make([]RegexTagMapping, 0, len(req.Mappings))
	for _, item := range req.Mappings {
		item.TagKey = strings.TrimSpace(item.TagKey)
		if item.TagKey == "" {
			continue
		}
		normalized = append(normalized, item)
	}
	req.Mappings = normalized
	return req, nil
}

func normalizeImageProcessorMutation(req ImageProcessorMutation) (ImageProcessorMutation, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.ProcessorType = strings.TrimSpace(req.ProcessorType)
	req.ConfigJSON = strings.TrimSpace(req.ConfigJSON)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Name == "" {
		return req, errors.New("image processor name is required")
	}
	if req.ProcessorType == "" {
		req.ProcessorType = "transform"
	}
	if req.ConfigJSON == "" {
		req.ConfigJSON = "{}"
	}
	if !json.Valid([]byte(req.ConfigJSON)) {
		return req, errors.New("image processor config must be valid JSON")
	}
	return req, nil
}

func normalizeAlertGroupMutation(req AlertGroupMutation) (AlertGroupMutation, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Remark = strings.TrimSpace(req.Remark)
	req.Receivers = normalizeStringList(req.Receivers)
	if req.Name == "" {
		return req, errors.New("alert group name is required")
	}
	return req, nil
}

func normalizeMessageChannelMutation(req MessageChannelMutation) (MessageChannelMutation, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.ChannelType = strings.TrimSpace(req.ChannelType)
	req.ConfigJSON = strings.TrimSpace(req.ConfigJSON)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Name == "" {
		return req, errors.New("message channel name is required")
	}
	if req.ChannelType == "" {
		req.ChannelType = "email"
	}
	if req.ConfigJSON == "" {
		req.ConfigJSON = "{}"
	}
	if !json.Valid([]byte(req.ConfigJSON)) {
		return req, errors.New("message channel config must be valid JSON")
	}
	return req, nil
}

func normalizeAlertPolicyMutation(req AlertPolicyMutation) (AlertPolicyMutation, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.Name == "" {
		return req, errors.New("alert policy name is required")
	}
	return req, nil
}

func normalizeTaskProgressMutation(req TaskProgressMutation) (TaskProgressMutation, error) {
	req.TaskType = strings.TrimSpace(req.TaskType)
	req.Status = strings.TrimSpace(req.Status)
	req.PayloadJSON = strings.TrimSpace(req.PayloadJSON)
	req.ResultJSON = strings.TrimSpace(req.ResultJSON)
	req.StartedAt = strings.TrimSpace(req.StartedAt)
	req.FinishedAt = strings.TrimSpace(req.FinishedAt)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.TaskType == "" {
		return req, errors.New("task type is required")
	}
	if req.Status == "" {
		req.Status = "pending"
	}
	if req.PayloadJSON == "" {
		req.PayloadJSON = "{}"
	}
	if !json.Valid([]byte(req.PayloadJSON)) {
		return req, errors.New("task payload must be valid JSON")
	}
	if req.ResultJSON != "" && req.ResultJSON != "{}" && !json.Valid([]byte(req.ResultJSON)) {
		return req, errors.New("task result must be valid JSON")
	}
	return req, nil
}

func normalizeSystemConfigMutation(req SystemConfigMutation) (SystemConfigMutation, error) {
	req.ConfigGroup = strings.TrimSpace(req.ConfigGroup)
	req.ConfigKey = strings.TrimSpace(req.ConfigKey)
	req.ConfigValue = strings.TrimSpace(req.ConfigValue)
	req.ValueType = strings.TrimSpace(req.ValueType)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.ConfigGroup == "" {
		req.ConfigGroup = "system"
	}
	if req.ConfigKey == "" {
		return req, errors.New("config key is required")
	}
	if req.ValueType == "" {
		req.ValueType = "string"
	}
	return req, nil
}

func normalizeFilePermissionMutation(req FilePermissionMutation) (FilePermissionMutation, error) {
	if req.UserID <= 0 {
		return req, errors.New("user is required")
	}
	if req.StorageID <= 0 {
		return req, errors.New("storage is required")
	}
	return req, nil
}

func normalizeLicenseMutation(req LicenseMutation) (LicenseMutation, error) {
	req.LicenseCode = strings.TrimSpace(req.LicenseCode)
	req.SerialNumber = strings.TrimSpace(req.SerialNumber)
	req.IssuedAt = strings.TrimSpace(req.IssuedAt)
	req.ExpiredAt = strings.TrimSpace(req.ExpiredAt)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.SerialNumber == "" {
		return req, errors.New("serial number is required")
	}
	return req, nil
}

func ensureAffected(result sql.Result, label string) error {
	if result == nil {
		return nil
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%s not found", label)
	}
	return nil
}
