package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"skybase/internal/domain/auth"
)

var ErrDepartmentNotFound = errors.New("department not found")
var ErrDepartmentNameRequired = errors.New("department name is required")
var ErrDepartmentParentNotFound = errors.New("parent department not found")
var ErrDepartmentParentInvalid = errors.New("department parent is invalid")
var ErrDepartmentHasChildren = errors.New("department has child departments")
var ErrDepartmentHasUsers = errors.New("department has users and cannot be deleted")

type DepartmentMutation struct {
	ParentID int64  `json:"parentId"`
	Name     string `json:"name"`
	Leader   string `json:"leader"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

type DepartmentService struct {
	db *sql.DB
}

func NewDepartmentService(db *sql.DB) *DepartmentService {
	return &DepartmentService{db: db}
}

func (s *DepartmentService) List(ctx context.Context) ([]auth.Department, error) {
	const query = `
		SELECT id, parent_id, ancestors, name, leader, phone, email, sort, status, created_at, updated_at
		FROM sys_dept
		WHERE deleted_at IS NULL
		ORDER BY sort ASC, id ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]auth.Department, 0)
	for rows.Next() {
		item, scanErr := scanDepartment(rows)
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

func (s *DepartmentService) Get(ctx context.Context, id int64) (auth.Department, error) {
	const query = `
		SELECT id, parent_id, ancestors, name, leader, phone, email, sort, status, created_at, updated_at
		FROM sys_dept
		WHERE id = ? AND deleted_at IS NULL
	`

	row := s.db.QueryRowContext(ctx, query, id)
	item, err := scanDepartment(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.Department{}, ErrDepartmentNotFound
		}
		return auth.Department{}, err
	}
	return item, nil
}

func (s *DepartmentService) Create(ctx context.Context, input DepartmentMutation) (auth.Department, error) {
	normalized, err := normalizeDepartment(input)
	if err != nil {
		return auth.Department{}, err
	}

	ancestors, err := s.departmentAncestors(ctx, normalized.ParentID, 0)
	if err != nil {
		return auth.Department{}, err
	}

	const query = `
		INSERT INTO sys_dept (parent_id, ancestors, name, leader, phone, email, sort, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(
		ctx,
		query,
		normalized.ParentID,
		ancestors,
		normalized.Name,
		normalized.Leader,
		normalized.Phone,
		normalized.Email,
		normalized.Sort,
		normalized.Status,
	)
	if err != nil {
		return auth.Department{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return auth.Department{}, err
	}
	return s.Get(ctx, id)
}

func (s *DepartmentService) Update(ctx context.Context, id int64, input DepartmentMutation) (auth.Department, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return auth.Department{}, err
	}

	normalized, err := normalizeDepartment(input)
	if err != nil {
		return auth.Department{}, err
	}

	ancestors, err := s.departmentAncestors(ctx, normalized.ParentID, id)
	if err != nil {
		return auth.Department{}, err
	}

	const query = `
		UPDATE sys_dept
		SET parent_id = ?, ancestors = ?, name = ?, leader = ?, phone = ?, email = ?, sort = ?, status = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	if _, err := s.db.ExecContext(
		ctx,
		query,
		normalized.ParentID,
		ancestors,
		normalized.Name,
		normalized.Leader,
		normalized.Phone,
		normalized.Email,
		normalized.Sort,
		normalized.Status,
		id,
	); err != nil {
		return auth.Department{}, err
	}

	return s.Get(ctx, id)
}

func (s *DepartmentService) Delete(ctx context.Context, id int64) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	var childCount int
	if err := s.db.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM sys_dept WHERE parent_id = ? AND deleted_at IS NULL",
		id,
	).Scan(&childCount); err != nil {
		return err
	}
	if childCount > 0 {
		return ErrDepartmentHasChildren
	}

	var userCount int
	if err := s.db.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM sys_user WHERE dept_id = ? AND deleted_at IS NULL",
		id,
	).Scan(&userCount); err != nil {
		return err
	}
	if userCount > 0 {
		return ErrDepartmentHasUsers
	}

	if _, err := s.db.ExecContext(
		ctx,
		"UPDATE sys_dept SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL",
		id,
	); err != nil {
		return err
	}
	return nil
}

func normalizeDepartment(input DepartmentMutation) (DepartmentMutation, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return DepartmentMutation{}, ErrDepartmentNameRequired
	}

	status := 0
	if input.Status != 0 {
		status = 1
	}

	parentID := input.ParentID
	if parentID < 0 {
		parentID = 0
	}

	return DepartmentMutation{
		ParentID: parentID,
		Name:     name,
		Leader:   strings.TrimSpace(input.Leader),
		Phone:    strings.TrimSpace(input.Phone),
		Email:    strings.TrimSpace(input.Email),
		Sort:     input.Sort,
		Status:   status,
	}, nil
}

func (s *DepartmentService) departmentAncestors(ctx context.Context, parentID, currentID int64) (string, error) {
	if parentID == 0 {
		return "", nil
	}
	if parentID == currentID {
		return "", ErrDepartmentParentInvalid
	}

	parent, err := s.Get(ctx, parentID)
	if err != nil {
		if errors.Is(err, ErrDepartmentNotFound) {
			return "", ErrDepartmentParentNotFound
		}
		return "", err
	}

	if currentID > 0 && stringsContainsID(parent.Ancestors, currentID) {
		return "", ErrDepartmentParentInvalid
	}

	if parent.Ancestors == "" {
		return fmt.Sprintf("%d", parent.ID), nil
	}
	return fmt.Sprintf("%s,%d", parent.Ancestors, parent.ID), nil
}

func stringsContainsID(ancestors string, id int64) bool {
	target := fmt.Sprintf("%d", id)
	for _, item := range strings.Split(ancestors, ",") {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}

type scanner interface {
	Scan(dest ...any) error
}

func scanDepartment(row scanner) (auth.Department, error) {
	var item auth.Department
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(
		&item.ID,
		&item.ParentID,
		&item.Ancestors,
		&item.Name,
		&item.Leader,
		&item.Phone,
		&item.Email,
		&item.Sort,
		&item.Status,
		&createdAt,
		&updatedAt,
	); err != nil {
		return auth.Department{}, err
	}
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func mysqlDuplicateField(err error) string {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return ""
	}

	message := strings.ToLower(mysqlErr.Message)
	switch {
	case strings.Contains(message, "uk_sys_user_username"):
		return "username"
	case strings.Contains(message, "uk_sys_user_phone"):
		return "phone"
	case strings.Contains(message, "uk_sys_role_key"):
		return "role_key"
	case strings.Contains(message, "uk_sys_role_name"):
		return "role_name"
	case strings.Contains(message, "uk_storage_name"):
		return "uk_storage_name"
	default:
		return "duplicate"
	}
}
