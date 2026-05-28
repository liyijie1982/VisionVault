export interface ApiEnvelope<T> {
  code: number;
  msg: string;
  data: T;
}

export interface AuthUser {
  id: number;
  username: string;
  nickname: string;
  realName: string;
  phone: string;
  email: string;
  status: number;
  deptId: number;
  roleKeys: string[];
  roleNames: string[];
  menuKeys: string[];
  passwordResetRequired: number;
  createdAt: string;
  updatedAt: string;
}

export interface LoginResponse {
  user: AuthUser;
}

export interface ChangePasswordPayload {
  currentPassword: string;
  newPassword: string;
}

export interface CaptchaResponse {
  captchaEnabled: boolean;
  uuid: string;
  img: string;
}

export interface ModuleSummary {
  key: string;
  name: string;
  description: string;
  status: string;
}

export interface HealthStatus {
  name: string;
  env: string;
  status: string;
  startedAt: string;
  now: string;
  modules: ModuleSummary[];
}

export interface StorageMetric {
  path: string;
  total: number;
  used: number;
  free: number;
}

export interface AgentRecord {
  id: number;
  hostSn: string;
  hostName: string;
  ip: string;
  groupId: number;
  storageId: number;
  sourcePaths: string[];
  pathPrefix: string;
  version: string;
  status: number;
  tags: string[];
  lastAccessTime: string;
  lastCommitTime: string;
  remark: string;
  cpu: number;
  mem: number;
  storage: StorageMetric[];
}

export interface GroupRecord {
  id: number;
  name: string;
  storageId: number;
  ipRange: string;
  pathPrefix: string;
  intervalTime: number;
  delTimeDays: number;
  transferSpeedLimit: number;
  workWindows: Array<{ startTime: string; endTime: string }>;
  fileFilterId: number;
  regexId: number;
  imageProcessId: number;
  alarmGroupId: number;
  logEnabled: boolean;
  status: number;
  createdAt: string;
  updatedAt: string;
}

export interface StorageRecord {
  id: number;
  name: string;
  type: "local" | "s3";
  endpoint: string;
  accessKey: string;
  secretKey: string;
  bucket: string;
  region: string;
  localPath: string;
  quota: number;
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface SyncLogRecord {
  id: string;
  agentIp: string;
  hostName: string;
  path: string;
  startTime: string;
  fileCount: number;
  fileSize: string;
  errorCount: number;
  logPath: string;
  commitTime: string;
}

export interface VersionRecord {
  id: string;
  version: string;
  filename: string;
  md5: string;
  status: string;
  updatedAt: string;
  releaseNotes: string;
  agentCount: number;
}

export interface VersionAgentSummary {
  version: string;
  agentCount: number;
  isLatest: boolean;
}

export interface VersionListResult {
  items: VersionRecord[];
  agentVersions: VersionAgentSummary[];
  totalAgents: number;
  publishedPackages: number;
  onlineAgents: number;
  currentPackageAgents: number;
}

export interface FileRecord {
  id: string;
  name: string;
  path: string;
  type: string;
  size: string;
  tags: string[];
  modifiedAt: string;
  storageId: number;
  storage: string;
}

export interface AgentPayload {
  id?: number;
  hostSn: string;
  hostName: string;
  ip: string;
  groupId: number;
  storageId: number;
  sourcePaths: string[];
  pathPrefix: string;
  version: string;
  status: number;
  tags: string[];
  remark: string;
}

export interface AgentDirectoryEntry {
  name: string;
  path: string;
}

export interface GroupPayload {
  name: string;
  storageId: number;
  ipRange: string;
  pathPrefix: string;
  intervalTime: number;
  delTimeDays: number;
  transferSpeedLimit: number;
  workWindows: Array<{ startTime: string; endTime: string }>;
  fileFilterId: number;
  regexId: number;
  imageProcessId: number;
  alarmGroupId: number;
  logEnabled: boolean;
  status: number;
}

export interface StoragePayload {
  name: string;
  type: "local" | "s3";
  endpoint: string;
  accessKey: string;
  secretKey: string;
  bucket: string;
  region: string;
  localPath: string;
  quota: number;
  status: number;
  remark: string;
}

export interface FilePayload {
  name: string;
  path: string;
  type: string;
  size: string;
  tags: string[];
  modifiedAt: string;
  storageId: number;
}

export interface RoleRecord {
  id: number;
  name: string;
  key: string;
  dataScope: string;
  sort: number;
  status: number;
  description: string;
  menuKeys: string[];
  createdAt: string;
  updatedAt: string;
}

export interface RolePayload {
  name: string;
  key: string;
  dataScope: string;
  sort: number;
  status: number;
  description: string;
  menuKeys: string[];
}

export interface MenuRecord {
  id: number;
  parentId: number;
  name: string;
  menuType: string;
  path: string;
  component: string;
  routeName: string;
  perms: string;
  icon: string;
  visible: number;
  status: number;
  sort: number;
}

export interface DepartmentRecord {
  id: number;
  parentId: number;
  ancestors: string;
  name: string;
  leader: string;
  phone: string;
  email: string;
  sort: number;
  status: number;
  createdAt: string;
  updatedAt: string;
}

export interface DepartmentPayload {
  parentId: number;
  name: string;
  leader: string;
  phone: string;
  email: string;
  sort: number;
  status: number;
}

export interface UserRecord {
  id: number;
  deptId: number;
  deptName: string;
  username: string;
  nickname: string;
  realName: string;
  phone: string;
  email: string;
  status: number;
  lastLoginIp: string;
  lastLoginAt: string;
  roleIds: number[];
  roleNames: string[];
  passwordResetRequired: number;
  createdAt: string;
  updatedAt: string;
}

export interface UserPayload {
  deptId: number;
  username: string;
  nickname: string;
  realName: string;
  phone: string;
  email: string;
  password?: string;
  status: number;
  roleIds: number[];
  passwordResetRequired: number;
}

export interface LoginLogRecord {
  id: number;
  userId: number;
  username: string;
  loginIp: string;
  userAgent: string;
  loginStatus: number;
  message: string;
  createdAt: string;
}

export interface LoginLogQuery {
  username?: string;
  loginIp?: string;
  loginStatus?: number;
  startAt?: string;
  endAt?: string;
  page?: number;
  pageSize?: number;
}

export interface PagedResult<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface FileFilterRecord {
  id: number;
  name: string;
  filterScope: string;
  listType: string;
  patterns: string[];
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface FileFilterPayload {
  name: string;
  filterScope: string;
  listType: string;
  patterns: string[];
  status: number;
  remark: string;
}

export interface RegexTagMapping {
  captureIndex: number;
  tagKey: string;
}

export interface RegexRuleRecord {
  id: number;
  name: string;
  sourceField: string;
  regexp: string;
  asPath: number;
  status: number;
  mappings: RegexTagMapping[];
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface RegexRulePayload {
  name: string;
  sourceField: string;
  regexp: string;
  asPath: number;
  status: number;
  mappings: RegexTagMapping[];
  remark: string;
}

export interface ImageProcessorRecord {
  id: number;
  name: string;
  processorType: string;
  configJson: string;
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface ImageProcessorPayload {
  name: string;
  processorType: string;
  configJson: string;
  status: number;
  remark: string;
}

export interface AlertGroupRecord {
  id: number;
  name: string;
  receivers: string[];
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface AlertGroupPayload {
  name: string;
  receivers: string[];
  status: number;
  remark: string;
}

export interface MessageChannelRecord {
  id: number;
  name: string;
  channelType: string;
  configJson: string;
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface MessageChannelPayload {
  name: string;
  channelType: string;
  configJson: string;
  status: number;
  remark: string;
}

export interface AlertPolicyRecord {
  id: number;
  name: string;
  cpuThreshold: number;
  memThreshold: number;
  diskThreshold: number;
  cpuConsecutiveTimes: number;
  memConsecutiveTimes: number;
  heartbeatTimeoutSeconds: number;
  sendFrequencySeconds: number;
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface AlertPolicyPayload {
  name: string;
  cpuThreshold: number;
  memThreshold: number;
  diskThreshold: number;
  cpuConsecutiveTimes: number;
  memConsecutiveTimes: number;
  heartbeatTimeoutSeconds: number;
  sendFrequencySeconds: number;
  status: number;
  remark: string;
}

export interface FileLogRecord {
  id: number;
  userId: number;
  username: string;
  storageId: number;
  storageName: string;
  filePath: string;
  operationType: string;
  resultStatus: string;
  clientIp: string;
  message: string;
  createdAt: string;
}

export interface FilePermissionRecord {
  id: number;
  userId: number;
  username: string;
  storageId: number;
  storageName: string;
  canView: number;
  canUpload: number;
  canDownload: number;
  canDelete: number;
  canBatchDownload: number;
  canDownloadToServer: number;
  createdAt: string;
  updatedAt: string;
}

export interface FilePermissionPayload {
  userId: number;
  storageId: number;
  canView: number;
  canUpload: number;
  canDownload: number;
  canDelete: number;
  canBatchDownload: number;
  canDownloadToServer: number;
}

export interface TaskProgressRecord {
  id: number;
  taskType: string;
  status: string;
  totalCount: number;
  successCount: number;
  failedCount: number;
  payloadJson: string;
  resultJson: string;
  startedAt: string;
  finishedAt: string;
  createdAt: string;
  remark: string;
}

export interface TaskProgressPayload {
  taskType: string;
  status: string;
  totalCount: number;
  successCount: number;
  failedCount: number;
  payloadJson: string;
  resultJson: string;
  startedAt: string;
  finishedAt: string;
  remark: string;
}

export interface AlertLogRecord {
  id: number;
  agentId: number;
  groupId: number;
  ruleId: number;
  messageChannelId: number;
  alertLevel: string;
  alertTitle: string;
  alertBody: string;
  sendStatus: string;
  failureReason: string;
  createdAt: string;
}

export interface SystemConfigRecord {
  id: number;
  configGroup: string;
  configKey: string;
  configValue: string;
  valueType: string;
  isEncrypted: number;
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface SystemConfigPayload {
  configGroup: string;
  configKey: string;
  configValue: string;
  valueType: string;
  isEncrypted: number;
  status: number;
  remark: string;
}

export interface LicenseRecord {
  id: number;
  licenseCode: string;
  serialNumber: string;
  maxAgentCount: number;
  issuedAt: string;
  expiredAt: string;
  trialDays: number;
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface LicensePayload {
  licenseCode: string;
  serialNumber: string;
  maxAgentCount: number;
  issuedAt: string;
  expiredAt: string;
  trialDays: number;
  status: number;
  remark: string;
}

export interface VersionVerifyResult {
  id: string;
  expected: string;
  actual: string;
  matched: boolean;
}
