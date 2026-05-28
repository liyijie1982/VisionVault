package http

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"skybase/internal/domain/agent"
	"skybase/internal/domain/auth"
	"skybase/internal/service"
	"skybase/pkg/response"
)

type Server struct {
	meta  *service.MetaService
	agent *service.AgentService
	ops   *service.OpsService
	auth  *service.AuthService
	role  *service.RoleService
	log   *service.LoginLogService
	dept  *service.DepartmentService
	user  *service.UserService
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UUID     string `json:"uuid"`
	Code     string `json:"code"`
}

type passwordResetRequest struct {
	Password string `json:"password"`
}

func NewServer(
	meta *service.MetaService,
	agentService *service.AgentService,
	opsService *service.OpsService,
	authService *service.AuthService,
	roleService *service.RoleService,
	loginLogService *service.LoginLogService,
	departmentService *service.DepartmentService,
	userService *service.UserService,
) *Server {
	return &Server{
		meta:  meta,
		agent: agentService,
		ops:   opsService,
		auth:  authService,
		role:  roleService,
		log:   loginLogService,
		dept:  departmentService,
		user:  userService,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/v1/auth/captcha", s.handleCaptcha)
	mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("/api/v1/auth/me", s.handleMe)
	mux.HandleFunc("/api/v1/auth/logout", s.handleLogout)
	mux.HandleFunc("/api/v1/auth/change-password", s.handleChangePassword)
	mux.HandleFunc("/api/v1/meta/modules", s.handleModules)
	mux.HandleFunc("/api/v1/agents", s.handleAgents)
	mux.HandleFunc("/api/v1/agents/", s.handleAgentByID)
	mux.HandleFunc("/api/v1/agent-directories/", s.handleAgentDirectories)
	mux.HandleFunc("/api/v1/groups", s.handleGroups)
	mux.HandleFunc("/api/v1/groups/", s.handleGroupByID)
	mux.HandleFunc("/api/v1/storage", s.handleStorage)
	mux.HandleFunc("/api/v1/storage/", s.handleStorageByID)
	mux.HandleFunc("/api/v1/files", s.handleFiles)
	mux.HandleFunc("/api/v1/files/", s.handleFileByID)
	mux.HandleFunc("/api/v1/files/upload", s.handleFileUpload)
	mux.HandleFunc("/api/v1/sync-logs", s.handleSyncLogs)
	mux.HandleFunc("/api/v1/scan-reports", s.handleScanReports)
	mux.HandleFunc("/api/v1/versions", s.handleVersions)
	mux.HandleFunc("/api/v1/versions/upload", s.handleVersionUpload)
	mux.HandleFunc("/api/v1/versions/", s.handleVersionActionByID)
	mux.HandleFunc("/api/v1/roles", s.handleRoles)
	mux.HandleFunc("/api/v1/roles/", s.handleRoleByID)
	mux.HandleFunc("/api/v1/menus", s.handleMenus)
	mux.HandleFunc("/api/v1/login-logs", s.handleLoginLogs)
	mux.HandleFunc("/api/v1/departments", s.handleDepartments)
	mux.HandleFunc("/api/v1/departments/", s.handleDepartmentByID)
	mux.HandleFunc("/api/v1/users", s.handleUsers)
	mux.HandleFunc("/api/v1/users/", s.handleUserByID)
	mux.HandleFunc("/api/v1/file-filters", s.handleFileFilters)
	mux.HandleFunc("/api/v1/file-filters/", s.handleFileFilterByID)
	mux.HandleFunc("/api/v1/regex-rules", s.handleRegexRules)
	mux.HandleFunc("/api/v1/regex-rules/", s.handleRegexRuleByID)
	mux.HandleFunc("/api/v1/image-processors", s.handleImageProcessors)
	mux.HandleFunc("/api/v1/image-processors/", s.handleImageProcessorByID)
	mux.HandleFunc("/api/v1/alert-groups", s.handleAlertGroups)
	mux.HandleFunc("/api/v1/alert-groups/", s.handleAlertGroupByID)
	mux.HandleFunc("/api/v1/message-channels", s.handleMessageChannels)
	mux.HandleFunc("/api/v1/message-channels/", s.handleMessageChannelByID)
	mux.HandleFunc("/api/v1/alert-policies", s.handleAlertPolicies)
	mux.HandleFunc("/api/v1/alert-policies/", s.handleAlertPolicyByID)
	mux.HandleFunc("/api/v1/file-logs", s.handleFileLogs)
	mux.HandleFunc("/api/v1/file-permissions", s.handleFilePermissions)
	mux.HandleFunc("/api/v1/file-permissions/", s.handleFilePermissionByID)
	mux.HandleFunc("/api/v1/task-progress", s.handleTaskProgress)
	mux.HandleFunc("/api/v1/task-progress/", s.handleTaskProgressByID)
	mux.HandleFunc("/api/v1/alert-logs", s.handleAlertLogs)
	mux.HandleFunc("/api/v1/system-configs", s.handleSystemConfigs)
	mux.HandleFunc("/api/v1/system-configs/", s.handleSystemConfigByID)
	mux.HandleFunc("/api/v1/licenses", s.handleLicenses)
	mux.HandleFunc("/api/v1/licenses/", s.handleLicenseByID)

	// SkyBase 现已将 Agent 开放接口统一收敛到 /sky 前缀。
	mux.HandleFunc("/sky/agent/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("/sky/agent/commit", s.handleCommit)
	mux.HandleFunc("/sky/agent/scan/commit", s.handleScanCommit)
	mux.HandleFunc("/sky/agent/version", s.handleVersion)
	mux.HandleFunc("/sky/agent/download", s.handleDownload)

	return s.withCORS(mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{
		"name":    s.meta.Health().Name,
		"service": "skybase",
		"status":  "bootstrapped",
	}))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusOK, response.Success(s.meta.Health()))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid login payload"))
		return
	}

	loginIP := clientIP(r)
	session, err := s.auth.Login(strings.TrimSpace(req.Username), req.Password, req.UUID, req.Code, loginIP)
	if err != nil {
		s.recordLoginAttempt(r, service.LoginLogRecordInput{
			Username:    strings.TrimSpace(req.Username),
			LoginIP:     loginIP,
			UserAgent:   r.UserAgent(),
			LoginStatus: 0,
			Message:     err.Error(),
		})
		if errors.Is(err, service.ErrInvalidVerificationCode) {
			response.WriteJSON(w, http.StatusUnauthorized, response.Error(http.StatusUnauthorized, err.Error()))
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.WriteJSON(w, http.StatusUnauthorized, response.Error(http.StatusUnauthorized, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	s.recordLoginAttempt(r, service.LoginLogRecordInput{
		UserID:      session.User.ID,
		Username:    session.User.Username,
		LoginIP:     loginIP,
		UserAgent:   r.UserAgent(),
		LoginStatus: 1,
		Message:     "login succeeded",
	})
	s.setSessionCookie(w, session.Token)
	response.WriteJSON(w, http.StatusOK, response.Success(service.LoginResult{User: session.User}))
}

func (s *Server) handleCaptcha(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	captcha, err := s.auth.GenerateCaptcha()
	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(captcha))
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	user, err := s.requireUser(r)
	if err != nil {
		s.writeUnauthorized(w)
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(user))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	s.auth.Logout(s.sessionToken(r))
	s.clearSessionCookie(w)
	response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{
		"signedOut": true,
	}))
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	var req service.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid password payload"))
		return
	}

	user, err := s.auth.ChangePassword(s.sessionToken(r), req)
	if err != nil {
		s.writePasswordError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(user))
}

func (s *Server) handleModules(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}

	health := s.meta.Health()
	response.WriteJSON(w, http.StatusOK, response.Success(health.Modules))
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		response.WriteJSON(w, http.StatusOK, response.Success(s.agent.ListAgents()))
	case http.MethodPost:
		var req service.AgentMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid agent payload"))
			return
		}
		item, err := s.agent.CreateAgent(req)
		if err != nil {
			s.writeAgentError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		response.WriteJSON(w, http.StatusOK, response.Success(s.agent.ListGroups()))
	case http.MethodPost:
		var req service.GroupMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid group payload"))
			return
		}
		item, err := s.agent.CreateGroup(req)
		if err != nil {
			s.writeGroupError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleStorage(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		response.WriteJSON(w, http.StatusOK, response.Success(s.agent.ListStorageTargets()))
	case http.MethodPost:
		var req service.StorageMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid storage payload"))
			return
		}
		item, err := s.agent.CreateStorageTarget(req)
		if err != nil {
			s.writeStorageError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		response.WriteJSON(w, http.StatusOK, response.Success(s.agent.ListFiles()))
	case http.MethodPost:
		var req service.FileMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid file payload"))
			return
		}
		item, err := s.agent.CreateFile(req)
		if err != nil {
			s.writeFileError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleAgentByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/agents/"), "agent")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := s.agent.GetAgent(id)
		if err != nil {
			s.writeAgentError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodPut:
		var req service.AgentMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid agent payload"))
			return
		}
		item, err := s.agent.UpdateAgent(id, req)
		if err != nil {
			s.writeAgentError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.agent.DeleteAgent(id); err != nil {
			s.writeAgentError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleAgentDirectories(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/agent-directories/"), "agent")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}

	items, err := s.agent.BrowseAgentDirectories(r.Context(), id, r.URL.Query().Get("path"))
	if err != nil {
		s.writeAgentError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Success(items))
}

func (s *Server) handleGroupByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/groups/"), "group")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := s.agent.GetGroup(id)
		if err != nil {
			s.writeGroupError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodPut:
		var req service.GroupMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid group payload"))
			return
		}
		item, err := s.agent.UpdateGroup(id, req)
		if err != nil {
			s.writeGroupError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.agent.DeleteGroup(id); err != nil {
			s.writeGroupError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleStorageByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/storage/"), "storage")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := s.agent.GetStorageTarget(id)
		if err != nil {
			s.writeStorageError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodPut:
		var req service.StorageMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid storage payload"))
			return
		}
		item, err := s.agent.UpdateStorageTarget(id, req)
		if err != nil {
			s.writeStorageError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.agent.DeleteStorageTarget(id); err != nil {
			s.writeStorageError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleFileByID(w http.ResponseWriter, r *http.Request) {
	user, err := s.requireUser(r)
	if err != nil {
		s.writeUnauthorized(w)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/files/")
	if strings.HasSuffix(path, "/download") {
		s.handleFileDownload(w, r, user, strings.TrimSuffix(path, "/download"))
		return
	}
	id, err := service.ParseStringID(path, "file")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := s.agent.GetFile(id)
		if err != nil {
			s.writeFileError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodPut:
		var req service.FileMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid file payload"))
			return
		}
		item, err := s.agent.UpdateFile(id, req)
		if err != nil {
			s.writeFileError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.agent.DeleteFile(id); err != nil {
			s.writeFileError(w, err)
			return
		}
		_ = s.ops.LogFileOperation(r.Context(), service.FileLogRecord{
			UserID:        user.ID,
			Username:      user.Username,
			FilePath:      id,
			OperationType: "delete",
			ResultStatus:  "success",
			ClientIP:      clientIP(r),
			Message:       "file deleted",
		})
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleFileUpload(w http.ResponseWriter, r *http.Request) {
	user, err := s.requireUser(r)
	if err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodPost {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid multipart payload"))
		return
	}
	storageID, parseErr := strconv.ParseInt(strings.TrimSpace(r.FormValue("storageId")), 10, 64)
	if parseErr != nil || storageID <= 0 {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "storageId is required"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "file is required"))
		return
	}
	defer file.Close()

	tags := strings.Split(r.FormValue("tags"), ",")
	item, err := s.agent.UploadFile(storageID, header.Filename, file, tags)
	if err != nil {
		s.writeFileError(w, err)
		return
	}
	_ = s.ops.LogFileOperation(r.Context(), service.FileLogRecord{
		UserID:        user.ID,
		Username:      user.Username,
		StorageID:     item.StorageID,
		StorageName:   item.Storage,
		FilePath:      item.Path,
		OperationType: "upload",
		ResultStatus:  "success",
		ClientIP:      clientIP(r),
		Message:       "file uploaded",
	})
	response.WriteJSON(w, http.StatusOK, response.Success(item))
}

func (s *Server) handleFileDownload(w http.ResponseWriter, r *http.Request, user auth.User, rawID string) {
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}
	id, err := service.ParseStringID(rawID, "file")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	filePath, contentType, err := s.agent.DownloadFile(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, service.ErrFileNotFound) {
			http.NotFound(w, r)
			return
		}
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	_ = s.ops.LogFileOperation(r.Context(), service.FileLogRecord{
		UserID:        user.ID,
		Username:      user.Username,
		FilePath:      filePath,
		OperationType: "download",
		ResultStatus:  "success",
		ClientIP:      clientIP(r),
		Message:       "file downloaded",
	})
	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, filePath)
}

func (s *Server) handleSyncLogs(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	query, err := parseSyncLogQuery(r)
	if err != nil {
		s.writeUserError(w, err)
		return
	}

	items, err := s.agent.ListSyncLogs(r.Context(), query)
	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(items))
}

func (s *Server) handleScanReports(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(s.agent.ListScanReports()))
}

func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	items, err := s.agent.ListVersions()
	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(items))
}

func (s *Server) handleVersionUpload(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodPost {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}
	if err := r.ParseMultipartForm(256 << 20); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid multipart payload"))
		return
	}
	version := strings.TrimSpace(r.FormValue("version"))
	releaseNotes := strings.TrimSpace(r.FormValue("releaseNotes"))
	activate := strings.EqualFold(strings.TrimSpace(r.FormValue("activate")), "true") || strings.TrimSpace(r.FormValue("activate")) == "1"
	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "package file is required"))
		return
	}
	defer file.Close()

	item, err := s.agent.UploadVersion(version, releaseNotes, activate, header.Filename, file)
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Success(item))
}

func (s *Server) handleVersionActionByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/versions/")
	if strings.HasSuffix(path, "/verify-md5") {
		s.handleVersionVerifyMD5(w, r, strings.TrimSuffix(path, "/verify-md5"))
		return
	}
	response.WriteJSON(w, http.StatusNotFound, response.Error(http.StatusNotFound, "not found"))
}

func (s *Server) handleVersionVerifyMD5(w http.ResponseWriter, r *http.Request, rawID string) {
	if r.Method != http.MethodPost {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}
	id, err := service.ParseStringID(rawID, "version")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	result, err := s.agent.VerifyVersionMD5(id)
	if err != nil {
		if errors.Is(err, service.ErrFileNotFound) || errors.Is(err, os.ErrNotExist) {
			response.WriteJSON(w, http.StatusNotFound, response.Error(http.StatusNotFound, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Success(result))
}

func (s *Server) handleRoles(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := s.role.List(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.RoleMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid role payload"))
			return
		}

		item, err := s.role.Create(r.Context(), req)
		if err != nil {
			s.writeRoleError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleRoleByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}

	id, err := service.ParseRoleID(strings.TrimPrefix(r.URL.Path, "/api/v1/roles/"))
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, getErr := s.role.Get(r.Context(), id)
		if getErr != nil {
			s.writeRoleError(w, getErr)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodPut:
		var req service.RoleMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid role payload"))
			return
		}

		item, updateErr := s.role.Update(r.Context(), id, req)
		if updateErr != nil {
			s.writeRoleError(w, updateErr)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if deleteErr := s.role.Delete(r.Context(), id); deleteErr != nil {
			s.writeRoleError(w, deleteErr)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{
			"deleted": true,
		}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleMenus(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	items, err := s.role.ListMenus(r.Context())
	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Success(items))
}

func (s *Server) handleLoginLogs(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	query, err := parseLoginLogQuery(r)
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}

	items, err := s.log.List(r.Context(), query)
	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(items))
}

func (s *Server) handleDepartments(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := s.dept.List(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.DepartmentMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid department payload"))
			return
		}

		item, err := s.dept.Create(r.Context(), req)
		if err != nil {
			s.writeDepartmentError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleDepartmentByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}

	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/departments/"), "department")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, getErr := s.dept.Get(r.Context(), id)
		if getErr != nil {
			s.writeDepartmentError(w, getErr)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodPut:
		var req service.DepartmentMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid department payload"))
			return
		}

		item, updateErr := s.dept.Update(r.Context(), id, req)
		if updateErr != nil {
			s.writeDepartmentError(w, updateErr)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if deleteErr := s.dept.Delete(r.Context(), id); deleteErr != nil {
			s.writeDepartmentError(w, deleteErr)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{
			"deleted": true,
		}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := s.user.List(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.UserMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid user payload"))
			return
		}

		item, err := s.user.Create(r.Context(), req)
		if err != nil {
			s.writeUserError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	if strings.HasSuffix(path, "/reset-password") {
		s.handleUserResetPassword(w, r, strings.TrimSuffix(path, "/reset-password"))
		return
	}

	id, err := service.ParseEntityID(path, "user")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, getErr := s.user.Get(r.Context(), id)
		if getErr != nil {
			s.writeUserError(w, getErr)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodPut:
		var req service.UserMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid user payload"))
			return
		}

		item, updateErr := s.user.Update(r.Context(), id, req)
		if updateErr != nil {
			s.writeUserError(w, updateErr)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if deleteErr := s.user.Delete(r.Context(), id); deleteErr != nil {
			s.writeUserError(w, deleteErr)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{
			"deleted": true,
		}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleUserResetPassword(w http.ResponseWriter, r *http.Request, rawID string) {
	if r.Method != http.MethodPost {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	id, err := service.ParseEntityID(rawID, "user")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}

	var req passwordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid password payload"))
		return
	}

	if err := s.user.ResetPassword(r.Context(), id, req.Password); err != nil {
		s.writePasswordError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{
		"reset": true,
	}))
}

func (s *Server) handleFileFilters(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListFileFilters(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.FileFilterMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid file filter payload"))
			return
		}
		item, err := s.ops.CreateFileFilter(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleFileFilterByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/file-filters/"), "file filter")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.FileFilterMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid file filter payload"))
			return
		}
		item, err := s.ops.UpdateFileFilter(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteFileFilter(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleRegexRules(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListRegexRules(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.RegexRuleMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid regex rule payload"))
			return
		}
		item, err := s.ops.CreateRegexRule(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleRegexRuleByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/regex-rules/"), "regex rule")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.RegexRuleMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid regex rule payload"))
			return
		}
		item, err := s.ops.UpdateRegexRule(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteRegexRule(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleImageProcessors(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListImageProcessors(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.ImageProcessorMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid image processor payload"))
			return
		}
		item, err := s.ops.CreateImageProcessor(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleImageProcessorByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/image-processors/"), "image processor")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.ImageProcessorMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid image processor payload"))
			return
		}
		item, err := s.ops.UpdateImageProcessor(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteImageProcessor(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleAlertGroups(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListAlertGroups(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.AlertGroupMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid alert group payload"))
			return
		}
		item, err := s.ops.CreateAlertGroup(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleAlertGroupByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/alert-groups/"), "alert group")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.AlertGroupMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid alert group payload"))
			return
		}
		item, err := s.ops.UpdateAlertGroup(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteAlertGroup(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleMessageChannels(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListMessageChannels(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.MessageChannelMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid message channel payload"))
			return
		}
		item, err := s.ops.CreateMessageChannel(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleMessageChannelByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/message-channels/"), "message channel")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.MessageChannelMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid message channel payload"))
			return
		}
		item, err := s.ops.UpdateMessageChannel(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteMessageChannel(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleAlertPolicies(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListAlertPolicies(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.AlertPolicyMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid alert policy payload"))
			return
		}
		item, err := s.ops.CreateAlertPolicy(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleAlertPolicyByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/alert-policies/"), "alert policy")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.AlertPolicyMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid alert policy payload"))
			return
		}
		item, err := s.ops.UpdateAlertPolicy(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteAlertPolicy(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleFileLogs(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}
	items, err := s.ops.ListFileLogs(r.Context())
	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Success(items))
}

func (s *Server) handleFilePermissions(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListFilePermissions(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.FilePermissionMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid file permission payload"))
			return
		}
		item, err := s.ops.CreateFilePermission(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleFilePermissionByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/file-permissions/"), "file permission")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.FilePermissionMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid file permission payload"))
			return
		}
		item, err := s.ops.UpdateFilePermission(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteFilePermission(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleTaskProgress(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListTaskProgress(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.TaskProgressMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid task progress payload"))
			return
		}
		item, err := s.ops.CreateTaskProgress(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleTaskProgressByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/task-progress/"), "task progress")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.TaskProgressMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid task progress payload"))
			return
		}
		item, err := s.ops.UpdateTaskProgress(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteTaskProgress(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleAlertLogs(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}
	items, err := s.ops.ListAlertLogs(r.Context())
	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}
	response.WriteJSON(w, http.StatusOK, response.Success(items))
}

func (s *Server) handleSystemConfigs(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListSystemConfigs(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.SystemConfigMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid system config payload"))
			return
		}
		item, err := s.ops.CreateSystemConfig(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleSystemConfigByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/system-configs/"), "system config")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.SystemConfigMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid system config payload"))
			return
		}
		item, err := s.ops.UpdateSystemConfig(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteSystemConfig(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleLicenses(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := s.ops.ListLicenses(r.Context())
		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(items))
	case http.MethodPost:
		var req service.LicenseMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid license payload"))
			return
		}
		item, err := s.ops.CreateLicense(r.Context(), req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleLicenseByID(w http.ResponseWriter, r *http.Request) {
	if _, err := s.requireUser(r); err != nil {
		s.writeUnauthorized(w)
		return
	}
	id, err := service.ParseEntityID(strings.TrimPrefix(r.URL.Path, "/api/v1/licenses/"), "license")
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req service.LicenseMutation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid license payload"))
			return
		}
		item, err := s.ops.UpdateLicense(r.Context(), id, req)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(item))
	case http.MethodDelete:
		if err := s.ops.DeleteLicense(r.Context(), id); err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
			return
		}
		response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{"deleted": true}))
	default:
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
	}
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	var req agent.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid heartbeat payload"))
		return
	}

	policy := s.agent.Heartbeat(req)
	response.WriteJSON(w, http.StatusOK, response.Success(policy))
}

func (s *Server) handleCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	values := map[string]string{
		"ip":            r.URL.Query().Get("ip"),
		"path":          r.URL.Query().Get("path"),
		"fileSize":      r.URL.Query().Get("fileSize"),
		"fileCount":     r.URL.Query().Get("fileCount"),
		"errCount":      r.URL.Query().Get("errCount"),
		"taskStartTime": r.URL.Query().Get("taskStartTime"),
		"tarListPath":   r.URL.Query().Get("tarListPath"),
	}
	if err := s.agent.RecordCommit(r.Context(), values); err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{
		"accepted": true,
	}))
}

func (s *Server) handleScanCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid scan payload"))
		return
	}
	s.agent.RecordScanCommit(body)

	response.WriteJSON(w, http.StatusOK, response.Success(map[string]any{
		"accepted": true,
	}))
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	versionPayload, err := s.agent.Version(r.URL.Query().Get("version"))
	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	response.WriteJSON(w, http.StatusOK, response.Success(versionPayload))
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.WriteJSON(w, http.StatusMethodNotAllowed, response.Error(http.StatusMethodNotAllowed, "method not allowed"))
		return
	}

	filePath, contentType, err := s.agent.Download(r.URL.Query().Get("id"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || err.Error() == "download package not found" {
			http.NotFound(w, r)
			return
		}
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
		return
	}

	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, filePath)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireUser(r *http.Request) (auth.User, error) {
	return s.auth.CurrentUser(s.sessionToken(r))
}

func (s *Server) sessionToken(r *http.Request) string {
	cookie, err := r.Cookie(service.SessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     service.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(s.auth.SessionTTL()),
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     service.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (s *Server) writeUnauthorized(w http.ResponseWriter) {
	response.WriteJSON(w, http.StatusUnauthorized, response.Error(http.StatusUnauthorized, service.ErrUnauthorized.Error()))
}

func (s *Server) writeAgentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAgentNotFound):
		response.WriteJSON(w, http.StatusNotFound, response.Error(http.StatusNotFound, err.Error()))
	case errors.Is(err, service.ErrAgentDirectoryUnavailable):
		response.WriteJSON(w, http.StatusBadGateway, response.Error(http.StatusBadGateway, err.Error()))
	case errors.Is(err, service.ErrAgentHostSNRequired),
		errors.Is(err, service.ErrAgentHostNameRequired),
		errors.Is(err, service.ErrAgentIPRequired),
		errors.Is(err, service.ErrAgentGroupNotFound),
		errors.Is(err, service.ErrAgentStorageNotFound),
		errors.Is(err, service.ErrAgentHostSNDuplicate),
		errors.Is(err, service.ErrAgentIPDuplicate):
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
	default:
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
	}
}

func (s *Server) writeGroupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrGroupNotFound):
		response.WriteJSON(w, http.StatusNotFound, response.Error(http.StatusNotFound, err.Error()))
	case errors.Is(err, service.ErrGroupNameRequired),
		errors.Is(err, service.ErrGroupStorageNotFound),
		errors.Is(err, service.ErrGroupWorkWindowsRequired),
		errors.Is(err, service.ErrGroupReferencedByAgent):
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
	default:
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
	}
}

func (s *Server) writeStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrStorageNotFound):
		response.WriteJSON(w, http.StatusNotFound, response.Error(http.StatusNotFound, err.Error()))
	case errors.Is(err, service.ErrStorageNameRequired),
		errors.Is(err, service.ErrStorageNameExists),
		errors.Is(err, service.ErrStorageTypeInvalid),
		errors.Is(err, service.ErrStoragePathRequired),
		errors.Is(err, service.ErrStorageEndpointInvalid),
		errors.Is(err, service.ErrStorageReferenced):
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
	default:
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
	}
}

func (s *Server) writeFileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrFileNotFound):
		response.WriteJSON(w, http.StatusNotFound, response.Error(http.StatusNotFound, err.Error()))
	case errors.Is(err, service.ErrFileNameRequired),
		errors.Is(err, service.ErrFilePathRequired),
		errors.Is(err, service.ErrFileStorageNotFound):
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
	default:
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
	}
}

func (s *Server) writeRoleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrRoleNotFound):
		response.WriteJSON(w, http.StatusNotFound, response.Error(http.StatusNotFound, err.Error()))
	case errors.Is(err, service.ErrRoleNameExists),
		errors.Is(err, service.ErrRoleKeyExists),
		errors.Is(err, service.ErrInvalidRoleName),
		errors.Is(err, service.ErrInvalidRoleKey),
		errors.Is(err, service.ErrProtectedRole):
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
	default:
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
	}
}

func (s *Server) writeDepartmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrDepartmentNotFound):
		response.WriteJSON(w, http.StatusNotFound, response.Error(http.StatusNotFound, err.Error()))
	case errors.Is(err, service.ErrDepartmentNameRequired),
		errors.Is(err, service.ErrDepartmentParentNotFound),
		errors.Is(err, service.ErrDepartmentParentInvalid),
		errors.Is(err, service.ErrDepartmentHasChildren),
		errors.Is(err, service.ErrDepartmentHasUsers):
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
	default:
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
	}
}

func (s *Server) writeUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		response.WriteJSON(w, http.StatusNotFound, response.Error(http.StatusNotFound, err.Error()))
	case errors.Is(err, service.ErrUsernameRequired),
		errors.Is(err, service.ErrUserNicknameRequired),
		errors.Is(err, service.ErrUserPhoneRequired),
		errors.Is(err, service.ErrUserDepartmentRequired),
		errors.Is(err, service.ErrUserPasswordRequired),
		errors.Is(err, service.ErrPasswordRequired),
		errors.Is(err, service.ErrPasswordTooShort),
		errors.Is(err, service.ErrUsernameExists),
		errors.Is(err, service.ErrUserPhoneExists):
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
	default:
		response.WriteJSON(w, http.StatusInternalServerError, response.Error(http.StatusInternalServerError, err.Error()))
	}
}

func (s *Server) writePasswordError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUnauthorized):
		response.WriteJSON(w, http.StatusUnauthorized, response.Error(http.StatusUnauthorized, err.Error()))
	case errors.Is(err, service.ErrCurrentPasswordIncorrect):
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
	case errors.Is(err, service.ErrPasswordRequired),
		errors.Is(err, service.ErrPasswordTooShort),
		errors.Is(err, service.ErrUserPasswordRequired):
		response.WriteJSON(w, http.StatusBadRequest, response.Error(http.StatusBadRequest, err.Error()))
	default:
		s.writeUserError(w, err)
	}
}

func (s *Server) recordLoginAttempt(r *http.Request, input service.LoginLogRecordInput) {
	if err := s.log.Record(r.Context(), input); err != nil {
		// 登录主流程不因审计日志落库失败而中断。
	}
}

func parseLoginLogQuery(r *http.Request) (service.LoginLogQuery, error) {
	values := r.URL.Query()
	query := service.LoginLogQuery{
		Username: strings.TrimSpace(values.Get("username")),
		LoginIP:  strings.TrimSpace(values.Get("loginIp")),
		Page:     parseIntOrDefault(values.Get("page"), 1),
		PageSize: parseIntOrDefault(values.Get("pageSize"), 20),
	}

	if rawStatus := strings.TrimSpace(values.Get("loginStatus")); rawStatus != "" {
		status, err := strconv.Atoi(rawStatus)
		if err != nil {
			return service.LoginLogQuery{}, errors.New("invalid loginStatus")
		}
		query.LoginStatus = &status
	}

	startAt, err := parseDateTime(values.Get("startAt"), false)
	if err != nil {
		return service.LoginLogQuery{}, err
	}
	if startAt != nil {
		query.StartAt = startAt
	}

	endAt, err := parseDateTime(values.Get("endAt"), true)
	if err != nil {
		return service.LoginLogQuery{}, err
	}
	if endAt != nil {
		query.EndAt = endAt
	}

	return query, nil
}

func parseSyncLogQuery(r *http.Request) (service.SyncLogQuery, error) {
	result := strings.TrimSpace(r.URL.Query().Get("result"))
	switch result {
	case "", "all", "failed", "success":
		return service.SyncLogQuery{Result: result}, nil
	default:
		return service.SyncLogQuery{}, errors.New("invalid result")
	}
}

func parseIntOrDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseDateTime(value string, endOfDay bool) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err != nil {
			continue
		}
		if layout == "2006-01-02" && endOfDay {
			parsed = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		return &parsed, nil
	}

	return nil, errors.New("invalid datetime format")
}

func clientIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}
		if header == "X-Forwarded-For" {
			value = strings.TrimSpace(strings.Split(value, ",")[0])
		}
		if value != "" {
			return value
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
