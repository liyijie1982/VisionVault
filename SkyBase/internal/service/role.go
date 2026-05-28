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

var ErrRoleNotFound = errors.New("role not found")
var ErrRoleNameExists = errors.New("role name already exists")
var ErrRoleKeyExists = errors.New("role key already exists")
var ErrProtectedRole = errors.New("protected role cannot be deleted")
var ErrInvalidRoleName = errors.New("role name is required")
var ErrInvalidRoleKey = errors.New("role key is required")

type RoleMutation struct {
	Name        string   `json:"name"`
	Key         string   `json:"key"`
	DataScope   string   `json:"dataScope"`
	Sort        int      `json:"sort"`
	Status      int      `json:"status"`
	Description string   `json:"description"`
	MenuKeys    []string `json:"menuKeys"`
}

type RoleService struct {
	db *sql.DB
}

type menuSeed struct {
	ParentKey string
	Name      string
	MenuType  string
	Path      string
	Component string
	RouteName string
	Perms     string
	Icon      string
	Visible   int
	Status    int
	Sort      int
	Remark    string
}

var defaultMenus = []menuSeed{
	{Name: "Overview", MenuType: "menu", Path: "/", Component: "OverviewPage", RouteName: "overview", Perms: "view:overview", Icon: "dashboard", Visible: 1, Status: 1, Sort: 1},
	{Name: "Backup", MenuType: "directory", RouteName: "backup", Icon: "archive", Visible: 1, Status: 1, Sort: 5},
	{ParentKey: "backup", Name: "Tape Devices", MenuType: "menu", Path: "/backup/devices", Component: "BackupDevicesPage", RouteName: "backup-devices", Perms: "view:backup-devices", Icon: "storage", Visible: 1, Status: 1, Sort: 6},
	{ParentKey: "backup", Name: "Task Management", MenuType: "menu", Path: "/backup/tasks", Component: "BackupTasksPage", RouteName: "backup-tasks", Perms: "view:backup-tasks", Icon: "list", Visible: 1, Status: 1, Sort: 7},
	{ParentKey: "backup", Name: "Backup Logs", MenuType: "menu", Path: "/backup/logs", Component: "BackupLogsPage", RouteName: "backup-logs", Perms: "view:backup-logs", Icon: "history", Visible: 1, Status: 1, Sort: 8},
	{Name: "Agent Control", MenuType: "directory", RouteName: "agent-control", Icon: "computer", Visible: 1, Status: 1, Sort: 10},
	{ParentKey: "agent-control", Name: "Agents", MenuType: "menu", Path: "/agents", Component: "AgentsPage", RouteName: "agents", Perms: "view:agents", Icon: "computer", Visible: 1, Status: 1, Sort: 11},
	{ParentKey: "agent-control", Name: "Groups", MenuType: "menu", Path: "/groups", Component: "GroupsPage", RouteName: "groups", Perms: "view:groups", Icon: "apps", Visible: 1, Status: 1, Sort: 12},
	{ParentKey: "agent-control", Name: "Sync Logs", MenuType: "menu", Path: "/sync-logs", Component: "SyncLogsPage", RouteName: "sync-logs", Perms: "view:sync-logs", Icon: "list", Visible: 1, Status: 1, Sort: 13},
	{ParentKey: "agent-control", Name: "Scan Reports", MenuType: "menu", Path: "/scan-reports", Component: "ScanReportsPage", RouteName: "scan-reports", Perms: "view:scan-reports", Icon: "bar-chart", Visible: 1, Status: 1, Sort: 14},
	{ParentKey: "agent-control", Name: "Monitor", MenuType: "menu", Path: "/monitor", Component: "MonitorPage", RouteName: "monitor", Perms: "view:monitor", Icon: "command", Visible: 1, Status: 1, Sort: 15},
	{ParentKey: "agent-control", Name: "Extraction Rules", MenuType: "menu", Path: "/extraction-rules", Component: "SystemModulePage", RouteName: "extraction-rules", Perms: "view:extraction-rules", Icon: "common", Visible: 1, Status: 1, Sort: 16},
	{ParentKey: "agent-control", Name: "Versions", MenuType: "menu", Path: "/versions", Component: "VersionsPage", RouteName: "versions", Perms: "view:versions", Icon: "archive", Visible: 1, Status: 1, Sort: 17},
	{Name: "File Control", MenuType: "directory", RouteName: "file-control", Icon: "folder", Visible: 1, Status: 1, Sort: 20},
	{ParentKey: "file-control", Name: "Files", MenuType: "menu", Path: "/files", Component: "FilesPage", RouteName: "files", Perms: "view:files", Icon: "folder", Visible: 1, Status: 1, Sort: 21},
	{ParentKey: "file-control", Name: "File Logs", MenuType: "menu", Path: "/file-logs", Component: "SystemModulePage", RouteName: "file-logs", Perms: "view:file-logs", Icon: "history", Visible: 1, Status: 1, Sort: 22},
	{ParentKey: "file-control", Name: "File Permissions", MenuType: "menu", Path: "/file-permissions", Component: "SystemModulePage", RouteName: "file-permissions", Perms: "view:file-permissions", Icon: "lock", Visible: 1, Status: 1, Sort: 23},
	{ParentKey: "file-control", Name: "Task Progress", MenuType: "menu", Path: "/task-progress", Component: "SystemModulePage", RouteName: "task-progress", Perms: "view:task-progress", Icon: "list", Visible: 1, Status: 1, Sort: 24},
	{Name: "Alerts", MenuType: "directory", RouteName: "alerts", Icon: "safe", Visible: 1, Status: 1, Sort: 30},
	{ParentKey: "alerts", Name: "Alert Groups", MenuType: "menu", Path: "/alerts/groups", Component: "SystemModulePage", RouteName: "alert-groups", Perms: "view:alert-groups", Icon: "apps", Visible: 1, Status: 1, Sort: 31},
	{ParentKey: "alerts", Name: "Alert Logs", MenuType: "menu", Path: "/alerts/logs", Component: "SystemModulePage", RouteName: "alert-logs", Perms: "view:alert-logs", Icon: "history", Visible: 1, Status: 1, Sort: 32},
	{ParentKey: "alerts", Name: "Message Channels", MenuType: "menu", Path: "/alerts/channels", Component: "SystemModulePage", RouteName: "message-channels", Perms: "view:message-channels", Icon: "common", Visible: 1, Status: 1, Sort: 33},
	{ParentKey: "alerts", Name: "Alert Policies", MenuType: "menu", Path: "/alerts/policies", Component: "SystemModulePage", RouteName: "alert-policies", Perms: "view:alert-policies", Icon: "safe", Visible: 1, Status: 1, Sort: 34},
	{Name: "System", MenuType: "directory", RouteName: "system", Icon: "settings", Visible: 1, Status: 1, Sort: 40},
	{ParentKey: "system", Name: "Storage", MenuType: "menu", Path: "/storage", Component: "StoragePage", RouteName: "storage", Perms: "view:storage", Icon: "storage", Visible: 1, Status: 1, Sort: 41},
	{ParentKey: "system", Name: "Users", MenuType: "menu", Path: "/system/users", Component: "UsersPage", RouteName: "system-users", Perms: "view:system-users", Icon: "user", Visible: 1, Status: 1, Sort: 42},
	{ParentKey: "system", Name: "Departments", MenuType: "menu", Path: "/system/departments", Component: "DepartmentsPage", RouteName: "system-departments", Perms: "view:system-departments", Icon: "user-group", Visible: 1, Status: 1, Sort: 43},
	{ParentKey: "system", Name: "Roles", MenuType: "menu", Path: "/system/roles", Component: "RolesPage", RouteName: "system-roles", Perms: "view:system-roles", Icon: "safe", Visible: 1, Status: 1, Sort: 44},
	{ParentKey: "system", Name: "Login Logs", MenuType: "menu", Path: "/system/login-logs", Component: "LoginLogsPage", RouteName: "system-login-logs", Perms: "view:system-login-logs", Icon: "history", Visible: 1, Status: 1, Sort: 45},
	{ParentKey: "system", Name: "Settings", MenuType: "menu", Path: "/system/parameters", Component: "SystemModulePage", RouteName: "system-parameters", Perms: "view:system-parameters", Icon: "common", Visible: 1, Status: 1, Sort: 46},
	{ParentKey: "system", Name: "License", MenuType: "menu", Path: "/system/license", Component: "SystemModulePage", RouteName: "system-license", Perms: "view:system-license", Icon: "lock", Visible: 1, Status: 1, Sort: 47},
}

func NewRoleService(db *sql.DB) *RoleService {
	return &RoleService{db: db}
}

func (s *RoleService) EnsureBootstrap(ctx context.Context, adminUsername string) error {
	if err := s.ensureMenus(ctx); err != nil {
		return err
	}
	if err := s.ensureRoles(ctx); err != nil {
		return err
	}
	if err := s.ensureAdminRole(ctx, adminUsername); err != nil {
		return err
	}
	return nil
}

func (s *RoleService) List(ctx context.Context) ([]auth.Role, error) {
	const query = `
		SELECT id, name, role_key, data_scope, sort, status, remark, created_at, updated_at
		FROM sys_role
		WHERE deleted_at IS NULL
		ORDER BY sort ASC, id ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]auth.Role, 0)
	roleIDs := make([]int64, 0)
	for rows.Next() {
		item, scanErr := scanRole(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
		roleIDs = append(roleIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	menuMap, err := s.roleMenuKeys(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].MenuKeys = menuMap[items[index].ID]
	}
	return items, nil
}

func (s *RoleService) Get(ctx context.Context, id int64) (auth.Role, error) {
	const query = `
		SELECT id, name, role_key, data_scope, sort, status, remark, created_at, updated_at
		FROM sys_role
		WHERE id = ? AND deleted_at IS NULL
	`

	row := s.db.QueryRowContext(ctx, query, id)
	item, err := scanRole(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return auth.Role{}, ErrRoleNotFound
		}
		return auth.Role{}, err
	}

	menuMap, err := s.roleMenuKeys(ctx, []int64{id})
	if err != nil {
		return auth.Role{}, err
	}
	item.MenuKeys = menuMap[id]
	return item, nil
}

func (s *RoleService) Create(ctx context.Context, input RoleMutation) (auth.Role, error) {
	normalized, err := normalizeRole(input)
	if err != nil {
		return auth.Role{}, err
	}
	menuKeys, err := s.normalizeMenuKeys(ctx, normalized.MenuKeys)
	if err != nil {
		return auth.Role{}, err
	}

	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO sys_role (name, role_key, data_scope, sort, status, remark) VALUES (?, ?, ?, ?, ?, ?)`,
		normalized.Name,
		normalized.Key,
		normalized.DataScope,
		normalized.Sort,
		normalized.Status,
		normalized.Description,
	)
	if err != nil {
		return auth.Role{}, mapRoleWriteError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return auth.Role{}, err
	}
	if err := s.replaceRoleMenus(ctx, id, menuKeys); err != nil {
		return auth.Role{}, err
	}
	return s.Get(ctx, id)
}

func (s *RoleService) Update(ctx context.Context, id int64, input RoleMutation) (auth.Role, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return auth.Role{}, err
	}

	normalized, err := normalizeRole(input)
	if err != nil {
		return auth.Role{}, err
	}
	menuKeys, err := s.normalizeMenuKeys(ctx, normalized.MenuKeys)
	if err != nil {
		return auth.Role{}, err
	}
	if current.Key == "super_admin" && normalized.Status == 0 {
		return auth.Role{}, ErrProtectedRole
	}

	if _, err := s.db.ExecContext(
		ctx,
		`UPDATE sys_role SET name = ?, role_key = ?, data_scope = ?, sort = ?, status = ?, remark = ? WHERE id = ? AND deleted_at IS NULL`,
		normalized.Name,
		normalized.Key,
		normalized.DataScope,
		normalized.Sort,
		normalized.Status,
		normalized.Description,
		id,
	); err != nil {
		return auth.Role{}, mapRoleWriteError(err)
	}
	if err := s.replaceRoleMenus(ctx, id, menuKeys); err != nil {
		return auth.Role{}, err
	}
	return s.Get(ctx, id)
}

func (s *RoleService) Delete(ctx context.Context, id int64) error {
	item, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if item.Key == "super_admin" {
		return ErrProtectedRole
	}

	if _, err := s.db.ExecContext(ctx, `DELETE FROM sys_role_menu WHERE role_id = ?`, id); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM sys_user_role WHERE role_id = ?`, id); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE sys_role SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, id)
	return err
}

func (s *RoleService) ListMenus(ctx context.Context) ([]auth.Menu, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, parent_id, name, menu_type, path, component, route_name, perms, icon, visible, status, sort
		FROM sys_menu
		WHERE deleted_at IS NULL
		ORDER BY sort ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]auth.Menu, 0)
	for rows.Next() {
		item, scanErr := scanMenu(rows)
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

func (s *RoleService) FindUserAccess(ctx context.Context, userID int64) ([]string, []string, []string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.role_key, r.name
		FROM sys_user_role ur
		INNER JOIN sys_role r ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.deleted_at IS NULL AND r.status = 1
		GROUP BY r.role_key, r.name
		ORDER BY MIN(r.sort) ASC, MIN(r.id) ASC
	`, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	roleKeys := make([]string, 0)
	roleNames := make([]string, 0)
	for rows.Next() {
		var roleKey string
		var roleName string
		if err := rows.Scan(&roleKey, &roleName); err != nil {
			return nil, nil, nil, err
		}
		roleKeys = append(roleKeys, roleKey)
		roleNames = append(roleNames, roleName)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	menuRows, err := s.db.QueryContext(ctx, `
		SELECT m.route_name
		FROM sys_user_role ur
		INNER JOIN sys_role r ON r.id = ur.role_id
		INNER JOIN sys_role_menu rm ON rm.role_id = r.id
		INNER JOIN sys_menu m ON m.id = rm.menu_id
		WHERE ur.user_id = ? AND r.deleted_at IS NULL AND r.status = 1 AND m.deleted_at IS NULL AND m.status = 1 AND m.route_name <> ''
		GROUP BY m.route_name
		ORDER BY MIN(m.sort) ASC, MIN(m.id) ASC
	`, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer menuRows.Close()

	menuKeys := make([]string, 0)
	for menuRows.Next() {
		var routeName string
		if err := menuRows.Scan(&routeName); err != nil {
			return nil, nil, nil, err
		}
		menuKeys = append(menuKeys, routeName)
	}
	if err := menuRows.Err(); err != nil {
		return nil, nil, nil, err
	}

	return roleKeys, roleNames, menuKeys, nil
}

func (s *RoleService) ensureMenus(ctx context.Context) error {
	existing, err := s.ListMenus(ctx)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	existingByRoute := make(map[string]auth.Menu, len(existing))
	for _, item := range existing {
		existingByRoute[item.RouteName] = item
	}

	idsByRoute := make(map[string]int64)
	seedRoutes := make(map[string]struct{}, len(defaultMenus))
	for _, seed := range defaultMenus {
		parentID := int64(0)
		if seed.ParentKey != "" {
			parentID = idsByRoute[seed.ParentKey]
		}
		seedRoutes[seed.RouteName] = struct{}{}

		if current, ok := existingByRoute[seed.RouteName]; ok {
			if _, err := tx.ExecContext(
				ctx,
				`UPDATE sys_menu
				 SET parent_id = ?, name = ?, menu_type = ?, path = ?, component = ?, perms = ?, icon = ?, visible = ?, status = ?, sort = ?, remark = ?
				 WHERE id = ? AND deleted_at IS NULL`,
				parentID,
				seed.Name,
				seed.MenuType,
				seed.Path,
				seed.Component,
				seed.Perms,
				seed.Icon,
				seed.Visible,
				seed.Status,
				seed.Sort,
				seed.Remark,
				current.ID,
			); err != nil {
				return err
			}
			idsByRoute[seed.RouteName] = current.ID
			continue
		}

		result, err := tx.ExecContext(
			ctx,
			`INSERT INTO sys_menu (parent_id, name, menu_type, path, component, route_name, perms, icon, visible, status, sort, remark) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			parentID,
			seed.Name,
			seed.MenuType,
			seed.Path,
			seed.Component,
			seed.RouteName,
			seed.Perms,
			seed.Icon,
			seed.Visible,
			seed.Status,
			seed.Sort,
			seed.Remark,
		)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		idsByRoute[seed.RouteName] = id
	}

	for _, legacyRoute := range []string{"data-governance", "operations", "scan-reports"} {
		if _, stillManaged := seedRoutes[legacyRoute]; stillManaged {
			continue
		}
		if current, ok := existingByRoute[legacyRoute]; ok {
			if _, err := tx.ExecContext(ctx, `UPDATE sys_menu SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, current.ID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *RoleService) ensureRoles(ctx context.Context) error {
	type roleSeed struct {
		Name        string
		Key         string
		DataScope   string
		Sort        int
		Status      int
		Description string
		MenuKeys    []string
	}

	seeds := []roleSeed{
		{
			Name:        "超级管理员",
			Key:         "super_admin",
			DataScope:   "all",
			Sort:        1,
			Status:      1,
			Description: "默认超级管理员角色",
			MenuKeys:    allMenuRouteNames(),
		},
		{
			Name:        "运维管理员",
			Key:         "ops_admin",
			DataScope:   "dept_and_children",
			Sort:        10,
			Status:      1,
			Description: "负责设备巡检、监控和运行维护",
			MenuKeys: []string{
				"overview", "backup-devices", "backup-tasks", "backup-logs", "agents", "groups", "storage", "files", "sync-logs", "monitor", "extraction-rules", "versions",
				"file-logs", "file-permissions", "task-progress", "alert-groups", "alert-logs", "message-channels", "alert-policies",
			},
		},
		{
			Name:        "审计员",
			Key:         "auditor",
			DataScope:   "custom",
			Sort:        20,
			Status:      1,
			Description: "只读访问日志、策略和关键配置",
			MenuKeys: []string{
				"overview", "groups", "files", "sync-logs", "file-logs", "task-progress", "alert-logs", "system-login-logs",
			},
		},
	}

	for _, seed := range seeds {
		var roleID int64
		err := s.db.QueryRowContext(ctx, `SELECT id FROM sys_role WHERE role_key = ? AND deleted_at IS NULL`, seed.Key).Scan(&roleID)
		switch {
		case err == nil:
			if _, err := s.db.ExecContext(ctx, `UPDATE sys_role SET name = ?, data_scope = ?, sort = ?, status = ?, remark = ? WHERE id = ?`, seed.Name, seed.DataScope, seed.Sort, seed.Status, seed.Description, roleID); err != nil {
				return err
			}
			if err := s.replaceRoleMenus(ctx, roleID, seed.MenuKeys); err != nil {
				return err
			}
		case errors.Is(err, sql.ErrNoRows):
			role, err := s.Create(ctx, RoleMutation{
				Name:        seed.Name,
				Key:         seed.Key,
				DataScope:   seed.DataScope,
				Sort:        seed.Sort,
				Status:      seed.Status,
				Description: seed.Description,
				MenuKeys:    seed.MenuKeys,
			})
			if err != nil {
				return err
			}
			roleID = role.ID
		default:
			return err
		}
	}

	return nil
}

func (s *RoleService) ensureAdminRole(ctx context.Context, adminUsername string) error {
	if strings.TrimSpace(adminUsername) == "" {
		return nil
	}

	var userID int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM sys_user WHERE username = ? AND deleted_at IS NULL`, strings.TrimSpace(adminUsername)).Scan(&userID); err != nil {
		return err
	}

	var roleID int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM sys_role WHERE role_key = 'super_admin' AND deleted_at IS NULL`).Scan(&roleID); err != nil {
		return err
	}

	_, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO sys_user_role (user_id, role_id) VALUES (?, ?)`, userID, roleID)
	return err
}

func (s *RoleService) roleMenuKeys(ctx context.Context, roleIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string)
	if len(roleIDs) == 0 {
		return result, nil
	}

	query, args := buildInt64InQuery(`
		SELECT rm.role_id, m.route_name
		FROM sys_role_menu rm
		INNER JOIN sys_menu m ON m.id = rm.menu_id
		WHERE rm.role_id IN (%s) AND m.deleted_at IS NULL AND m.route_name <> ''
		ORDER BY m.sort ASC, m.id ASC
	`, roleIDs)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleID int64
		var routeName string
		if err := rows.Scan(&roleID, &routeName); err != nil {
			return nil, err
		}
		result[roleID] = append(result[roleID], routeName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *RoleService) normalizeMenuKeys(ctx context.Context, raw []string) ([]string, error) {
	keys := uniqueStrings(raw)
	if len(keys) == 0 {
		return keys, nil
	}

	query, args := buildStringInQuery(`
		SELECT route_name
		FROM sys_menu
		WHERE route_name IN (%s) AND deleted_at IS NULL AND status = 1
	`, keys)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	allowed := make(map[string]struct{})
	for rows.Next() {
		var routeName string
		if err := rows.Scan(&routeName); err != nil {
			return nil, err
		}
		allowed[routeName] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, ok := allowed[key]; ok {
			filtered = append(filtered, key)
		}
	}
	return filtered, nil
}

func (s *RoleService) replaceRoleMenus(ctx context.Context, roleID int64, menuKeys []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM sys_role_menu WHERE role_id = ?`, roleID); err != nil {
		return err
	}

	if len(menuKeys) > 0 {
		query, args := buildStringInQuery(`
			SELECT id
			FROM sys_menu
			WHERE route_name IN (%s) AND deleted_at IS NULL AND status = 1
			ORDER BY sort ASC, id ASC
		`, menuKeys)
		rows, err := tx.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		menuIDs := make([]int64, 0, len(menuKeys))

		for rows.Next() {
			var menuID int64
			if err := rows.Scan(&menuID); err != nil {
				_ = rows.Close()
				return err
			}
			menuIDs = append(menuIDs, menuID)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}

		for _, menuID := range menuIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO sys_role_menu (role_id, menu_id) VALUES (?, ?)`, roleID, menuID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func normalizeRole(input RoleMutation) (RoleMutation, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return RoleMutation{}, ErrInvalidRoleName
	}

	key := strings.TrimSpace(input.Key)
	if key == "" {
		return RoleMutation{}, ErrInvalidRoleKey
	}

	dataScope := strings.TrimSpace(input.DataScope)
	if dataScope == "" {
		dataScope = "custom"
	}

	status := 0
	if input.Status != 0 {
		status = 1
	}

	return RoleMutation{
		Name:        name,
		Key:         key,
		DataScope:   dataScope,
		Sort:        input.Sort,
		Status:      status,
		Description: strings.TrimSpace(input.Description),
		MenuKeys:    uniqueStrings(input.MenuKeys),
	}, nil
}

func scanRole(row scanner) (auth.Role, error) {
	var item auth.Role
	var createdAt time.Time
	var updatedAt time.Time
	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Key,
		&item.DataScope,
		&item.Sort,
		&item.Status,
		&item.Description,
		&createdAt,
		&updatedAt,
	); err != nil {
		return auth.Role{}, err
	}
	item.CreatedAt = createdAt.Format(mysqlTimeFormat)
	item.UpdatedAt = updatedAt.Format(mysqlTimeFormat)
	return item, nil
}

func scanMenu(row scanner) (auth.Menu, error) {
	var item auth.Menu
	if err := row.Scan(
		&item.ID,
		&item.ParentID,
		&item.Name,
		&item.MenuType,
		&item.Path,
		&item.Component,
		&item.RouteName,
		&item.Perms,
		&item.Icon,
		&item.Visible,
		&item.Status,
		&item.Sort,
	); err != nil {
		return auth.Menu{}, err
	}
	return item, nil
}

func mapRoleWriteError(err error) error {
	switch mysqlDuplicateField(err) {
	case "role_key":
		return ErrRoleKeyExists
	case "role_name":
		return ErrRoleNameExists
	default:
		return err
	}
}

func ParseRoleID(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("role id is required")
	}

	var id int64
	if _, err := fmt.Sscanf(raw, "%d", &id); err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid role id")
	}
	return id, nil
}

func allMenuRouteNames() []string {
	keys := make([]string, 0, len(defaultMenus))
	for _, item := range defaultMenus {
		if item.RouteName != "" {
			keys = append(keys, item.RouteName)
		}
	}
	return keys
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{})
	items := make([]string, 0, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
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

func buildStringInQuery(pattern string, values []string) (string, []any) {
	holders := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for _, value := range values {
		holders = append(holders, "?")
		args = append(args, value)
	}
	return fmt.Sprintf(pattern, strings.Join(holders, ",")), args
}

func buildInt64InQuery(pattern string, values []int64) (string, []any) {
	holders := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for _, value := range values {
		holders = append(holders, "?")
		args = append(args, value)
	}
	return fmt.Sprintf(pattern, strings.Join(holders, ",")), args
}
