<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { Message } from "@arco-design/web-vue";
import { IconApps, IconCalendarClock, IconCheckCircle, IconCommon, IconLock, IconSafe, IconUser } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import {
  createAlertGroup,
  createAlertPolicy,
  createFileFilter,
  createFilePermission,
  createImageProcessor,
  createLicense,
  createMessageChannel,
  createRegexRule,
  createSystemConfig,
  createTaskProgress,
  deleteAlertGroup,
  deleteAlertPolicy,
  deleteFileFilter,
  deleteFilePermission,
  deleteImageProcessor,
  deleteLicense,
  deleteMessageChannel,
  deleteRegexRule,
  deleteSystemConfig,
  deleteTaskProgress,
  fetchAlertGroups,
  fetchAlertLogs,
  fetchAlertPolicies,
  fetchFileFilters,
  fetchFileLogs,
  fetchFilePermissions,
  fetchImageProcessors,
  fetchLicenses,
  fetchMessageChannels,
  fetchRegexRules,
  fetchStorageTargets,
  fetchSystemConfigs,
  fetchTaskProgress,
  fetchUsers,
  updateAlertGroup,
  updateAlertPolicy,
  updateFileFilter,
  updateFilePermission,
  updateImageProcessor,
  updateLicense,
  updateMessageChannel,
  updateRegexRule,
  updateSystemConfig,
  updateTaskProgress
} from "../api/skybase";
import type {
  AlertGroupPayload,
  AlertGroupRecord,
  AlertLogRecord,
  AlertPolicyPayload,
  AlertPolicyRecord,
  FileFilterPayload,
  FileFilterRecord,
  FileLogRecord,
  FilePermissionPayload,
  FilePermissionRecord,
  ImageProcessorPayload,
  ImageProcessorRecord,
  LicensePayload,
  LicenseRecord,
  MessageChannelPayload,
  MessageChannelRecord,
  RegexRulePayload,
  RegexRuleRecord,
  StorageRecord,
  SystemConfigPayload,
  SystemConfigRecord,
  TaskProgressPayload,
  TaskProgressRecord,
  UserRecord
} from "../types";

const route = useRoute();
const loading = ref(false);
const editorVisible = ref(false);
const editorType = ref("");
const editingId = ref<number | null>(null);

const fileFilters = ref<FileFilterRecord[]>([]);
const regexRules = ref<RegexRuleRecord[]>([]);
const imageProcessors = ref<ImageProcessorRecord[]>([]);
const fileLogs = ref<FileLogRecord[]>([]);
const filePermissions = ref<FilePermissionRecord[]>([]);
const taskProgress = ref<TaskProgressRecord[]>([]);
const alertGroups = ref<AlertGroupRecord[]>([]);
const alertLogs = ref<AlertLogRecord[]>([]);
const messageChannels = ref<MessageChannelRecord[]>([]);
const alertPolicies = ref<AlertPolicyRecord[]>([]);
const systemConfigs = ref<SystemConfigRecord[]>([]);
const licenses = ref<LicenseRecord[]>([]);
const users = ref<UserRecord[]>([]);
const storages = ref<StorageRecord[]>([]);

const fileFilterForm = reactive<FileFilterPayload>({
  name: "",
  filterScope: "extension",
  listType: "whitelist",
  patterns: [],
  status: 1,
  remark: ""
});
const regexRuleForm = reactive<RegexRulePayload>({
  name: "",
  sourceField: "path",
  regexp: "",
  asPath: 0,
  status: 1,
  mappings: [],
  remark: ""
});
const imageProcessorForm = reactive<ImageProcessorPayload>({
  name: "",
  processorType: "transform",
  configJson: "{}",
  status: 1,
  remark: ""
});
const filePermissionForm = reactive<FilePermissionPayload>({
  userId: 0,
  storageId: 0,
  canView: 1,
  canUpload: 0,
  canDownload: 0,
  canDelete: 0,
  canBatchDownload: 0,
  canDownloadToServer: 0
});
const taskProgressForm = reactive<TaskProgressPayload>({
  taskType: "",
  status: "pending",
  totalCount: 0,
  successCount: 0,
  failedCount: 0,
  payloadJson: "{}",
  resultJson: "{}",
  startedAt: "",
  finishedAt: "",
  remark: ""
});
const alertGroupForm = reactive<AlertGroupPayload>({
  name: "",
  receivers: [],
  status: 1,
  remark: ""
});
const messageChannelForm = reactive<MessageChannelPayload>({
  name: "",
  channelType: "email",
  configJson: "{}",
  status: 1,
  remark: ""
});
const alertPolicyForm = reactive<AlertPolicyPayload>({
  name: "",
  cpuThreshold: 0,
  memThreshold: 0,
  diskThreshold: 0,
  cpuConsecutiveTimes: 0,
  memConsecutiveTimes: 0,
  heartbeatTimeoutSeconds: 0,
  sendFrequencySeconds: 0,
  status: 1,
  remark: ""
});
const systemConfigForm = reactive<SystemConfigPayload>({
  configGroup: "system",
  configKey: "",
  configValue: "",
  valueType: "string",
  isEncrypted: 0,
  status: 1,
  remark: ""
});
const licenseForm = reactive<LicensePayload>({
  licenseCode: "",
  serialNumber: "",
  maxAgentCount: 0,
  issuedAt: "",
  expiredAt: "",
  trialDays: 30,
  status: 1,
  remark: ""
});

const fileFilterPatternsInput = ref("");
const regexMappingsInput = ref("");
const receiversInput = ref("");

const moduleKey = computed(() => String(route.name));

const pageMeta = computed(() => {
  const metaMap: Record<string, { title: string; description: string; primaryAction: string; secondaryAction: string }> = {
    "extraction-rules": { title: "Extraction Rules", description: "Manage file filters, regex extraction rules, and image processor configs.", primaryAction: "Add Rule", secondaryAction: "Refresh" },
    "file-logs": { title: "File Logs", description: "Review real file upload, download, and delete operations.", primaryAction: "Export Logs", secondaryAction: "Refresh" },
    "file-permissions": { title: "File Permissions", description: "Grant storage-specific permissions to platform users.", primaryAction: "Grant Permission", secondaryAction: "Refresh" },
    "task-progress": { title: "Task Progress", description: "Track and update async job execution state.", primaryAction: "Create Task", secondaryAction: "Refresh" },
    "alert-groups": { title: "Alert Groups", description: "Manage alert receiver groups and escalation members.", primaryAction: "Create Group", secondaryAction: "Refresh" },
    "alert-logs": { title: "Alert Logs", description: "Inspect alert delivery history and failures.", primaryAction: "Export Logs", secondaryAction: "Refresh" },
    "message-channels": { title: "Message Channels", description: "Configure outbound notification channels and connection config.", primaryAction: "Add Channel", secondaryAction: "Refresh" },
    "alert-policies": { title: "Alert Policies", description: "Set CPU, memory, disk, and heartbeat thresholds.", primaryAction: "Create Policy", secondaryAction: "Refresh" },
    "system-parameters": { title: "Settings", description: "Edit system-level configuration values.", primaryAction: "Add Config", secondaryAction: "Refresh" },
    "system-license": { title: "License", description: "Manage license records and validity windows.", primaryAction: "Add License", secondaryAction: "Refresh" }
  };
  return metaMap[moduleKey.value] ?? metaMap["system-parameters"];
});

const stats = computed(() => {
  switch (moduleKey.value) {
    case "extraction-rules":
      return [
        { label: "File Filters", value: fileFilters.value.length, hint: "Filter definitions", icon: IconCommon },
        { label: "Regex Rules", value: regexRules.value.length, hint: "Extraction rules", icon: IconCheckCircle },
        { label: "Image Processors", value: imageProcessors.value.length, hint: "Processing presets", icon: IconApps },
        { label: "Active Items", value: fileFilters.value.filter((item) => item.status === 1).length + regexRules.value.filter((item) => item.status === 1).length + imageProcessors.value.filter((item) => item.status === 1).length, hint: "Enabled definitions", icon: IconSafe }
      ];
    case "file-logs":
      return [
        { label: "Log Entries", value: fileLogs.value.length, hint: "Visible file operations", icon: IconCommon },
        { label: "Uploads", value: fileLogs.value.filter((item) => item.operationType === "upload").length, hint: "Upload actions", icon: IconApps },
        { label: "Downloads", value: fileLogs.value.filter((item) => item.operationType === "download").length, hint: "Download actions", icon: IconCheckCircle },
        { label: "Failures", value: fileLogs.value.filter((item) => item.resultStatus !== "success").length, hint: "Non-success operations", icon: IconLock }
      ];
    case "file-permissions":
      return [
        { label: "Permission Rows", value: filePermissions.value.length, hint: "Explicit grants", icon: IconSafe },
        { label: "Upload Grants", value: filePermissions.value.filter((item) => item.canUpload === 1).length, hint: "Can upload", icon: IconApps },
        { label: "Delete Grants", value: filePermissions.value.filter((item) => item.canDelete === 1).length, hint: "Can delete", icon: IconLock },
        { label: "Users Covered", value: new Set(filePermissions.value.map((item) => item.userId)).size, hint: "Distinct users", icon: IconUser }
      ];
    case "task-progress":
      return [
        { label: "Tasks", value: taskProgress.value.length, hint: "Tracked tasks", icon: IconApps },
        { label: "Running", value: taskProgress.value.filter((item) => item.status === "running").length, hint: "Active jobs", icon: IconCalendarClock },
        { label: "Succeeded", value: taskProgress.value.filter((item) => item.status === "success").length, hint: "Completed jobs", icon: IconCheckCircle },
        { label: "Failed", value: taskProgress.value.filter((item) => item.status === "failed").length, hint: "Needs retry", icon: IconLock }
      ];
    case "alert-groups":
      return [
        { label: "Alert Groups", value: alertGroups.value.length, hint: "Receiver groups", icon: IconApps },
        { label: "Active", value: alertGroups.value.filter((item) => item.status === 1).length, hint: "Enabled groups", icon: IconCheckCircle },
        { label: "Receivers", value: alertGroups.value.reduce((sum, item) => sum + item.receivers.length, 0), hint: "Configured recipients", icon: IconUser },
        { label: "Disabled", value: alertGroups.value.filter((item) => item.status === 0).length, hint: "Inactive groups", icon: IconLock }
      ];
    case "alert-logs":
      return [
        { label: "Alerts", value: alertLogs.value.length, hint: "Visible alert logs", icon: IconApps },
        { label: "Delivered", value: alertLogs.value.filter((item) => item.sendStatus === "success").length, hint: "Successful sends", icon: IconCheckCircle },
        { label: "Pending", value: alertLogs.value.filter((item) => item.sendStatus === "pending").length, hint: "Queued sends", icon: IconCalendarClock },
        { label: "Failures", value: alertLogs.value.filter((item) => item.sendStatus !== "success" && item.sendStatus !== "pending").length, hint: "Failed sends", icon: IconLock }
      ];
    case "message-channels":
      return [
        { label: "Channels", value: messageChannels.value.length, hint: "Configured channels", icon: IconApps },
        { label: "Email", value: messageChannels.value.filter((item) => item.channelType === "email").length, hint: "Email channels", icon: IconCommon },
        { label: "Enabled", value: messageChannels.value.filter((item) => item.status === 1).length, hint: "Enabled channels", icon: IconCheckCircle },
        { label: "Disabled", value: messageChannels.value.filter((item) => item.status === 0).length, hint: "Disabled channels", icon: IconLock }
      ];
    case "alert-policies":
      return [
        { label: "Policies", value: alertPolicies.value.length, hint: "Threshold rules", icon: IconSafe },
        { label: "Enabled", value: alertPolicies.value.filter((item) => item.status === 1).length, hint: "Active policies", icon: IconCheckCircle },
        { label: "Heartbeat Watch", value: alertPolicies.value.filter((item) => item.heartbeatTimeoutSeconds > 0).length, hint: "With heartbeat checks", icon: IconCalendarClock },
        { label: "Throttled", value: alertPolicies.value.filter((item) => item.sendFrequencySeconds > 0).length, hint: "Rate-limited policies", icon: IconLock }
      ];
    case "system-parameters":
      return [
        { label: "Config Keys", value: systemConfigs.value.length, hint: "Stored parameters", icon: IconCommon },
        { label: "System Group", value: systemConfigs.value.filter((item) => item.configGroup === "system").length, hint: "System configs", icon: IconApps },
        { label: "Encrypted", value: systemConfigs.value.filter((item) => item.isEncrypted === 1).length, hint: "Encrypted values", icon: IconLock },
        { label: "Enabled", value: systemConfigs.value.filter((item) => item.status === 1).length, hint: "Active values", icon: IconCheckCircle }
      ];
    case "system-license":
      return [
        { label: "License Rows", value: licenses.value.length, hint: "License records", icon: IconSafe },
        { label: "Active", value: licenses.value.filter((item) => item.status === 1).length, hint: "Enabled licenses", icon: IconCheckCircle },
        { label: "Max Agents", value: licenses.value.reduce((max, item) => Math.max(max, item.maxAgentCount), 0), hint: "Largest capacity", icon: IconApps },
        { label: "Trial Days", value: licenses.value[0]?.trialDays ?? 0, hint: "Current trial setting", icon: IconCalendarClock }
      ];
    default:
      return [];
  }
});

function parseListInput(value: string) {
  return value.split(/[\n,]/).map((item) => item.trim()).filter(Boolean);
}

function exportCurrentData() {
  let data: unknown = null;
  switch (moduleKey.value) {
    case "file-logs":
      data = fileLogs.value;
      break;
    case "alert-logs":
      data = alertLogs.value;
      break;
    default:
      data = {
        fileFilters: fileFilters.value,
        regexRules: regexRules.value,
        imageProcessors: imageProcessors.value,
        filePermissions: filePermissions.value,
        taskProgress: taskProgress.value,
        alertGroups: alertGroups.value,
        messageChannels: messageChannels.value,
        alertPolicies: alertPolicies.value,
        systemConfigs: systemConfigs.value,
        licenses: licenses.value
      };
      break;
  }
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  window.open(url, "_blank");
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

function resetEditorState() {
  editingId.value = null;
  fileFilterForm.name = "";
  fileFilterForm.filterScope = "extension";
  fileFilterForm.listType = "whitelist";
  fileFilterForm.patterns = [];
  fileFilterForm.status = 1;
  fileFilterForm.remark = "";
  fileFilterPatternsInput.value = "";

  regexRuleForm.name = "";
  regexRuleForm.sourceField = "path";
  regexRuleForm.regexp = "";
  regexRuleForm.asPath = 0;
  regexRuleForm.status = 1;
  regexRuleForm.mappings = [];
  regexRuleForm.remark = "";
  regexMappingsInput.value = "";

  imageProcessorForm.name = "";
  imageProcessorForm.processorType = "transform";
  imageProcessorForm.configJson = "{}";
  imageProcessorForm.status = 1;
  imageProcessorForm.remark = "";

  filePermissionForm.userId = users.value[0]?.id ?? 0;
  filePermissionForm.storageId = storages.value[0]?.id ?? 0;
  filePermissionForm.canView = 1;
  filePermissionForm.canUpload = 0;
  filePermissionForm.canDownload = 0;
  filePermissionForm.canDelete = 0;
  filePermissionForm.canBatchDownload = 0;
  filePermissionForm.canDownloadToServer = 0;

  taskProgressForm.taskType = "";
  taskProgressForm.status = "pending";
  taskProgressForm.totalCount = 0;
  taskProgressForm.successCount = 0;
  taskProgressForm.failedCount = 0;
  taskProgressForm.payloadJson = "{}";
  taskProgressForm.resultJson = "{}";
  taskProgressForm.startedAt = "";
  taskProgressForm.finishedAt = "";
  taskProgressForm.remark = "";

  alertGroupForm.name = "";
  alertGroupForm.receivers = [];
  alertGroupForm.status = 1;
  alertGroupForm.remark = "";
  receiversInput.value = "";

  messageChannelForm.name = "";
  messageChannelForm.channelType = "email";
  messageChannelForm.configJson = "{}";
  messageChannelForm.status = 1;
  messageChannelForm.remark = "";

  alertPolicyForm.name = "";
  alertPolicyForm.cpuThreshold = 0;
  alertPolicyForm.memThreshold = 0;
  alertPolicyForm.diskThreshold = 0;
  alertPolicyForm.cpuConsecutiveTimes = 0;
  alertPolicyForm.memConsecutiveTimes = 0;
  alertPolicyForm.heartbeatTimeoutSeconds = 0;
  alertPolicyForm.sendFrequencySeconds = 0;
  alertPolicyForm.status = 1;
  alertPolicyForm.remark = "";

  systemConfigForm.configGroup = "system";
  systemConfigForm.configKey = "";
  systemConfigForm.configValue = "";
  systemConfigForm.valueType = "string";
  systemConfigForm.isEncrypted = 0;
  systemConfigForm.status = 1;
  systemConfigForm.remark = "";

  licenseForm.licenseCode = "";
  licenseForm.serialNumber = "";
  licenseForm.maxAgentCount = 0;
  licenseForm.issuedAt = "";
  licenseForm.expiredAt = "";
  licenseForm.trialDays = 30;
  licenseForm.status = 1;
  licenseForm.remark = "";
}

function openCreate(kind?: string) {
  resetEditorState();
  editorType.value = kind ?? moduleKey.value;
  editorVisible.value = true;
}

function openEdit(record: any, kind?: string) {
  resetEditorState();
  editorType.value = kind ?? moduleKey.value;
  editingId.value = record.id;
  switch (editorType.value) {
    case "file-filter":
      Object.assign(fileFilterForm, record);
      fileFilterPatternsInput.value = record.patterns.join(", ");
      break;
    case "regex-rule":
      Object.assign(regexRuleForm, record);
      regexMappingsInput.value = record.mappings.map((item: { captureIndex: number; tagKey: string }) => `${item.captureIndex}:${item.tagKey}`).join("\n");
      break;
    case "image-processor":
      Object.assign(imageProcessorForm, record);
      break;
    case "file-permissions":
      Object.assign(filePermissionForm, record);
      break;
    case "task-progress":
      Object.assign(taskProgressForm, record);
      break;
    case "alert-groups":
      Object.assign(alertGroupForm, record);
      receiversInput.value = record.receivers.join(", ");
      break;
    case "message-channels":
      Object.assign(messageChannelForm, record);
      break;
    case "alert-policies":
      Object.assign(alertPolicyForm, record);
      break;
    case "system-parameters":
      Object.assign(systemConfigForm, record);
      break;
    case "system-license":
      Object.assign(licenseForm, record);
      break;
  }
  editorVisible.value = true;
}

async function removeRecord(record: any, kind?: string) {
  try {
    const target = kind ?? moduleKey.value;
    switch (target) {
      case "file-filter":
        await deleteFileFilter(record.id);
        break;
      case "regex-rule":
        await deleteRegexRule(record.id);
        break;
      case "image-processor":
        await deleteImageProcessor(record.id);
        break;
      case "file-permissions":
        await deleteFilePermission(record.id);
        break;
      case "task-progress":
        await deleteTaskProgress(record.id);
        break;
      case "alert-groups":
        await deleteAlertGroup(record.id);
        break;
      case "message-channels":
        await deleteMessageChannel(record.id);
        break;
      case "alert-policies":
        await deleteAlertPolicy(record.id);
        break;
      case "system-parameters":
        await deleteSystemConfig(record.id);
        break;
      case "system-license":
        await deleteLicense(record.id);
        break;
    }
    Message.success("Deleted");
    await loadModuleData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Delete failed");
  }
}

async function submitEditor() {
  try {
    switch (editorType.value) {
      case "file-filter": {
        fileFilterForm.patterns = parseListInput(fileFilterPatternsInput.value);
        editingId.value ? await updateFileFilter(editingId.value, fileFilterForm) : await createFileFilter(fileFilterForm);
        break;
      }
      case "regex-rule": {
        regexRuleForm.mappings = regexMappingsInput.value
          .split("\n")
          .map((line) => line.trim())
          .filter(Boolean)
          .map((line) => {
            const [captureIndex = "0", tagKey = ""] = line.split(":");
            return { captureIndex: Number(captureIndex) || 0, tagKey: tagKey.trim() };
          })
          .filter((item) => item.tagKey);
        editingId.value ? await updateRegexRule(editingId.value, regexRuleForm) : await createRegexRule(regexRuleForm);
        break;
      }
      case "image-processor":
        editingId.value ? await updateImageProcessor(editingId.value, imageProcessorForm) : await createImageProcessor(imageProcessorForm);
        break;
      case "file-permissions":
        editingId.value ? await updateFilePermission(editingId.value, filePermissionForm) : await createFilePermission(filePermissionForm);
        break;
      case "task-progress":
        editingId.value ? await updateTaskProgress(editingId.value, taskProgressForm) : await createTaskProgress(taskProgressForm);
        break;
      case "alert-groups":
        alertGroupForm.receivers = parseListInput(receiversInput.value);
        editingId.value ? await updateAlertGroup(editingId.value, alertGroupForm) : await createAlertGroup(alertGroupForm);
        break;
      case "message-channels":
        editingId.value ? await updateMessageChannel(editingId.value, messageChannelForm) : await createMessageChannel(messageChannelForm);
        break;
      case "alert-policies":
        editingId.value ? await updateAlertPolicy(editingId.value, alertPolicyForm) : await createAlertPolicy(alertPolicyForm);
        break;
      case "system-parameters":
        editingId.value ? await updateSystemConfig(editingId.value, systemConfigForm) : await createSystemConfig(systemConfigForm);
        break;
      case "system-license":
        editingId.value ? await updateLicense(editingId.value, licenseForm) : await createLicense(licenseForm);
        break;
    }
    Message.success("Saved");
    editorVisible.value = false;
    await loadModuleData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Save failed");
  }
}

async function loadModuleData() {
  loading.value = true;
  try {
    switch (moduleKey.value) {
      case "extraction-rules":
        [fileFilters.value, regexRules.value, imageProcessors.value] = await Promise.all([fetchFileFilters(), fetchRegexRules(), fetchImageProcessors()]);
        break;
      case "file-logs":
        fileLogs.value = await fetchFileLogs();
        break;
      case "file-permissions":
        [filePermissions.value, users.value, storages.value] = await Promise.all([fetchFilePermissions(), fetchUsers(), fetchStorageTargets()]);
        break;
      case "task-progress":
        taskProgress.value = await fetchTaskProgress();
        break;
      case "alert-groups":
        alertGroups.value = await fetchAlertGroups();
        break;
      case "alert-logs":
        alertLogs.value = await fetchAlertLogs();
        break;
      case "message-channels":
        messageChannels.value = await fetchMessageChannels();
        break;
      case "alert-policies":
        alertPolicies.value = await fetchAlertPolicies();
        break;
      case "system-parameters":
        systemConfigs.value = await fetchSystemConfigs();
        break;
      case "system-license":
        licenses.value = await fetchLicenses();
        break;
    }
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load module data");
  } finally {
    loading.value = false;
  }
}

watch(
  () => route.name,
  () => {
    loadModuleData();
  }
);

onMounted(() => {
  loadModuleData();
});
</script>

<template>
  <div class="page-grid">
    <PageHeader :title="pageMeta.title" :description="pageMeta.description">
      <a-space>
        <a-button type="outline" @click="moduleKey === 'file-logs' || moduleKey === 'alert-logs' ? exportCurrentData() : loadModuleData()">
          {{ moduleKey === "file-logs" || moduleKey === "alert-logs" ? pageMeta.primaryAction : pageMeta.secondaryAction }}
        </a-button>
        <a-button
          v-if="moduleKey !== 'file-logs' && moduleKey !== 'alert-logs'"
          type="primary"
          @click="moduleKey === 'extraction-rules' ? openCreate('file-filter') : openCreate()"
        >
          {{ pageMeta.primaryAction }}
        </a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard v-for="stat in stats" :key="stat.label" :label="stat.label" :value="stat.value" :hint="stat.hint" :icon="stat.icon" />
    </div>

    <template v-if="moduleKey === 'extraction-rules'">
      <div class="content-grid content-grid--equal">
        <a-card class="panel-card" title="File Filters">
          <template #extra><a-button size="small" @click="openCreate('file-filter')">Add</a-button></template>
          <a-table :data="fileFilters" :loading="loading" :pagination="{ pageSize: 5 }" row-key="id">
            <template #columns>
              <a-table-column title="Name" data-index="name" />
              <a-table-column title="Scope" data-index="filterScope" />
              <a-table-column title="List Type" data-index="listType" />
              <a-table-column title="Patterns">
                <template #cell="{ record }">{{ record.patterns.join(", ") }}</template>
              </a-table-column>
              <a-table-column title="Actions" :width="120">
                <template #cell="{ record }">
                  <a-space>
                    <a-button size="mini" @click="openEdit(record, 'file-filter')">Edit</a-button>
                    <a-button size="mini" status="danger" @click="removeRecord(record, 'file-filter')">Delete</a-button>
                  </a-space>
                </template>
              </a-table-column>
            </template>
          </a-table>
        </a-card>

        <a-card class="panel-card" title="Regex Rules">
          <template #extra><a-button size="small" @click="openCreate('regex-rule')">Add</a-button></template>
          <a-table :data="regexRules" :loading="loading" :pagination="{ pageSize: 5 }" row-key="id">
            <template #columns>
              <a-table-column title="Name" data-index="name" />
              <a-table-column title="Source" data-index="sourceField" />
              <a-table-column title="Regex" data-index="regexp" />
              <a-table-column title="Actions" :width="120">
                <template #cell="{ record }">
                  <a-space>
                    <a-button size="mini" @click="openEdit(record, 'regex-rule')">Edit</a-button>
                    <a-button size="mini" status="danger" @click="removeRecord(record, 'regex-rule')">Delete</a-button>
                  </a-space>
                </template>
              </a-table-column>
            </template>
          </a-table>
        </a-card>
      </div>

      <a-card class="panel-card" title="Image Processors">
        <template #extra><a-button size="small" @click="openCreate('image-processor')">Add</a-button></template>
        <a-table :data="imageProcessors" :loading="loading" :pagination="{ pageSize: 6 }" row-key="id">
          <template #columns>
            <a-table-column title="Name" data-index="name" />
            <a-table-column title="Type" data-index="processorType" />
            <a-table-column title="Config JSON" data-index="configJson" />
            <a-table-column title="Actions" :width="120">
              <template #cell="{ record }">
                <a-space>
                  <a-button size="mini" @click="openEdit(record, 'image-processor')">Edit</a-button>
                  <a-button size="mini" status="danger" @click="removeRecord(record, 'image-processor')">Delete</a-button>
                </a-space>
              </template>
            </a-table-column>
          </template>
        </a-table>
      </a-card>
    </template>

    <a-card v-else-if="moduleKey === 'file-logs'" class="panel-card" title="File Operation Logs">
      <a-table :data="fileLogs" :loading="loading" :pagination="{ pageSize: 12 }" row-key="id">
        <template #columns>
          <a-table-column title="Time" data-index="createdAt" :width="180" />
          <a-table-column title="User" data-index="username" :width="140" />
          <a-table-column title="Storage" data-index="storageName" :width="160" />
          <a-table-column title="Operation" data-index="operationType" :width="120" />
          <a-table-column title="Status" data-index="resultStatus" :width="120" />
          <a-table-column title="File Path" data-index="filePath" />
        </template>
      </a-table>
    </a-card>

    <a-card v-else-if="moduleKey === 'file-permissions'" class="panel-card" title="File Permissions">
      <template #extra><a-button size="small" @click="openCreate()">Grant Permission</a-button></template>
      <a-table :data="filePermissions" :loading="loading" :pagination="{ pageSize: 10 }" row-key="id">
        <template #columns>
          <a-table-column title="User" data-index="username" />
          <a-table-column title="Storage" data-index="storageName" />
          <a-table-column title="View"><template #cell="{ record }">{{ record.canView ? "Yes" : "No" }}</template></a-table-column>
          <a-table-column title="Upload"><template #cell="{ record }">{{ record.canUpload ? "Yes" : "No" }}</template></a-table-column>
          <a-table-column title="Download"><template #cell="{ record }">{{ record.canDownload ? "Yes" : "No" }}</template></a-table-column>
          <a-table-column title="Delete"><template #cell="{ record }">{{ record.canDelete ? "Yes" : "No" }}</template></a-table-column>
          <a-table-column title="Actions" :width="120">
            <template #cell="{ record }">
              <a-space>
                <a-button size="mini" @click="openEdit(record)">Edit</a-button>
                <a-button size="mini" status="danger" @click="removeRecord(record)">Delete</a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-card v-else-if="moduleKey === 'task-progress'" class="panel-card" title="Task Progress">
      <template #extra><a-button size="small" @click="openCreate()">Create Task</a-button></template>
      <a-table :data="taskProgress" :loading="loading" :pagination="{ pageSize: 10 }" row-key="id">
        <template #columns>
          <a-table-column title="Task Type" data-index="taskType" />
          <a-table-column title="Status" data-index="status" />
          <a-table-column title="Total" data-index="totalCount" />
          <a-table-column title="Success" data-index="successCount" />
          <a-table-column title="Failed" data-index="failedCount" />
          <a-table-column title="Actions" :width="120">
            <template #cell="{ record }">
              <a-space>
                <a-button size="mini" @click="openEdit(record)">Edit</a-button>
                <a-button size="mini" status="danger" @click="removeRecord(record)">Delete</a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-card v-else-if="moduleKey === 'alert-groups'" class="panel-card" title="Alert Groups">
      <template #extra><a-button size="small" @click="openCreate()">Create Group</a-button></template>
      <a-table :data="alertGroups" :loading="loading" :pagination="{ pageSize: 10 }" row-key="id">
        <template #columns>
          <a-table-column title="Name" data-index="name" />
          <a-table-column title="Receivers"><template #cell="{ record }">{{ record.receivers.join(", ") }}</template></a-table-column>
          <a-table-column title="Status"><template #cell="{ record }">{{ record.status === 1 ? "Enabled" : "Disabled" }}</template></a-table-column>
          <a-table-column title="Actions" :width="120">
            <template #cell="{ record }">
              <a-space>
                <a-button size="mini" @click="openEdit(record)">Edit</a-button>
                <a-button size="mini" status="danger" @click="removeRecord(record)">Delete</a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-card v-else-if="moduleKey === 'alert-logs'" class="panel-card" title="Alert Logs">
      <a-table :data="alertLogs" :loading="loading" :pagination="{ pageSize: 12 }" row-key="id">
        <template #columns>
          <a-table-column title="Time" data-index="createdAt" :width="180" />
          <a-table-column title="Title" data-index="alertTitle" />
          <a-table-column title="Level" data-index="alertLevel" :width="120" />
          <a-table-column title="Send Status" data-index="sendStatus" :width="120" />
          <a-table-column title="Failure Reason" data-index="failureReason" />
        </template>
      </a-table>
    </a-card>

    <a-card v-else-if="moduleKey === 'message-channels'" class="panel-card" title="Message Channels">
      <template #extra><a-button size="small" @click="openCreate()">Add Channel</a-button></template>
      <a-table :data="messageChannels" :loading="loading" :pagination="{ pageSize: 10 }" row-key="id">
        <template #columns>
          <a-table-column title="Name" data-index="name" />
          <a-table-column title="Type" data-index="channelType" />
          <a-table-column title="Config" data-index="configJson" />
          <a-table-column title="Actions" :width="120">
            <template #cell="{ record }">
              <a-space>
                <a-button size="mini" @click="openEdit(record)">Edit</a-button>
                <a-button size="mini" status="danger" @click="removeRecord(record)">Delete</a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-card v-else-if="moduleKey === 'alert-policies'" class="panel-card" title="Alert Policies">
      <template #extra><a-button size="small" @click="openCreate()">Create Policy</a-button></template>
      <a-table :data="alertPolicies" :loading="loading" :pagination="{ pageSize: 10 }" row-key="id">
        <template #columns>
          <a-table-column title="Name" data-index="name" />
          <a-table-column title="CPU" data-index="cpuThreshold" />
          <a-table-column title="Memory" data-index="memThreshold" />
          <a-table-column title="Disk" data-index="diskThreshold" />
          <a-table-column title="Heartbeat Timeout" data-index="heartbeatTimeoutSeconds" />
          <a-table-column title="Actions" :width="120">
            <template #cell="{ record }">
              <a-space>
                <a-button size="mini" @click="openEdit(record)">Edit</a-button>
                <a-button size="mini" status="danger" @click="removeRecord(record)">Delete</a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-card v-else-if="moduleKey === 'system-parameters'" class="panel-card" title="System Parameters">
      <template #extra><a-button size="small" @click="openCreate()">Add Config</a-button></template>
      <a-table :data="systemConfigs" :loading="loading" :pagination="{ pageSize: 10 }" row-key="id">
        <template #columns>
          <a-table-column title="Group" data-index="configGroup" />
          <a-table-column title="Key" data-index="configKey" />
          <a-table-column title="Value" data-index="configValue" />
          <a-table-column title="Type" data-index="valueType" />
          <a-table-column title="Actions" :width="120">
            <template #cell="{ record }">
              <a-space>
                <a-button size="mini" @click="openEdit(record)">Edit</a-button>
                <a-button size="mini" status="danger" @click="removeRecord(record)">Delete</a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-card v-else-if="moduleKey === 'system-license'" class="panel-card" title="License Records">
      <template #extra><a-button size="small" @click="openCreate()">Add License</a-button></template>
      <a-table :data="licenses" :loading="loading" :pagination="{ pageSize: 10 }" row-key="id">
        <template #columns>
          <a-table-column title="Serial Number" data-index="serialNumber" />
          <a-table-column title="Max Agents" data-index="maxAgentCount" />
          <a-table-column title="Issued At" data-index="issuedAt" />
          <a-table-column title="Expired At" data-index="expiredAt" />
          <a-table-column title="Actions" :width="120">
            <template #cell="{ record }">
              <a-space>
                <a-button size="mini" @click="openEdit(record)">Edit</a-button>
                <a-button size="mini" status="danger" @click="removeRecord(record)">Delete</a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-modal :visible="editorVisible" :title="`Edit ${editorType}`" @ok="submitEditor" @cancel="editorVisible = false">
      <a-form :model="{}" layout="vertical">
        <template v-if="editorType === 'file-filter'">
          <a-form-item label="Name"><a-input v-model="fileFilterForm.name" /></a-form-item>
          <a-form-item label="Scope"><a-input v-model="fileFilterForm.filterScope" /></a-form-item>
          <a-form-item label="List Type"><a-input v-model="fileFilterForm.listType" /></a-form-item>
          <a-form-item label="Patterns"><a-textarea v-model="fileFilterPatternsInput" :auto-size="{ minRows: 3, maxRows: 5 }" /></a-form-item>
        </template>

        <template v-else-if="editorType === 'regex-rule'">
          <a-form-item label="Name"><a-input v-model="regexRuleForm.name" /></a-form-item>
          <a-form-item label="Source Field"><a-input v-model="regexRuleForm.sourceField" /></a-form-item>
          <a-form-item label="Regexp"><a-input v-model="regexRuleForm.regexp" /></a-form-item>
          <a-form-item label="Mappings (index:tagKey per line)"><a-textarea v-model="regexMappingsInput" :auto-size="{ minRows: 3, maxRows: 5 }" /></a-form-item>
        </template>

        <template v-else-if="editorType === 'image-processor'">
          <a-form-item label="Name"><a-input v-model="imageProcessorForm.name" /></a-form-item>
          <a-form-item label="Processor Type"><a-input v-model="imageProcessorForm.processorType" /></a-form-item>
          <a-form-item label="Config JSON"><a-textarea v-model="imageProcessorForm.configJson" :auto-size="{ minRows: 4, maxRows: 8 }" /></a-form-item>
        </template>

        <template v-else-if="editorType === 'file-permissions'">
          <a-form-item label="User">
            <a-select v-model="filePermissionForm.userId">
              <a-option v-for="item in users" :key="item.id" :value="item.id">{{ item.username }}</a-option>
            </a-select>
          </a-form-item>
          <a-form-item label="Storage">
            <a-select v-model="filePermissionForm.storageId">
              <a-option v-for="item in storages" :key="item.id" :value="item.id">{{ item.name }}</a-option>
            </a-select>
          </a-form-item>
          <a-form-item label="Permissions">
            <a-checkbox-group direction="vertical">
              <a-checkbox :model-value="filePermissionForm.canView === 1" @change="(value) => (filePermissionForm.canView = value ? 1 : 0)">View</a-checkbox>
              <a-checkbox :model-value="filePermissionForm.canUpload === 1" @change="(value) => (filePermissionForm.canUpload = value ? 1 : 0)">Upload</a-checkbox>
              <a-checkbox :model-value="filePermissionForm.canDownload === 1" @change="(value) => (filePermissionForm.canDownload = value ? 1 : 0)">Download</a-checkbox>
              <a-checkbox :model-value="filePermissionForm.canDelete === 1" @change="(value) => (filePermissionForm.canDelete = value ? 1 : 0)">Delete</a-checkbox>
              <a-checkbox :model-value="filePermissionForm.canBatchDownload === 1" @change="(value) => (filePermissionForm.canBatchDownload = value ? 1 : 0)">Batch Download</a-checkbox>
              <a-checkbox :model-value="filePermissionForm.canDownloadToServer === 1" @change="(value) => (filePermissionForm.canDownloadToServer = value ? 1 : 0)">Download To Server</a-checkbox>
            </a-checkbox-group>
          </a-form-item>
        </template>

        <template v-else-if="editorType === 'task-progress'">
          <a-form-item label="Task Type"><a-input v-model="taskProgressForm.taskType" /></a-form-item>
          <a-form-item label="Status"><a-input v-model="taskProgressForm.status" /></a-form-item>
          <a-form-item label="Payload JSON"><a-textarea v-model="taskProgressForm.payloadJson" :auto-size="{ minRows: 4, maxRows: 8 }" /></a-form-item>
          <a-form-item label="Result JSON"><a-textarea v-model="taskProgressForm.resultJson" :auto-size="{ minRows: 4, maxRows: 8 }" /></a-form-item>
        </template>

        <template v-else-if="editorType === 'alert-groups'">
          <a-form-item label="Name"><a-input v-model="alertGroupForm.name" /></a-form-item>
          <a-form-item label="Receivers"><a-textarea v-model="receiversInput" :auto-size="{ minRows: 3, maxRows: 5 }" /></a-form-item>
        </template>

        <template v-else-if="editorType === 'message-channels'">
          <a-form-item label="Name"><a-input v-model="messageChannelForm.name" /></a-form-item>
          <a-form-item label="Channel Type"><a-input v-model="messageChannelForm.channelType" /></a-form-item>
          <a-form-item label="Config JSON"><a-textarea v-model="messageChannelForm.configJson" :auto-size="{ minRows: 4, maxRows: 8 }" /></a-form-item>
        </template>

        <template v-else-if="editorType === 'alert-policies'">
          <a-form-item label="Name"><a-input v-model="alertPolicyForm.name" /></a-form-item>
          <a-form-item label="CPU Threshold"><a-input-number v-model="alertPolicyForm.cpuThreshold" style="width: 100%" /></a-form-item>
          <a-form-item label="Memory Threshold"><a-input-number v-model="alertPolicyForm.memThreshold" style="width: 100%" /></a-form-item>
          <a-form-item label="Disk Threshold"><a-input-number v-model="alertPolicyForm.diskThreshold" style="width: 100%" /></a-form-item>
          <a-form-item label="Heartbeat Timeout Seconds"><a-input-number v-model="alertPolicyForm.heartbeatTimeoutSeconds" style="width: 100%" /></a-form-item>
        </template>

        <template v-else-if="editorType === 'system-parameters'">
          <a-form-item label="Group"><a-input v-model="systemConfigForm.configGroup" /></a-form-item>
          <a-form-item label="Key"><a-input v-model="systemConfigForm.configKey" /></a-form-item>
          <a-form-item label="Value"><a-textarea v-model="systemConfigForm.configValue" :auto-size="{ minRows: 3, maxRows: 6 }" /></a-form-item>
          <a-form-item label="Type"><a-input v-model="systemConfigForm.valueType" /></a-form-item>
        </template>

        <template v-else-if="editorType === 'system-license'">
          <a-form-item label="Serial Number"><a-input v-model="licenseForm.serialNumber" /></a-form-item>
          <a-form-item label="License Code"><a-textarea v-model="licenseForm.licenseCode" :auto-size="{ minRows: 4, maxRows: 8 }" /></a-form-item>
          <a-form-item label="Max Agent Count"><a-input-number v-model="licenseForm.maxAgentCount" style="width: 100%" /></a-form-item>
          <a-form-item label="Issued At"><a-input v-model="licenseForm.issuedAt" placeholder="YYYY-MM-DD HH:mm:ss" /></a-form-item>
          <a-form-item label="Expired At"><a-input v-model="licenseForm.expiredAt" placeholder="YYYY-MM-DD HH:mm:ss" /></a-form-item>
        </template>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.panel-card :deep(.arco-table .arco-btn-size-mini) {
  padding-inline: 10px;
}
</style>
