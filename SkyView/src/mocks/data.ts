import type {
  AgentRecord,
  DepartmentRecord,
  FileRecord,
  GroupRecord,
  HealthStatus,
  LoginLogRecord,
  MenuRecord,
  ModuleSummary,
  PagedResult,
  RoleRecord,
  StorageRecord,
  UserRecord,
  VersionRecord
} from "../types";

export const mockModules: ModuleSummary[] = [
  { key: "auth", name: "RBAC", description: "Users, roles, menus, and settings", status: "planned" },
  { key: "storage", name: "Storage", description: "Local and S3-compatible storage management", status: "planned" },
  { key: "agent", name: "Agent", description: "Agent lifecycle, heartbeat, and rollout management", status: "planned" },
  { key: "monitor", name: "Monitor", description: "Monitoring, dashboards, and alerting", status: "planned" },
  { key: "jobs", name: "Jobs", description: "Bulk download, indexing, and async workflows", status: "planned" }
];

export const mockHealth: HealthStatus = {
  name: "SkyBase",
  env: "development",
  status: "ok",
  startedAt: "2026-05-11T08:00:00Z",
  now: "2026-05-11T10:00:00Z",
  modules: mockModules
};

export const mockVersions: VersionRecord[] = [
  {
    id: "pkg-20260511",
    version: "1.9.3.3",
    filename: "SyncAgent-install-std-x86-1.9.3.3.exe",
    md5: "38bb2c8d5af8f9f0b3bc9479d48213f0",
    status: "Active",
    updatedAt: "2026-05-11 14:20",
    releaseNotes: "Stability refresh, policy parsing fixes, and startup optimization.",
    agentCount: 12
  },
  {
    id: "pkg-20260415",
    version: "1.9.2.8",
    filename: "SyncAgent-install-std-x86-1.9.2.8.exe",
    md5: "4f934c11a328663e4d7030e567145eaf",
    status: "Archived",
    updatedAt: "2026-04-15 09:00",
    releaseNotes: "Scan report packaging update.",
    agentCount: 3
  }
];

export const mockStorageTargets: StorageRecord[] = [
  {
    id: 1,
    name: "Factory Archive",
    type: "local",
    endpoint: "",
    accessKey: "",
    secretKey: "",
    bucket: "",
    region: "",
    localPath: "/srv/archive/factory-a",
    quota: 12_000_000_000_000,
    status: 1,
    remark: "Primary local archive for line cameras",
    createdAt: "2026-04-01 10:00",
    updatedAt: "2026-05-10 18:05"
  },
  {
    id: 2,
    name: "Cold Object Vault",
    type: "s3",
    endpoint: "https://s3.internal.example.com",
    accessKey: "AKIA******",
    secretKey: "********",
    bucket: "visionvault-cold",
    region: "ap-southeast-1",
    localPath: "",
    quota: 24_000_000_000_000,
    status: 1,
    remark: "Secondary retention target",
    createdAt: "2026-03-18 08:20",
    updatedAt: "2026-05-09 16:30"
  },
  {
    id: 3,
    name: "QA Staging",
    type: "local",
    endpoint: "",
    accessKey: "",
    secretKey: "",
    bucket: "",
    region: "",
    localPath: "/mnt/staging/qa",
    quota: 2_000_000_000_000,
    status: 0,
    remark: "Reserved for regression verification",
    createdAt: "2026-02-11 11:10",
    updatedAt: "2026-05-08 11:10"
  }
];

export const mockGroups: GroupRecord[] = [
  {
    id: 1,
    name: "Assembly Line A",
    storageId: 1,
    ipRange: "10.16.1.0/24",
    pathPrefix: "/assembly-a",
    intervalTime: 300,
    delTimeDays: 30,
    transferSpeedLimit: 6,
    workWindows: [
      { startTime: "08:00", endTime: "12:00" },
      { startTime: "13:30", endTime: "20:00" }
    ],
    fileFilterId: 3,
    regexId: 2,
    imageProcessId: 1,
    alarmGroupId: 1,
    logEnabled: true,
    status: 1,
    createdAt: "2026-03-01 09:00",
    updatedAt: "2026-05-10 09:10"
  },
  {
    id: 2,
    name: "Regional Quality Labs",
    storageId: 2,
    ipRange: "10.18.12.0/24",
    pathPrefix: "/quality-labs",
    intervalTime: 600,
    delTimeDays: 90,
    transferSpeedLimit: 4,
    workWindows: [{ startTime: "00:00", endTime: "23:59" }],
    fileFilterId: 2,
    regexId: 1,
    imageProcessId: 2,
    alarmGroupId: 2,
    logEnabled: false,
    status: 1,
    createdAt: "2026-03-19 15:10",
    updatedAt: "2026-05-07 17:20"
  }
];

export const mockRoles: RoleRecord[] = [
  {
    id: 1,
    name: "超级管理员",
    key: "super_admin",
    dataScope: "all",
    sort: 1,
    status: 1,
    description: "默认超级管理员角色",
    menuKeys: ["overview", "backup-devices", "backup-tasks", "backup-logs", "agents", "groups", "sync-logs", "monitor", "extraction-rules", "versions", "storage", "files", "file-logs", "file-permissions", "task-progress", "alert-groups", "alert-logs", "message-channels", "alert-policies", "system-users", "system-departments", "system-roles", "system-login-logs", "system-parameters", "system-license"],
    createdAt: "2026-05-01 09:00",
    updatedAt: "2026-05-01 09:00"
  },
  {
    id: 2,
    name: "运维管理员",
    key: "ops_admin",
    dataScope: "dept_and_children",
    sort: 10,
    status: 1,
    description: "负责设备巡检、监控和运行维护",
    menuKeys: ["overview", "backup-devices", "backup-tasks", "backup-logs", "agents", "groups", "sync-logs", "monitor", "extraction-rules", "versions", "storage", "files", "file-logs", "file-permissions", "task-progress", "alert-groups", "alert-logs", "message-channels", "alert-policies"],
    createdAt: "2026-05-03 10:15",
    updatedAt: "2026-05-09 14:00"
  },
  {
    id: 3,
    name: "审计员",
    key: "auditor",
    dataScope: "custom",
    sort: 20,
    status: 1,
    description: "只读访问日志、策略和关键配置",
    menuKeys: ["overview", "groups", "files", "sync-logs", "file-logs", "task-progress", "alert-logs", "system-login-logs"],
    createdAt: "2026-05-04 11:20",
    updatedAt: "2026-05-08 18:30"
  }
];

export const mockMenus: MenuRecord[] = [
  { id: 1, parentId: 0, name: "Overview", menuType: "menu", path: "/", component: "OverviewPage", routeName: "overview", perms: "view:overview", icon: "dashboard", visible: 1, status: 1, sort: 1 },
  { id: 28, parentId: 0, name: "备份", menuType: "directory", path: "", component: "", routeName: "backup", perms: "", icon: "archive", visible: 1, status: 1, sort: 5 },
  { id: 29, parentId: 28, name: "磁带设备", menuType: "menu", path: "/backup/devices", component: "BackupDevicesPage", routeName: "backup-devices", perms: "view:backup-devices", icon: "storage", visible: 1, status: 1, sort: 6 },
  { id: 30, parentId: 28, name: "任务管理", menuType: "menu", path: "/backup/tasks", component: "BackupTasksPage", routeName: "backup-tasks", perms: "view:backup-tasks", icon: "list", visible: 1, status: 1, sort: 7 },
  { id: 31, parentId: 28, name: "备份日志", menuType: "menu", path: "/backup/logs", component: "BackupLogsPage", routeName: "backup-logs", perms: "view:backup-logs", icon: "history", visible: 1, status: 1, sort: 8 },
  { id: 2, parentId: 0, name: "Agent Control", menuType: "directory", path: "", component: "", routeName: "agent-control", perms: "", icon: "computer", visible: 1, status: 1, sort: 10 },
  { id: 3, parentId: 2, name: "Agents", menuType: "menu", path: "/agents", component: "AgentsPage", routeName: "agents", perms: "view:agents", icon: "computer", visible: 1, status: 1, sort: 11 },
  { id: 4, parentId: 2, name: "Groups", menuType: "menu", path: "/groups", component: "GroupsPage", routeName: "groups", perms: "view:groups", icon: "apps", visible: 1, status: 1, sort: 12 },
  { id: 5, parentId: 2, name: "Sync Logs", menuType: "menu", path: "/sync-logs", component: "SyncLogsPage", routeName: "sync-logs", perms: "view:sync-logs", icon: "list", visible: 1, status: 1, sort: 13 },
  { id: 6, parentId: 2, name: "Monitor", menuType: "menu", path: "/monitor", component: "MonitorPage", routeName: "monitor", perms: "view:monitor", icon: "command", visible: 1, status: 1, sort: 14 },
  { id: 7, parentId: 2, name: "Extraction Rules", menuType: "menu", path: "/extraction-rules", component: "SystemModulePage", routeName: "extraction-rules", perms: "view:extraction-rules", icon: "common", visible: 1, status: 1, sort: 15 },
  { id: 8, parentId: 2, name: "Versions", menuType: "menu", path: "/versions", component: "VersionsPage", routeName: "versions", perms: "view:versions", icon: "archive", visible: 1, status: 1, sort: 16 },
  { id: 9, parentId: 0, name: "File Control", menuType: "directory", path: "", component: "", routeName: "file-control", perms: "", icon: "folder", visible: 1, status: 1, sort: 20 },
  { id: 10, parentId: 9, name: "Files", menuType: "menu", path: "/files", component: "FilesPage", routeName: "files", perms: "view:files", icon: "folder", visible: 1, status: 1, sort: 21 },
  { id: 11, parentId: 9, name: "File Logs", menuType: "menu", path: "/file-logs", component: "SystemModulePage", routeName: "file-logs", perms: "view:file-logs", icon: "history", visible: 1, status: 1, sort: 22 },
  { id: 12, parentId: 9, name: "File Permissions", menuType: "menu", path: "/file-permissions", component: "SystemModulePage", routeName: "file-permissions", perms: "view:file-permissions", icon: "lock", visible: 1, status: 1, sort: 23 },
  { id: 13, parentId: 9, name: "Task Progress", menuType: "menu", path: "/task-progress", component: "SystemModulePage", routeName: "task-progress", perms: "view:task-progress", icon: "list", visible: 1, status: 1, sort: 24 },
  { id: 15, parentId: 0, name: "Alerts", menuType: "directory", path: "", component: "", routeName: "alerts", perms: "", icon: "safe", visible: 1, status: 1, sort: 30 },
  { id: 16, parentId: 15, name: "Alert Groups", menuType: "menu", path: "/alerts/groups", component: "SystemModulePage", routeName: "alert-groups", perms: "view:alert-groups", icon: "apps", visible: 1, status: 1, sort: 31 },
  { id: 17, parentId: 15, name: "Alert Logs", menuType: "menu", path: "/alerts/logs", component: "SystemModulePage", routeName: "alert-logs", perms: "view:alert-logs", icon: "history", visible: 1, status: 1, sort: 32 },
  { id: 18, parentId: 15, name: "Message Channels", menuType: "menu", path: "/alerts/channels", component: "SystemModulePage", routeName: "message-channels", perms: "view:message-channels", icon: "common", visible: 1, status: 1, sort: 33 },
  { id: 19, parentId: 15, name: "Alert Policies", menuType: "menu", path: "/alerts/policies", component: "SystemModulePage", routeName: "alert-policies", perms: "view:alert-policies", icon: "safe", visible: 1, status: 1, sort: 34 },
  { id: 20, parentId: 0, name: "System", menuType: "directory", path: "", component: "", routeName: "system", perms: "", icon: "settings", visible: 1, status: 1, sort: 40 },
  { id: 21, parentId: 20, name: "Storage", menuType: "menu", path: "/storage", component: "StoragePage", routeName: "storage", perms: "view:storage", icon: "storage", visible: 1, status: 1, sort: 41 },
  { id: 22, parentId: 20, name: "Users", menuType: "menu", path: "/system/users", component: "UsersPage", routeName: "system-users", perms: "view:system-users", icon: "user", visible: 1, status: 1, sort: 42 },
  { id: 23, parentId: 20, name: "Departments", menuType: "menu", path: "/system/departments", component: "DepartmentsPage", routeName: "system-departments", perms: "view:system-departments", icon: "user-group", visible: 1, status: 1, sort: 43 },
  { id: 24, parentId: 20, name: "Roles", menuType: "menu", path: "/system/roles", component: "RolesPage", routeName: "system-roles", perms: "view:system-roles", icon: "safe", visible: 1, status: 1, sort: 44 },
  { id: 25, parentId: 20, name: "Login Logs", menuType: "menu", path: "/system/login-logs", component: "LoginLogsPage", routeName: "system-login-logs", perms: "view:system-login-logs", icon: "history", visible: 1, status: 1, sort: 45 },
  { id: 26, parentId: 20, name: "Settings", menuType: "menu", path: "/system/parameters", component: "SystemModulePage", routeName: "system-parameters", perms: "view:system-parameters", icon: "common", visible: 1, status: 1, sort: 46 },
  { id: 27, parentId: 20, name: "License", menuType: "menu", path: "/system/license", component: "SystemModulePage", routeName: "system-license", perms: "view:system-license", icon: "lock", visible: 1, status: 1, sort: 47 }
];

export const mockDepartments: DepartmentRecord[] = [
  {
    id: 1,
    parentId: 0,
    ancestors: "",
    name: "Platform Operations",
    leader: "Lena Wu",
    phone: "13800000001",
    email: "ops@visionvault.local",
    sort: 1,
    status: 1,
    createdAt: "2026-05-01 09:00:00",
    updatedAt: "2026-05-10 11:20:00"
  },
  {
    id: 2,
    parentId: 1,
    ancestors: "1",
    name: "Security Audit",
    leader: "Mason Li",
    phone: "13800000002",
    email: "audit@visionvault.local",
    sort: 10,
    status: 1,
    createdAt: "2026-05-02 10:00:00",
    updatedAt: "2026-05-11 14:15:00"
  },
  {
    id: 3,
    parentId: 1,
    ancestors: "1",
    name: "Field Support",
    leader: "Avery Chen",
    phone: "13800000003",
    email: "field@visionvault.local",
    sort: 20,
    status: 1,
    createdAt: "2026-05-03 08:30:00",
    updatedAt: "2026-05-09 16:40:00"
  }
];

export const mockUsers: UserRecord[] = [
  {
    id: 1,
    deptId: 1,
    deptName: "Platform Operations",
    username: "ops_admin",
    nickname: "Ops Admin",
    realName: "Olivia Sun",
    phone: "13900000001",
    email: "ops_admin@visionvault.local",
    status: 1,
    lastLoginIp: "10.16.1.18",
    lastLoginAt: "2026-05-12 09:21:16",
    roleIds: [2],
    roleNames: ["运维管理员"],
    passwordResetRequired: 0,
    createdAt: "2026-05-01 09:20:00",
    updatedAt: "2026-05-12 09:21:16"
  },
  {
    id: 2,
    deptId: 2,
    deptName: "Security Audit",
    username: "auditor_lee",
    nickname: "Audit Lee",
    realName: "Lee Fang",
    phone: "13900000002",
    email: "auditor@visionvault.local",
    status: 1,
    lastLoginIp: "10.18.12.66",
    lastLoginAt: "2026-05-12 08:11:08",
    roleIds: [3],
    roleNames: ["审计员"],
    passwordResetRequired: 1,
    createdAt: "2026-05-03 10:15:00",
    updatedAt: "2026-05-11 18:00:00"
  },
  {
    id: 3,
    deptId: 3,
    deptName: "Field Support",
    username: "field_cn",
    nickname: "Field CN",
    realName: "Chen Ning",
    phone: "13900000003",
    email: "field@visionvault.local",
    status: 0,
    lastLoginIp: "",
    lastLoginAt: "",
    roleIds: [],
    roleNames: [],
    passwordResetRequired: 1,
    createdAt: "2026-05-05 13:40:00",
    updatedAt: "2026-05-10 12:20:00"
  }
];

export const mockLoginLogs: LoginLogRecord[] = [
  {
    id: 1,
    userId: 1,
    username: "admin",
    loginIp: "10.16.1.18",
    userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/136.0.0.0",
    loginStatus: 1,
    message: "login succeeded",
    createdAt: "2026-05-12 09:21:16"
  },
  {
    id: 2,
    userId: 0,
    username: "admin",
    loginIp: "10.16.1.18",
    userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/136.0.0.0",
    loginStatus: 0,
    message: "invalid verification code",
    createdAt: "2026-05-12 09:20:44"
  },
  {
    id: 3,
    userId: 0,
    username: "ops_admin",
    loginIp: "10.18.12.66",
    userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/605.1.15",
    loginStatus: 0,
    message: "invalid username or password",
    createdAt: "2026-05-12 08:11:08"
  },
  {
    id: 4,
    userId: 1,
    username: "admin",
    loginIp: "127.0.0.1",
    userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/136.0.0.0",
    loginStatus: 1,
    message: "login succeeded",
    createdAt: "2026-05-11 18:03:52"
  }
];

export const mockLoginLogPage: PagedResult<LoginLogRecord> = {
  items: mockLoginLogs,
  total: mockLoginLogs.length,
  page: 1,
  pageSize: 20
};

export const mockAgents: AgentRecord[] = [
  {
    id: 1,
    hostSn: "SN-AF-001",
    hostName: "line-a-host-01",
    ip: "10.16.1.18",
    groupId: 1,
    storageId: 1,
    sourcePaths: ["D:/CameraArchive", "D:/CameraArchive/StationA1"],
    pathPrefix: "/assembly-a",
    version: "1.9.3.3",
    status: 1,
    tags: ["assembly", "line-a", "camera"],
    lastAccessTime: "2026-05-11 17:41",
    lastCommitTime: "2026-05-11 17:36",
    remark: "Primary collector for station A1",
    cpu: 42,
    mem: 58,
    storage: [
      { path: "C:", total: 512, used: 281, free: 231 },
      { path: "D:", total: 2048, used: 1610, free: 438 }
    ]
  },
  {
    id: 2,
    hostSn: "SN-AF-014",
    hostName: "line-a-host-14",
    ip: "10.16.1.43",
    groupId: 1,
    storageId: 1,
    sourcePaths: ["D:/CameraArchive"],
    pathPrefix: "/assembly-a",
    version: "1.9.2.8",
    status: 0,
    tags: ["assembly", "line-a", "offline"],
    lastAccessTime: "2026-05-11 15:10",
    lastCommitTime: "2026-05-11 14:54",
    remark: "Pending network check",
    cpu: 0,
    mem: 0,
    storage: [
      { path: "C:", total: 512, used: 325, free: 187 }
    ]
  },
  {
    id: 3,
    hostSn: "SN-QL-008",
    hostName: "lab-node-08",
    ip: "10.18.12.66",
    groupId: 2,
    storageId: 2,
    sourcePaths: ["/data/inspection", "/data/inspection/incoming"],
    pathPrefix: "/quality-labs",
    version: "1.9.3.3",
    status: 1,
    tags: ["quality", "lab", "night-shift"],
    lastAccessTime: "2026-05-11 17:42",
    lastCommitTime: "2026-05-11 17:35",
    remark: "Night shift ingest node",
    cpu: 67,
    mem: 73,
    storage: [
      { path: "/", total: 1024, used: 611, free: 413 }
    ]
  }
];

export const mockFiles: FileRecord[] = [
  {
    id: "file-1",
    name: "station-a1-20260511-090000.jpg",
    path: "/assembly-a/2026/05/11/station-a1-20260511-090000.jpg",
    type: "jpg",
    size: "18.4 MB",
    tags: ["assembly", "station-a1"],
    modifiedAt: "2026-05-11 09:00",
    storageId: 1,
    storage: "Factory Archive"
  },
  {
    id: "file-2",
    name: "lab-camera-08-20260511-080500.png",
    path: "/quality-labs/2026/05/11/lab-camera-08-20260511-080500.png",
    type: "png",
    size: "5.8 MB",
    tags: ["lab", "night-shift"],
    modifiedAt: "2026-05-11 08:05",
    storageId: 2,
    storage: "Cold Object Vault"
  },
  {
    id: "file-3",
    name: "inspection-report-20260510.zip",
    path: "/quality-labs/reports/inspection-report-20260510.zip",
    type: "zip",
    size: "420 MB",
    tags: ["report", "archive"],
    modifiedAt: "2026-05-10 23:10",
    storageId: 2,
    storage: "Cold Object Vault"
  }
];
