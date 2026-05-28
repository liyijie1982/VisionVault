import { request } from "./http";
import type {
  AgentRecord,
  AgentPayload,
  AgentDirectoryEntry,
  AlertGroupPayload,
  AlertGroupRecord,
  AlertLogRecord,
  AlertPolicyPayload,
  AlertPolicyRecord,
  DepartmentPayload,
  DepartmentRecord,
  FileFilterPayload,
  FileFilterRecord,
  FileLogRecord,
  FilePermissionPayload,
  FilePermissionRecord,
  FilePayload,
  FileRecord,
  GroupRecord,
  GroupPayload,
  HealthStatus,
  ImageProcessorPayload,
  ImageProcessorRecord,
  LicensePayload,
  LicenseRecord,
  LoginLogQuery,
  LoginLogRecord,
  MenuRecord,
  MessageChannelPayload,
  MessageChannelRecord,
  ModuleSummary,
  PagedResult,
  RegexRulePayload,
  RegexRuleRecord,
  RolePayload,
  RoleRecord,
  StorageRecord,
  StoragePayload,
  SyncLogRecord,
  SystemConfigPayload,
  SystemConfigRecord,
  TaskProgressPayload,
  TaskProgressRecord,
  UserPayload,
  UserRecord,
  VersionAgentSummary,
  VersionListResult,
  VersionVerifyResult,
  VersionRecord
} from "../types";
import { buildApiUrl } from "./http";

function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

function normalizePagedResult<T>(
  page: PagedResult<T> | null | undefined,
  normalizeItem?: (item: T) => T
): PagedResult<T> {
  const items = asArray(page?.items).map((item) => (normalizeItem ? normalizeItem(item) : item));
  return {
    items,
    total: typeof page?.total === "number" ? page.total : items.length,
    page: typeof page?.page === "number" ? page.page : 1,
    pageSize: typeof page?.pageSize === "number" ? page.pageSize : items.length
  };
}

function normalizeAgentRecord(record: AgentRecord): AgentRecord {
  return {
    ...record,
    tags: asArray(record.tags),
    sourcePaths: asArray(record.sourcePaths),
    storage: asArray(record.storage)
  };
}

function normalizeGroupRecord(record: GroupRecord): GroupRecord {
  return {
    ...record,
    intervalTime: typeof record.intervalTime === "number" ? record.intervalTime : 0,
    transferSpeedLimit: typeof record.transferSpeedLimit === "number" ? record.transferSpeedLimit : 0,
    workWindows: asArray(record.workWindows).map((item) => ({
      startTime: item?.startTime ?? "",
      endTime: item?.endTime ?? ""
    }))
  };
}

function normalizeRoleRecord(record: RoleRecord): RoleRecord {
  return {
    ...record,
    menuKeys: asArray(record.menuKeys)
  };
}

function normalizeMenuRecord(record: MenuRecord): MenuRecord {
  return {
    ...record,
    name: record.name ?? "",
    routeName: record.routeName ?? ""
  };
}

function normalizeDepartmentRecord(record: DepartmentRecord): DepartmentRecord {
  return {
    ...record,
    name: record.name ?? "",
    leader: record.leader ?? "",
    phone: record.phone ?? "",
    email: record.email ?? ""
  };
}

function normalizeUserRecord(record: UserRecord): UserRecord {
  return {
    ...record,
    deptName: record.deptName ?? "",
    realName: record.realName ?? "",
    phone: record.phone ?? "",
    email: record.email ?? "",
    lastLoginIp: record.lastLoginIp ?? "",
    lastLoginAt: record.lastLoginAt ?? "",
    roleIds: Array.isArray(record.roleIds) ? record.roleIds : [],
    roleNames: Array.isArray(record.roleNames) ? record.roleNames : []
  };
}

function normalizeStorageRecord(record: StorageRecord): StorageRecord {
  return {
    ...record,
    name: record.name ?? "",
    endpoint: record.endpoint ?? "",
    accessKey: record.accessKey ?? "",
    secretKey: record.secretKey ?? "",
    bucket: record.bucket ?? "",
    region: record.region ?? "",
    localPath: record.localPath ?? "",
    remark: record.remark ?? ""
  };
}

function normalizeFileRecord(record: FileRecord): FileRecord {
  return {
    ...record,
    name: record.name ?? "",
    path: record.path ?? "",
    type: record.type ?? "",
    size: record.size ?? "",
    tags: asArray(record.tags),
    modifiedAt: record.modifiedAt ?? "",
    storage: record.storage ?? ""
  };
}

function normalizeSyncLogRecord(record: SyncLogRecord): SyncLogRecord {
  return {
    ...record,
    agentIp: record.agentIp ?? "",
    hostName: record.hostName ?? "",
    path: record.path ?? "",
    startTime: record.startTime ?? "",
    fileSize: record.fileSize ?? "",
    logPath: record.logPath ?? "",
    commitTime: record.commitTime ?? ""
  };
}

function normalizeVersionRecord(record: VersionRecord): VersionRecord {
  return {
    ...record,
    version: record.version ?? "",
    filename: record.filename ?? "",
    md5: record.md5 ?? "",
    status: record.status ?? "",
    updatedAt: record.updatedAt ?? "",
    releaseNotes: record.releaseNotes ?? "",
    agentCount: typeof record.agentCount === "number" ? record.agentCount : 0
  };
}

function normalizeVersionListResult(result: VersionListResult | null | undefined): VersionListResult {
  return {
    items: asArray(result?.items).map(normalizeVersionRecord),
    agentVersions: asArray(result?.agentVersions).map((item) => ({
      version: item?.version ?? "",
      agentCount: typeof item?.agentCount === "number" ? item.agentCount : 0,
      isLatest: Boolean(item?.isLatest)
    })) as VersionAgentSummary[],
    totalAgents: typeof result?.totalAgents === "number" ? result.totalAgents : 0,
    publishedPackages: typeof result?.publishedPackages === "number" ? result.publishedPackages : 0,
    onlineAgents: typeof result?.onlineAgents === "number" ? result.onlineAgents : 0,
    currentPackageAgents: typeof result?.currentPackageAgents === "number" ? result.currentPackageAgents : 0
  };
}

export function fetchHealth() {
  return request<HealthStatus>("/healthz");
}

export function fetchModules() {
  return request<ModuleSummary[]>("/api/v1/meta/modules");
}

export async function fetchCurrentVersion() {
  const data = await request<{ id: string; version: string; filename: string; md5: string } | null>(
    "/sky/agent/version?version=0.0.0"
  );
  if (!data) {
    return null;
  }
  return {
    id: data.id,
    version: data.version,
    filename: data.filename,
    md5: data.md5,
    status: "Active",
    updatedAt: "Live backend package",
    releaseNotes: "Provided by current SkyBase backend.",
    agentCount: 0
  };
}

export async function fetchVersions() {
  const result = await request<VersionListResult | null>("/api/v1/versions");
  return normalizeVersionListResult(result);
}

export async function fetchAgents(): Promise<AgentRecord[]> {
  const agents = await request<AgentRecord[] | null>("/api/v1/agents");
  return asArray(agents).map(normalizeAgentRecord);
}

export async function createAgent(payload: AgentPayload): Promise<AgentRecord> {
  return request("/api/v1/agents", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateAgent(id: number, payload: AgentPayload): Promise<AgentRecord> {
  return request(`/api/v1/agents/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function fetchAgentDirectories(agentId: number, path = ""): Promise<AgentDirectoryEntry[]> {
  const query = path ? `?path=${encodeURIComponent(path)}` : "";
  return request<AgentDirectoryEntry[]>(`/api/v1/agent-directories/${agentId}${query}`);
}

export async function deleteAgent(id: number): Promise<{ deleted: boolean }> {
  return request(`/api/v1/agents/${id}`, {
    method: "DELETE"
  });
}

export async function fetchGroups(): Promise<GroupRecord[]> {
  const groups = await request<GroupRecord[] | null>("/api/v1/groups");
  return asArray(groups).map(normalizeGroupRecord);
}

export async function createGroup(payload: GroupPayload): Promise<GroupRecord> {
  return request("/api/v1/groups", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateGroup(id: number, payload: GroupPayload): Promise<GroupRecord> {
  return request(`/api/v1/groups/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteGroup(id: number): Promise<{ deleted: boolean }> {
  return request(`/api/v1/groups/${id}`, {
    method: "DELETE"
  });
}

export async function fetchRoles(): Promise<RoleRecord[]> {
  const roles = await request<RoleRecord[] | null>("/api/v1/roles");
  return asArray(roles).map(normalizeRoleRecord);
}

export async function fetchMenus(): Promise<MenuRecord[]> {
  const menus = await request<MenuRecord[] | null>("/api/v1/menus");
  return asArray(menus).map(normalizeMenuRecord);
}

export async function fetchDepartments(): Promise<DepartmentRecord[]> {
  const departments = await request<DepartmentRecord[] | null>("/api/v1/departments");
  return asArray(departments).map(normalizeDepartmentRecord);
}

export async function createDepartment(payload: DepartmentPayload): Promise<DepartmentRecord> {
  return request("/api/v1/departments", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateDepartment(id: number, payload: DepartmentPayload): Promise<DepartmentRecord> {
  return request(`/api/v1/departments/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteDepartment(id: number): Promise<{ deleted: boolean }> {
  return request(`/api/v1/departments/${id}`, {
    method: "DELETE"
  });
}

export async function fetchUsers(): Promise<UserRecord[]> {
  const users = await request<UserRecord[] | null>("/api/v1/users");
  return asArray(users).map(normalizeUserRecord);
}

export async function createUser(payload: UserPayload): Promise<UserRecord> {
  return request("/api/v1/users", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateUser(id: number, payload: UserPayload): Promise<UserRecord> {
  return request(`/api/v1/users/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteUser(id: number): Promise<{ deleted: boolean }> {
  return request(`/api/v1/users/${id}`, {
    method: "DELETE"
  });
}

export async function resetUserPassword(id: number, password: string): Promise<{ reset: boolean }> {
  return request(`/api/v1/users/${id}/reset-password`, {
    method: "POST",
    body: JSON.stringify({ password })
  });
}

export async function fetchLoginLogs(query: LoginLogQuery): Promise<PagedResult<LoginLogRecord>> {
  const params = new URLSearchParams();
  if (query.username?.trim()) {
    params.set("username", query.username.trim());
  }
  if (query.loginIp?.trim()) {
    params.set("loginIp", query.loginIp.trim());
  }
  if (typeof query.loginStatus === "number") {
    params.set("loginStatus", String(query.loginStatus));
  }
  if (query.startAt?.trim()) {
    params.set("startAt", query.startAt.trim());
  }
  if (query.endAt?.trim()) {
    params.set("endAt", query.endAt.trim());
  }
  params.set("page", String(query.page ?? 1));
  params.set("pageSize", String(query.pageSize ?? 20));

  const page = await request<PagedResult<LoginLogRecord> | null>(`/api/v1/login-logs?${params.toString()}`);
  return normalizePagedResult(page);
}

export async function createRole(payload: RolePayload): Promise<RoleRecord> {
  return request("/api/v1/roles", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateRole(id: number, payload: RolePayload): Promise<RoleRecord> {
  return request(`/api/v1/roles/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteRole(id: number): Promise<{ deleted: boolean }> {
  return request(`/api/v1/roles/${id}`, {
    method: "DELETE"
  });
}

export async function fetchStorageTargets(): Promise<StorageRecord[]> {
  const targets = await request<StorageRecord[] | null>("/api/v1/storage");
  return asArray(targets).map(normalizeStorageRecord);
}

export async function createStorageTarget(payload: StoragePayload): Promise<StorageRecord> {
  return request("/api/v1/storage", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateStorageTarget(id: number, payload: StoragePayload): Promise<StorageRecord> {
  return request(`/api/v1/storage/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteStorageTarget(id: number): Promise<{ deleted: boolean }> {
  return request(`/api/v1/storage/${id}`, {
    method: "DELETE"
  });
}

export async function fetchFiles(): Promise<FileRecord[]> {
  const files = await request<FileRecord[] | null>("/api/v1/files");
  return asArray(files).map(normalizeFileRecord);
}

export async function createFile(payload: FilePayload): Promise<FileRecord> {
  return request("/api/v1/files", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateFile(id: string, payload: FilePayload): Promise<FileRecord> {
  return request(`/api/v1/files/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteFile(id: string): Promise<{ deleted: boolean }> {
  return request(`/api/v1/files/${id}`, {
    method: "DELETE"
  });
}

export async function fetchSyncLogs(result = "all"): Promise<SyncLogRecord[]> {
  const params = new URLSearchParams();
  if (result && result !== "all") {
    params.set("result", result);
  }
  const query = params.toString();
  const logs = await request<SyncLogRecord[] | null>(`/api/v1/sync-logs${query ? `?${query}` : ""}`);
  return asArray(logs).map(normalizeSyncLogRecord);
}

export function buildFileDownloadUrl(id: string) {
  return buildApiUrl(`/api/v1/files/${id}/download`);
}

export async function uploadFile(storageId: number, file: File, tags: string[]) {
  const body = new FormData();
  body.append("storageId", String(storageId));
  body.append("file", file);
  body.append("tags", tags.join(","));

  const response = await fetch(buildApiUrl("/api/v1/files/upload"), {
    method: "POST",
    credentials: "include",
    body
  });
  const payload = await response.json();
  if (!response.ok || payload.code !== 0) {
    throw new Error(payload.msg || "Failed to upload file");
  }
  return normalizeFileRecord(payload.data as FileRecord);
}

export async function uploadVersion(version: string, file: File, releaseNotes: string, activate: boolean) {
  const body = new FormData();
  body.append("version", version);
  body.append("file", file);
  body.append("releaseNotes", releaseNotes);
  body.append("activate", activate ? "true" : "false");

  const response = await fetch(buildApiUrl("/api/v1/versions/upload"), {
    method: "POST",
    credentials: "include",
    body
  });
  const payload = await response.json();
  if (!response.ok || payload.code !== 0) {
    throw new Error(payload.msg || "Failed to upload package");
  }
  return normalizeVersionRecord(payload.data as VersionRecord);
}

export function downloadAgentPackageById(id: string) {
  return buildApiUrl(`/sky/agent/download?id=${encodeURIComponent(id)}`);
}

export async function verifyVersionMD5(id: string) {
  return request<VersionVerifyResult>(`/api/v1/versions/${id}/verify-md5`, {
    method: "POST"
  });
}

export async function fetchFileFilters() {
  return request<FileFilterRecord[]>("/api/v1/file-filters");
}

export async function createFileFilter(payload: FileFilterPayload) {
  return request<FileFilterRecord>("/api/v1/file-filters", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateFileFilter(id: number, payload: FileFilterPayload) {
  return request<FileFilterRecord>(`/api/v1/file-filters/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteFileFilter(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/file-filters/${id}`, { method: "DELETE" });
}

export async function fetchRegexRules() {
  return request<RegexRuleRecord[]>("/api/v1/regex-rules");
}

export async function createRegexRule(payload: RegexRulePayload) {
  return request<RegexRuleRecord>("/api/v1/regex-rules", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateRegexRule(id: number, payload: RegexRulePayload) {
  return request<RegexRuleRecord>(`/api/v1/regex-rules/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteRegexRule(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/regex-rules/${id}`, { method: "DELETE" });
}

export async function fetchImageProcessors() {
  return request<ImageProcessorRecord[]>("/api/v1/image-processors");
}

export async function createImageProcessor(payload: ImageProcessorPayload) {
  return request<ImageProcessorRecord>("/api/v1/image-processors", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateImageProcessor(id: number, payload: ImageProcessorPayload) {
  return request<ImageProcessorRecord>(`/api/v1/image-processors/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteImageProcessor(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/image-processors/${id}`, { method: "DELETE" });
}

export async function fetchAlertGroups() {
  return request<AlertGroupRecord[]>("/api/v1/alert-groups");
}

export async function createAlertGroup(payload: AlertGroupPayload) {
  return request<AlertGroupRecord>("/api/v1/alert-groups", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateAlertGroup(id: number, payload: AlertGroupPayload) {
  return request<AlertGroupRecord>(`/api/v1/alert-groups/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteAlertGroup(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/alert-groups/${id}`, { method: "DELETE" });
}

export async function fetchMessageChannels() {
  return request<MessageChannelRecord[]>("/api/v1/message-channels");
}

export async function createMessageChannel(payload: MessageChannelPayload) {
  return request<MessageChannelRecord>("/api/v1/message-channels", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateMessageChannel(id: number, payload: MessageChannelPayload) {
  return request<MessageChannelRecord>(`/api/v1/message-channels/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteMessageChannel(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/message-channels/${id}`, { method: "DELETE" });
}

export async function fetchAlertPolicies() {
  return request<AlertPolicyRecord[]>("/api/v1/alert-policies");
}

export async function createAlertPolicy(payload: AlertPolicyPayload) {
  return request<AlertPolicyRecord>("/api/v1/alert-policies", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateAlertPolicy(id: number, payload: AlertPolicyPayload) {
  return request<AlertPolicyRecord>(`/api/v1/alert-policies/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteAlertPolicy(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/alert-policies/${id}`, { method: "DELETE" });
}

export async function fetchFileLogs() {
  return request<FileLogRecord[]>("/api/v1/file-logs");
}

export async function fetchFilePermissions() {
  return request<FilePermissionRecord[]>("/api/v1/file-permissions");
}

export async function createFilePermission(payload: FilePermissionPayload) {
  return request<FilePermissionRecord>("/api/v1/file-permissions", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateFilePermission(id: number, payload: FilePermissionPayload) {
  return request<FilePermissionRecord>(`/api/v1/file-permissions/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteFilePermission(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/file-permissions/${id}`, { method: "DELETE" });
}

export async function fetchTaskProgress() {
  return request<TaskProgressRecord[]>("/api/v1/task-progress");
}

export async function createTaskProgress(payload: TaskProgressPayload) {
  return request<TaskProgressRecord>("/api/v1/task-progress", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateTaskProgress(id: number, payload: TaskProgressPayload) {
  return request<TaskProgressRecord>(`/api/v1/task-progress/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteTaskProgress(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/task-progress/${id}`, { method: "DELETE" });
}

export async function fetchAlertLogs() {
  return request<AlertLogRecord[]>("/api/v1/alert-logs");
}

export async function fetchSystemConfigs() {
  return request<SystemConfigRecord[]>("/api/v1/system-configs");
}

export async function createSystemConfig(payload: SystemConfigPayload) {
  return request<SystemConfigRecord>("/api/v1/system-configs", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateSystemConfig(id: number, payload: SystemConfigPayload) {
  return request<SystemConfigRecord>(`/api/v1/system-configs/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteSystemConfig(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/system-configs/${id}`, { method: "DELETE" });
}

export async function fetchLicenses() {
  return request<LicenseRecord[]>("/api/v1/licenses");
}

export async function createLicense(payload: LicensePayload) {
  return request<LicenseRecord>("/api/v1/licenses", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function updateLicense(id: number, payload: LicensePayload) {
  return request<LicenseRecord>(`/api/v1/licenses/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function deleteLicense(id: number) {
  return request<{ deleted: boolean }>(`/api/v1/licenses/${id}`, { method: "DELETE" });
}
