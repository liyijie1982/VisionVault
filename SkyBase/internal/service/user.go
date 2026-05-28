package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"skybase/internal/domain/auth"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUsernameRequired = errors.New("username is required")
var ErrUserNicknameRequired = errors.New("user nickname is required")
var ErrUserPhoneRequired = errors.New("user phone is required")
var ErrUserDepartmentRequired = errors.New("user department is required")
var ErrUserPasswordRequired = errors.New("initial password is required")
var ErrUsernameExists = errors.New("username already exists")
var ErrUserPhoneExists = errors.New("phone already exists")

type UserMutation struct {
	DeptID                int64   `json:"deptId"`
	Username              string  `json:"username"`
	Nickname              string  `json:"nickname"`
	RealName              string  `json:"realName"`
	Phone                 string  `json:"phone"`
	Email                 string  `json:"email"`
	Password              string  `json:"password"`
	Status                int     `json:"status"`
	RoleIDs               []int64 `json:"roleIds"`
	PasswordResetRequired int     `json:"passwordResetRequired"`
}

type UserService struct {
	db   *sql.DB
	dept *DepartmentService
}

func NewUserService(db *sql.DB, deptService *DepartmentService) *UserService {
	return &UserService{db: db, dept: deptService}
}

func (s *UserService) List(ctx context.Context) ([]auth.UserRecord, error) {
	const query = `
		SELECT
			u.id, u.dept_id, COALESCE(d.name, ''), u.username, u.nickname, u.real_name, u.phone, u.email,
			u.status, u.last_login_ip, u.last_login_at, u.password_reset_required, u.created_at, u.updated_at
		FROM sys_user u
		LEFT JOIN sys_dept d ON d.id = u.dept_id AND d.deleted_at IS NULL
		WHERE u.deleted_at IS NULL
		ORDER BY u.updated_at DESC, u.id DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]auth.UserRecord, 0)
	ids := make([]int64, 0)
	for rows.Next() {
		item, scanErr := scanUserRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
		ids = append(ids, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	roleMap, err := s.loadUserRoles(ctx, ids)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].RoleIDs = roleMap[items[index].ID].ids
		items[index].RoleNames = roleMap[items[index].ID].names
	}
	return items, nil
}

func (s *UserService) Get(ctx context.Context, id int64) (auth.UserRecord, error) {
	const query = `
		SELECT
			u.id, u.dept_id, COALESCE(d.name, ''), u.username, u.nickname, u.real_name, u.phone, u.email,
			u.status, u.last_login_ip, u.last_login_at, u.password_reset_required, u.created_at, u.updated_at
		FROM sys_user u
		LEFT JOIN sys_dept d ON d.id = u.dept_id AND d.deleted_at IS NULL
		WHERE u.id = ? AND u.deleted_at IS NULL
	`

	row := s.db.QueryRowContext(ctx, query, id)
	item, err := scanUserRecord(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.UserRecord{}, ErrUserNotFound
		}
		return auth.UserRecord{}, err
	}
	roleMap, err := s.loadUserRoles(ctx, []int64{id})
	if err != nil {
		return auth.UserRecord{}, err
	}
	item.RoleIDs = roleMap[id].ids
	item.RoleNames = roleMap[id].names
	return item, nil
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (auth.UserRecord, string, error) {
	const query = `
		SELECT
			u.id, u.dept_id, COALESCE(d.name, ''), u.username, u.nickname, u.real_name, u.phone, u.email,
			u.status, u.last_login_ip, u.last_login_at, u.password_reset_required, u.created_at, u.updated_at, u.password_hash
		FROM sys_user u
		LEFT JOIN sys_dept d ON d.id = u.dept_id AND d.deleted_at IS NULL
		WHERE u.username = ? AND u.deleted_at IS NULL
	`

	var item auth.UserRecord
	var createdAt time.Time
	var updatedAt time.Time
	var lastLoginAt sql.NullTime
	var passwordHash string
	err := s.db.QueryRowContext(ctx, query, strings.TrimSpace(username)).Scan(
		&item.ID,
		&item.DeptID,
		&item.DeptName,
		&item.Username,
		&item.Nickname,
		&item.RealName,
		&item.Phone,
		&item.Email,
		&item.Status,
		&item.LastLoginIP,
		&lastLoginAt,
		&item.PasswordResetRequired,
		&createdAt,
		&updatedAt,
		&passwordHash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.UserRecord{}, "", ErrUserNotFound
		}
		return auth.UserRecord{}, "", err
	}

	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	if lastLoginAt.Valid {
		item.LastLoginAt = lastLoginAt.Time.Format(mysqlTimeFormat)
	}
	roleMap, err := s.loadUserRoles(ctx, []int64{item.ID})
	if err != nil {
		return auth.UserRecord{}, "", err
	}
	item.RoleIDs = roleMap[item.ID].ids
	item.RoleNames = roleMap[item.ID].names
	return item, passwordHash, nil
}

func (s *UserService) Create(ctx context.Context, input UserMutation) (auth.UserRecord, error) {
	normalized, err := normalizeUser(input, true)
	if err != nil {
		return auth.UserRecord{}, err
	}
	if _, err := s.dept.Get(ctx, normalized.DeptID); err != nil {
		if errors.Is(err, ErrDepartmentNotFound) {
			return auth.UserRecord{}, ErrUserDepartmentRequired
		}
		return auth.UserRecord{}, err
	}
	if err := s.ensureRolesExist(ctx, normalized.RoleIDs); err != nil {
		return auth.UserRecord{}, err
	}

	passwordHash, err := hashPassword(normalized.Password)
	if err != nil {
		return auth.UserRecord{}, err
	}

	const query = `
		INSERT INTO sys_user (
			dept_id, username, nickname, real_name, phone, email, password_hash, avatar, status,
			last_login_ip, last_login_at, password_reset_required
		) VALUES (?, ?, ?, ?, ?, ?, ?, '', ?, '', NULL, ?)
	`

	result, err := s.db.ExecContext(
		ctx,
		query,
		normalized.DeptID,
		normalized.Username,
		normalized.Nickname,
		normalized.RealName,
		normalized.Phone,
		normalized.Email,
		passwordHash,
		normalized.Status,
		1,
	)
	if err != nil {
		return auth.UserRecord{}, mapUserWriteError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return auth.UserRecord{}, err
	}
	if err := s.replaceUserRoles(ctx, id, normalized.RoleIDs); err != nil {
		return auth.UserRecord{}, err
	}
	return s.Get(ctx, id)
}

func (s *UserService) Update(ctx context.Context, id int64, input UserMutation) (auth.UserRecord, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return auth.UserRecord{}, err
	}

	normalized, err := normalizeUser(input, false)
	if err != nil {
		return auth.UserRecord{}, err
	}
	if _, err := s.dept.Get(ctx, normalized.DeptID); err != nil {
		if errors.Is(err, ErrDepartmentNotFound) {
			return auth.UserRecord{}, ErrUserDepartmentRequired
		}
		return auth.UserRecord{}, err
	}
	if err := s.ensureRolesExist(ctx, normalized.RoleIDs); err != nil {
		return auth.UserRecord{}, err
	}

	const query = `
		UPDATE sys_user
		SET dept_id = ?, username = ?, nickname = ?, real_name = ?, phone = ?, email = ?, status = ?, password_reset_required = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	if _, err := s.db.ExecContext(
		ctx,
		query,
		normalized.DeptID,
		normalized.Username,
		normalized.Nickname,
		normalized.RealName,
		normalized.Phone,
		normalized.Email,
		normalized.Status,
		normalized.PasswordResetRequired,
		id,
	); err != nil {
		return auth.UserRecord{}, mapUserWriteError(err)
	}
	if err := s.replaceUserRoles(ctx, id, normalized.RoleIDs); err != nil {
		return auth.UserRecord{}, err
	}

	return s.Get(ctx, id)
}

func (s *UserService) ResetPassword(ctx context.Context, id int64, password string) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE sys_user SET password_hash = ?, password_reset_required = 1 WHERE id = ? AND deleted_at IS NULL`,
		passwordHash,
		id,
	)
	return err
}

func (s *UserService) ChangePassword(ctx context.Context, id int64, password string) error {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE sys_user SET password_hash = ?, password_reset_required = 0 WHERE id = ? AND deleted_at IS NULL`,
		passwordHash,
		id,
	)
	return err
}

func (s *UserService) UpdateLoginMeta(ctx context.Context, id int64, loginIP string, loginAt time.Time) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE sys_user SET last_login_ip = ?, last_login_at = ? WHERE id = ? AND deleted_at IS NULL`,
		strings.TrimSpace(loginIP),
		loginAt,
		id,
	)
	return err
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}

	_, err := s.db.ExecContext(
		ctx,
		"UPDATE sys_user SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL",
		id,
	)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, "DELETE FROM sys_user_role WHERE user_id = ?", id)
	return err
}

func normalizeUser(input UserMutation, requirePassword bool) (UserMutation, error) {
	username := strings.TrimSpace(input.Username)
	if username == "" {
		return UserMutation{}, ErrUsernameRequired
	}

	nickname := strings.TrimSpace(input.Nickname)
	if nickname == "" {
		return UserMutation{}, ErrUserNicknameRequired
	}

	phone := strings.TrimSpace(input.Phone)
	if phone == "" {
		return UserMutation{}, ErrUserPhoneRequired
	}

	if input.DeptID <= 0 {
		return UserMutation{}, ErrUserDepartmentRequired
	}

	password := strings.TrimSpace(input.Password)
	if requirePassword && password == "" {
		return UserMutation{}, ErrUserPasswordRequired
	}

	status := 0
	if input.Status != 0 {
		status = 1
	}

	passwordResetRequired := 0
	if input.PasswordResetRequired != 0 {
		passwordResetRequired = 1
	}

	return UserMutation{
		DeptID:                input.DeptID,
		Username:              username,
		Nickname:              nickname,
		RealName:              strings.TrimSpace(input.RealName),
		Phone:                 phone,
		Email:                 strings.TrimSpace(input.Email),
		Password:              password,
		Status:                status,
		RoleIDs:               uniqueInt64s(input.RoleIDs),
		PasswordResetRequired: passwordResetRequired,
	}, nil
}

func mapUserWriteError(err error) error {
	switch mysqlDuplicateField(err) {
	case "username":
		return ErrUsernameExists
	case "phone":
		return ErrUserPhoneExists
	default:
		return err
	}
}

func scanUserRecord(row scanner) (auth.UserRecord, error) {
	var item auth.UserRecord
	var createdAt time.Time
	var updatedAt time.Time
	var lastLoginAt sql.NullTime
	if err := row.Scan(
		&item.ID,
		&item.DeptID,
		&item.DeptName,
		&item.Username,
		&item.Nickname,
		&item.RealName,
		&item.Phone,
		&item.Email,
		&item.Status,
		&item.LastLoginIP,
		&lastLoginAt,
		&item.PasswordResetRequired,
		&createdAt,
		&updatedAt,
	); err != nil {
		return auth.UserRecord{}, err
	}

	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	if lastLoginAt.Valid {
		item.LastLoginAt = lastLoginAt.Time.Format(mysqlTimeFormat)
	}
	return item, nil
}

func ParseEntityID(raw, label string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("%s id is required", label)
	}

	var id int64
	if _, err := fmt.Sscanf(raw, "%d", &id); err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s id", label)
	}
	return id, nil
}

type userRoleBinding struct {
	ids   []int64
	names []string
}

func (s *UserService) loadUserRoles(ctx context.Context, userIDs []int64) (map[int64]userRoleBinding, error) {
	result := make(map[int64]userRoleBinding)
	if len(userIDs) == 0 {
		return result, nil
	}

	query, args := buildInt64InQuery(`
		SELECT ur.user_id, r.id, r.name
		FROM sys_user_role ur
		INNER JOIN sys_role r ON r.id = ur.role_id
		WHERE ur.user_id IN (%s) AND r.deleted_at IS NULL
		ORDER BY r.sort ASC, r.id ASC
	`, userIDs)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		var roleID int64
		var roleName string
		if err := rows.Scan(&userID, &roleID, &roleName); err != nil {
			return nil, err
		}
		item := result[userID]
		item.ids = append(item.ids, roleID)
		item.names = append(item.names, roleName)
		result[userID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *UserService) ensureRolesExist(ctx context.Context, roleIDs []int64) error {
	if len(roleIDs) == 0 {
		return nil
	}

	query, args := buildInt64InQuery(`
		SELECT COUNT(1)
		FROM sys_role
		WHERE id IN (%s) AND deleted_at IS NULL AND status = 1
	`, roleIDs)

	var count int
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return err
	}
	if count != len(roleIDs) {
		return fmt.Errorf("one or more roles are invalid")
	}
	return nil
}

func (s *UserService) replaceUserRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM sys_user_role WHERE user_id = ?`, userID); err != nil {
		return err
	}
	for _, roleID := range uniqueInt64s(roleIDs) {
		if _, err := tx.ExecContext(ctx, `INSERT INTO sys_user_role (user_id, role_id) VALUES (?, ?)`, userID, roleID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{})
	items := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	return items
}
