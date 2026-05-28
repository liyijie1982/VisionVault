package service

import (
	"context"
	"database/sql"
	"skybase/internal/domain/auth"
	"strings"
	"time"
)

const mysqlTimeFormat = "2006-01-02 15:04:05"

type LoginLogService struct {
	db *sql.DB
}

type LoginLogRecordInput struct {
	UserID      int64
	Username    string
	LoginIP     string
	UserAgent   string
	LoginStatus int
	Message     string
}

type LoginLogQuery struct {
	Username    string
	LoginIP     string
	LoginStatus *int
	StartAt     *time.Time
	EndAt       *time.Time
	Page        int
	PageSize    int
}

type LoginLogListResult struct {
	Items    []auth.LoginLog `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

func NewLoginLogService(db *sql.DB) *LoginLogService {
	return &LoginLogService{db: db}
}

func (s *LoginLogService) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *LoginLogService) Record(ctx context.Context, input LoginLogRecordInput) error {
	const query = `
		INSERT INTO sys_login_log (user_id, username, login_ip, user_agent, login_status, message)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(
		ctx,
		query,
		input.UserID,
		limitString(input.Username, 64),
		limitString(input.LoginIP, 64),
		limitString(input.UserAgent, 512),
		input.LoginStatus,
		limitString(input.Message, 500),
	)
	return err
}

func (s *LoginLogService) List(ctx context.Context, query LoginLogQuery) (LoginLogListResult, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	conditions := []string{"1=1"}
	args := make([]any, 0, 8)

	if username := strings.TrimSpace(query.Username); username != "" {
		conditions = append(conditions, "username LIKE ?")
		args = append(args, "%"+username+"%")
	}
	if loginIP := strings.TrimSpace(query.LoginIP); loginIP != "" {
		conditions = append(conditions, "login_ip LIKE ?")
		args = append(args, "%"+loginIP+"%")
	}
	if query.LoginStatus != nil {
		conditions = append(conditions, "login_status = ?")
		args = append(args, *query.LoginStatus)
	}
	if query.StartAt != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, query.StartAt.Format(mysqlTimeFormat))
	}
	if query.EndAt != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, query.EndAt.Format(mysqlTimeFormat))
	}

	whereClause := strings.Join(conditions, " AND ")

	var total int64
	countSQL := "SELECT COUNT(1) FROM sys_login_log WHERE " + whereClause
	if err := s.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return LoginLogListResult{}, err
	}

	dataSQL := `
		SELECT id, user_id, username, login_ip, user_agent, login_status, message, created_at
		FROM sys_login_log
		WHERE ` + whereClause + `
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`
	dataArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	rows, err := s.db.QueryContext(ctx, dataSQL, dataArgs...)
	if err != nil {
		return LoginLogListResult{}, err
	}
	defer rows.Close()

	items := make([]auth.LoginLog, 0, pageSize)
	for rows.Next() {
		var item auth.LoginLog
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Username,
			&item.LoginIP,
			&item.UserAgent,
			&item.LoginStatus,
			&item.Message,
			&createdAt,
		); err != nil {
			return LoginLogListResult{}, err
		}
		item.CreatedAt = createdAt.Format(mysqlTimeFormat)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return LoginLogListResult{}, err
	}

	return LoginLogListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func limitString(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}
