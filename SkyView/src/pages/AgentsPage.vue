<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconApps, IconCheckCircle, IconCloseCircle, IconCode, IconDelete, IconEdit, IconEye, IconFolder, IconRefresh } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import MetricBar from "../components/MetricBar.vue";
import { createAgent, deleteAgent, fetchAgentDirectories, fetchAgents, fetchGroups, fetchStorageTargets, updateAgent } from "../api/skybase";
import type { AgentDirectoryEntry, AgentPayload, AgentRecord, GroupRecord, StorageRecord } from "../types";

const agents = ref<AgentRecord[]>([]);
const groups = ref<GroupRecord[]>([]);
const storageTargets = ref<StorageRecord[]>([]);
const loading = ref(false);
const submitting = ref(false);
const modalVisible = ref(false);
const directoryModalVisible = ref(false);
const drawerVisible = ref(false);
const editingId = ref<number | null>(null);
const selectedAgent = ref<AgentRecord | null>(null);
const keyword = ref("");
const statusFilter = ref("all");
const directoryLoading = ref(false);
const directoryPath = ref("");
const directoryEntries = ref<AgentDirectoryEntry[]>([]);
const directoryKeyword = ref("");

const form = reactive<AgentPayload>({
  hostSn: "",
  hostName: "",
  ip: "",
  groupId: 0,
  storageId: 0,
  sourcePaths: [],
  pathPrefix: "",
  version: "",
  status: 1,
  tags: [],
  remark: ""
});

const tagInput = ref("");

const onlineCount = computed(() => agents.value.filter((item) => item.status === 1).length);
const offlineCount = computed(() => agents.value.length - onlineCount.value);
const versionSkew = computed(() => new Set(agents.value.map((item) => item.version)).size);
const modalTitle = computed(() => (editingId.value ? "Edit Agent" : "Register Agent"));
const selectedGroup = computed(() => groups.value.find((item) => item.id === form.groupId) ?? null);
const groupPathPrefixLabel = computed(() => {
  const value = selectedGroup.value?.pathPrefix?.trim();
  if (!value) {
    return "No group path prefix/";
  }
  return value.endsWith("/") ? value : `${value}/`;
});
const filteredDirectoryEntries = computed(() => {
  const query = directoryKeyword.value.trim().toLowerCase();
  if (!query) {
    return directoryEntries.value;
  }
  return directoryEntries.value.filter((item) => item.name.toLowerCase().includes(query) || item.path.toLowerCase().includes(query));
});

const filteredAgents = computed(() =>
  agents.value.filter((item) => {
    const query = keyword.value.trim().toLowerCase();
    const matchesKeyword = !query || item.hostName.toLowerCase().includes(query) || item.hostSn.toLowerCase().includes(query) || item.ip.toLowerCase().includes(query);
    const matchesStatus = statusFilter.value === "all" || String(item.status) === statusFilter.value;
    return matchesKeyword && matchesStatus;
  })
);

function groupName(groupId: number) {
  return groups.value.find((item) => item.id === groupId)?.name ?? "Unassigned";
}

function storageName(storageId: number) {
  return storageTargets.value.find((item) => item.id === storageId)?.name ?? "Unknown";
}

function statusMeta(status: number) {
  if (status === 1) {
    return { label: "Online", color: "green" as const };
  }
  if (status === 2) {
    return { label: "Paused", color: "orange" as const };
  }
  return { label: "Offline", color: "gray" as const };
}

function ensureTrailingSlash(value: string) {
  return value.endsWith("/") ? value : `${value}/`;
}

function buildTargetPathPrefix(groupPathPrefix: string, agentPathPrefix: string) {
  const normalizedGroupPathPrefix = groupPathPrefix.trim();
  const normalizedAgentPathPrefix = agentPathPrefix.trim();

  if (!normalizedGroupPathPrefix && !normalizedAgentPathPrefix) {
    return "";
  }
  if (!normalizedGroupPathPrefix) {
    return ensureTrailingSlash(normalizedAgentPathPrefix);
  }
  if (!normalizedAgentPathPrefix) {
    return ensureTrailingSlash(normalizedGroupPathPrefix);
  }

  return ensureTrailingSlash(
    `${ensureTrailingSlash(normalizedGroupPathPrefix)}${normalizedAgentPathPrefix.replace(/^\/+/, "")}`
  );
}

function extractAgentPathPrefix(fullPathPrefix: string, groupPathPrefix: string) {
  const normalizedFullPathPrefix = fullPathPrefix.trim().replace(/\/+$/, "");
  const normalizedGroupPathPrefix = groupPathPrefix.trim().replace(/\/+$/, "");

  if (!normalizedGroupPathPrefix) {
    return normalizedFullPathPrefix;
  }
  if (normalizedFullPathPrefix === normalizedGroupPathPrefix) {
    return "";
  }

  const groupPrefixWithSlash = ensureTrailingSlash(normalizedGroupPathPrefix);
  if (!normalizedFullPathPrefix.startsWith(groupPrefixWithSlash)) {
    return normalizedFullPathPrefix;
  }

  return normalizedFullPathPrefix.slice(groupPrefixWithSlash.length).replace(/^\/+/, "");
}

function resetForm() {
  editingId.value = null;
  form.hostSn = "";
  form.hostName = "";
  form.ip = "";
  form.groupId = groups.value[0]?.id ?? 0;
  form.storageId = storageTargets.value[0]?.id ?? 0;
  form.sourcePaths = [];
  form.pathPrefix = "";
  form.version = "";
  form.status = 1;
  form.tags = [];
  form.remark = "";
  tagInput.value = "";
}

function openCreateModal() {
  resetForm();
  syncAgentDefaults();
  modalVisible.value = true;
}

function openEditModal(record: AgentRecord) {
  editingId.value = record.id;
  form.hostSn = record.hostSn;
  form.hostName = record.hostName;
  form.ip = record.ip;
  form.groupId = record.groupId;
  form.storageId = record.storageId;
  form.sourcePaths = [...record.sourcePaths];
  form.pathPrefix = extractAgentPathPrefix(record.pathPrefix, groups.value.find((item) => item.id === record.groupId)?.pathPrefix ?? "");
  form.version = record.version;
  form.status = record.status;
  form.tags = [...record.tags];
  form.remark = record.remark;
  tagInput.value = record.tags.join(", ");
  modalVisible.value = true;
}

function closeModal() {
  modalVisible.value = false;
  directoryModalVisible.value = false;
  directoryPath.value = "";
  directoryEntries.value = [];
  resetForm();
}

function openDetail(agent: AgentRecord) {
  selectedAgent.value = agent;
  drawerVisible.value = true;
}

function syncAgentDefaults() {
  const group = groups.value.find((item) => item.id === form.groupId);
  if (group) {
    if (!form.storageId) {
      form.storageId = group.storageId;
    }
  }
}

async function loadData() {
  loading.value = true;
  try {
    [agents.value, groups.value, storageTargets.value] = await Promise.all([
      fetchAgents(),
      fetchGroups(),
      fetchStorageTargets()
    ]);
    if (!form.groupId) {
      form.groupId = groups.value[0]?.id ?? 0;
    }
    if (!form.storageId) {
      form.storageId = storageTargets.value[0]?.id ?? 0;
    }
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load agents");
  } finally {
    loading.value = false;
  }
}

async function loadAgentDirectoryLevel(path = "") {
  if (!editingId.value) {
    Message.error("Select an existing agent before browsing directories");
    return;
  }
  directoryLoading.value = true;
  try {
    directoryEntries.value = await fetchAgentDirectories(editingId.value, path);
    directoryPath.value = path;
    directoryKeyword.value = "";
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load agent directories");
  } finally {
    directoryLoading.value = false;
  }
}

async function openDirectoryModal() {
  directoryModalVisible.value = true;
  await loadAgentDirectoryLevel(directoryPath.value);
}

function parentDirectory(path: string) {
  const trimmed = path.trim();
  if (!trimmed || trimmed === "/") {
    return "";
  }
  const unixPath = trimmed.replace(/\\/g, "/").replace(/\/+$/, "");
  if (/^[A-Za-z]:$/.test(unixPath)) {
    return "";
  }
  const lastSlash = unixPath.lastIndexOf("/");
  if (lastSlash <= 0) {
    return "";
  }
  return unixPath.slice(0, lastSlash);
}

async function navigateToParentDirectory() {
  await loadAgentDirectoryLevel(parentDirectory(directoryPath.value));
}

function toggleSourcePath(path: string) {
  const index = form.sourcePaths.indexOf(path);
  if (index >= 0) {
    form.sourcePaths.splice(index, 1);
    return;
  }
  form.sourcePaths.push(path);
  form.sourcePaths.sort((left, right) => left.localeCompare(right));
}

function removeSourcePath(path: string) {
  form.sourcePaths = form.sourcePaths.filter((item) => item !== path);
}

async function submitAgent() {
  if (!form.hostName.trim()) {
    Message.error("Host name is required");
    return;
  }

  submitting.value = true;
  try {
    const normalizedHostName = form.hostName.trim();
    const normalizedIp = form.ip.trim();
    const groupPathPrefix = selectedGroup.value?.pathPrefix ?? "";
    const payload: AgentPayload = {
      ...form,
      id: editingId.value ?? undefined,
      hostSn: form.hostSn.trim() || normalizedHostName || normalizedIp,
      hostName: normalizedHostName,
      ip: normalizedIp,
      sourcePaths: [...form.sourcePaths],
      pathPrefix: buildTargetPathPrefix(groupPathPrefix, form.pathPrefix),
      version: form.version.trim(),
      remark: form.remark.trim(),
      tags: tagInput.value.split(",").map((item) => item.trim()).filter(Boolean)
    };

    if (editingId.value) {
      await updateAgent(editingId.value, payload);
      Message.success("Agent updated");
    } else {
      await createAgent(payload);
      Message.success("Agent created");
    }
    closeModal();
    await loadData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to save agent");
  } finally {
    submitting.value = false;
  }
}

async function removeAgent(record: AgentRecord) {
  try {
    await deleteAgent(record.id);
    Message.success(`Deleted agent ${record.hostName}`);
    if (selectedAgent.value?.id === record.id) {
      drawerVisible.value = false;
      selectedAgent.value = null;
    }
    await loadData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to delete agent");
  }
}

onMounted(() => {
  loadData();
});
</script>

<template>
  <div class="page-grid">
    <PageHeader title="" description="">
      <a-space>
        <a-button @click="loadData">
          <template #icon><IconRefresh /></template>
          Refresh
        </a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Total Agents" :value="agents.length" hint="Managed collector nodes" :icon="IconApps" />
      <StatCard label="Online" :value="onlineCount" hint="Heartbeat currently active" :icon="IconCheckCircle" />
      <StatCard label="Offline" :value="offlineCount" hint="Requires attention" :icon="IconCloseCircle" />
      <StatCard label="Version Spread" :value="versionSkew" hint="Distinct package versions" :icon="IconCode" />
    </div>

    <a-card class="panel-card">
      <div class="toolbar-row agent-toolbar">
        <a-input-search v-model="keyword" placeholder="Search by host name, SN, or IP" allow-clear />
        <a-select v-model="statusFilter" style="width: 220px">
          <a-option value="all">All statuses</a-option>
          <a-option value="1">Online</a-option>
          <a-option value="2">Paused</a-option>
          <a-option value="0">Offline</a-option>
        </a-select>
      </div>
    </a-card>

    <a-card class="panel-card" title="Agent Inventory">
      <a-table :data="filteredAgents" :loading="loading" :pagination="{ pageSize: 6 }" row-key="id">
        <template #columns>
          <a-table-column title="Host Name" data-index="hostName" />
          <a-table-column title="IP Address" data-index="ip" />
          <a-table-column title="Version" data-index="version" />
          <a-table-column title="Status">
            <template #cell="{ record }">
              <a-tag :color="statusMeta(record.status).color">{{ statusMeta(record.status).label }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Group">
            <template #cell="{ record }">{{ groupName(record.groupId) }}</template>
          </a-table-column>
          <a-table-column title="Storage">
            <template #cell="{ record }">{{ storageName(record.storageId) }}</template>
          </a-table-column>
          <a-table-column title="Last Heartbeat" data-index="lastAccessTime" />
          <a-table-column title="Actions" :width="140">
            <template #cell="{ record }">
              <a-space size="mini">
                <a-tooltip content="Inspect">
                  <a-button class="action-icon-button" type="text" @click="openDetail(record)">
                    <template #icon><IconEye /></template>
                  </a-button>
                </a-tooltip>
                <a-tooltip content="Edit">
                  <a-button class="action-icon-button" type="text" @click="openEditModal(record)">
                    <template #icon><IconEdit /></template>
                  </a-button>
                </a-tooltip>
                <a-popconfirm content="Delete this agent?" @ok="removeAgent(record)">
                  <a-tooltip content="Delete">
                    <a-button class="action-icon-button action-icon-button--danger" type="text" status="danger">
                      <template #icon><IconDelete /></template>
                    </a-button>
                  </a-tooltip>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-modal
      :visible="modalVisible"
      :title="modalTitle"
      :confirm-loading="submitting"
      :width="980"
      unmount-on-close
      @ok="submitAgent"
      @cancel="closeModal"
    >
      <a-form :model="form" layout="vertical" class="modal-bordered-form agent-modal-form">
        <div class="agent-modal-layout">
          <section class="agent-modal-group">
            <div class="agent-modal-group__title">Basic Information</div>
            <a-form-item field="hostName" label="Host Name" required>
              <a-input v-model="form.hostName" placeholder="Enter host name" />
            </a-form-item>
            <a-form-item field="ip" label="IP Address">
              <a-input v-model="form.ip" readonly placeholder="System generated IP address" />
            </a-form-item>
            <a-form-item field="version" label="Version">
              <a-input v-model="form.version" readonly placeholder="System generated version" />
            </a-form-item>
            <a-form-item field="tags" label="Tags">
              <a-input v-model="tagInput" placeholder="Comma separated tags" />
            </a-form-item>
            <a-form-item field="remark" label="Remark">
              <a-textarea v-model="form.remark" :max-length="300" placeholder="Enter remark" :auto-size="{ minRows: 5, maxRows: 8 }" />
            </a-form-item>
          </section>

          <section class="agent-modal-group agent-modal-group--split">
            <div class="agent-modal-group__title">Policy Settings</div>
            <a-form-item field="groupId" label="Group">
              <a-select v-model="form.groupId" :fallback-option="false" @change="syncAgentDefaults">
                <a-option v-for="item in groups" :key="item.id" :value="item.id">{{ item.name }}</a-option>
              </a-select>
            </a-form-item>
            <a-form-item field="pathPrefix" label="Target Path Prefix">
              <a-input v-model="form.pathPrefix" placeholder="Enter target path prefix">
                <template #prepend>{{ groupPathPrefixLabel }}</template>
              </a-input>
            </a-form-item>
            <a-form-item field="status" label="Status">
              <a-radio-group v-model="form.status" type="button">
                <a-radio :value="1">Online</a-radio>
                <a-radio :value="2">Paused</a-radio>
                <a-radio :value="0">Offline</a-radio>
              </a-radio-group>
            </a-form-item>
            <a-form-item field="sourcePaths" label="Source Directories">
              <div class="source-paths-field">
                <div v-if="form.sourcePaths.length" class="source-paths-list">
                  <a-tag v-for="path in form.sourcePaths" :key="path" closable @close="removeSourcePath(path)">
                    {{ path }}
                  </a-tag>
                </div>
                <div v-else class="source-paths-empty">No directories selected</div>
                <a-button @click="openDirectoryModal">
                  <template #icon><IconFolder /></template>
                  Browse Host Directories
                </a-button>
              </div>
            </a-form-item>
          </section>
        </div>
      </a-form>
    </a-modal>

    <a-modal
      :visible="directoryModalVisible"
      title="Browse Source Directories"
      :footer="false"
      :width="760"
      unmount-on-close
      @cancel="directoryModalVisible = false"
    >
      <div class="directory-browser">
        <div class="directory-browser__toolbar">
          <a-input :model-value="directoryPath || '/'" readonly />
          <a-space>
            <a-button @click="loadAgentDirectoryLevel('')">Root</a-button>
            <a-button @click="navigateToParentDirectory">Up</a-button>
            <a-button @click="loadAgentDirectoryLevel(directoryPath)">
              <template #icon><IconRefresh /></template>
              Refresh
            </a-button>
          </a-space>
        </div>

        <a-input-search v-model="directoryKeyword" class="directory-browser__search" allow-clear placeholder="Search folders in current directory" />

        <div class="directory-browser__selected">
          <div class="directory-browser__selected-title">Selected Directories</div>
          <div v-if="form.sourcePaths.length" class="source-paths-list">
            <a-tag v-for="path in form.sourcePaths" :key="path" closable @close="removeSourcePath(path)">
              {{ path }}
            </a-tag>
          </div>
          <div v-else class="source-paths-empty">No directories selected</div>
        </div>

        <a-spin :loading="directoryLoading" style="width: 100%">
          <div v-if="filteredDirectoryEntries.length" class="directory-browser__list">
            <div v-for="entry in filteredDirectoryEntries" :key="entry.path" class="directory-row">
              <div class="directory-row__meta">
                <IconFolder />
                <div class="directory-row__text">
                  <div class="directory-row__name">{{ entry.name }}</div>
                  <div class="directory-row__path">{{ entry.path }}</div>
                </div>
              </div>
              <a-space>
                <a-button size="mini" @click="loadAgentDirectoryLevel(entry.path)">Open</a-button>
                <a-button size="mini" :type="form.sourcePaths.includes(entry.path) ? 'primary' : 'outline'" @click="toggleSourcePath(entry.path)">
                  {{ form.sourcePaths.includes(entry.path) ? "Selected" : "Select" }}
                </a-button>
              </a-space>
            </div>
          </div>
          <a-empty v-else :description="directoryKeyword ? 'No matching folders found' : 'No directories found'" />
        </a-spin>
      </div>
    </a-modal>

    <a-drawer :visible="drawerVisible" width="480px" @cancel="drawerVisible = false" @ok="drawerVisible = false">
      <template #title>Agent Detail</template>
      <div v-if="selectedAgent" class="drawer-stack">
        <div class="detail-meta">
          <div class="detail-meta__title">{{ selectedAgent.hostName }}</div>
          <div class="detail-meta__subtitle">{{ selectedAgent.ip }} · {{ selectedAgent.version }}</div>
        </div>

        <a-card class="panel-card panel-card--subtle" title="Runtime Status">
          <MetricBar label="CPU Usage" :value="selectedAgent.cpu" tone="warning" />
          <MetricBar label="Memory Usage" :value="selectedAgent.mem" tone="default" />
        </a-card>

        <a-card class="panel-card panel-card--subtle" title="Storage Metrics">
          <div v-if="selectedAgent.storage.length" v-for="item in selectedAgent.storage" :key="item.path" class="storage-metric">
            <div class="storage-metric__header">
              <span>{{ item.path }}</span>
              <span>{{ item.used }} / {{ item.total }} GB</span>
            </div>
            <MetricBar label="Used" :value="item.total ? Math.round((item.used / item.total) * 100) : 0" tone="danger" />
          </div>
          <a-empty v-else description="No storage metrics available" />
        </a-card>

        <a-card class="panel-card panel-card--subtle" title="Policy Snapshot">
          <div class="kv-list">
            <div class="kv-list__row"><span>Group</span><span>{{ groupName(selectedAgent.groupId) }}</span></div>
            <div class="kv-list__row"><span>Storage</span><span>{{ storageName(selectedAgent.storageId) }}</span></div>
            <div class="kv-list__row"><span>Source Directories</span><span>{{ selectedAgent.sourcePaths.join(", ") || "None" }}</span></div>
            <div class="kv-list__row"><span>Path Prefix</span><span>{{ selectedAgent.pathPrefix }}</span></div>
            <div class="kv-list__row"><span>Tags</span><span>{{ selectedAgent.tags.join(", ") || "None" }}</span></div>
          </div>
        </a-card>
      </div>
    </a-drawer>
  </div>
</template>

<style scoped>
.agent-toolbar {
  flex-wrap: nowrap;
  align-items: center;
}

.agent-toolbar :deep(.arco-input-search) {
  flex: 1 1 auto;
  min-width: 0;
}

.agent-modal-form {
  padding-top: 8px;
}

.agent-modal-layout {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 32px;
}

.agent-modal-layout::after {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 50%;
  width: 1px;
  background: var(--color-border-2);
  content: "";
  transform: translateX(-0.5px);
}

.agent-modal-group {
  min-width: 0;
}

.agent-modal-group--split {
  padding-left: 16px;
}

.agent-modal-group__title {
  margin-bottom: 20px;
  color: var(--color-text-1);
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.source-paths-field,
.directory-browser {
  display: grid;
  gap: 12px;
}

.source-paths-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.source-paths-empty {
  color: var(--color-text-3);
  font-size: 13px;
}

.directory-browser__toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  align-items: center;
}

.directory-browser__selected {
  display: grid;
  gap: 8px;
}

.directory-browser__search {
  background: #fff;
  border: 1px solid #c9d2de;
  border-radius: 8px;
  box-shadow: none;
  overflow: hidden;
}

.directory-browser__search:hover {
  border-color: #9fb2c8;
}

.directory-browser__search:focus-within {
  border-color: #3c6ea8;
  box-shadow: 0 0 0 2px rgba(60, 110, 168, 0.12);
}

.directory-browser__search :deep(.arco-input-wrapper),
.directory-browser__search :deep(.arco-input-wrapper:hover),
.directory-browser__search :deep(.arco-input-wrapper.arco-input-focus) {
  background: transparent;
  border: 0;
  box-shadow: none;
}

.directory-browser__selected-title {
  color: var(--color-text-2);
  font-size: 13px;
  font-weight: 600;
}

.directory-browser__list {
  display: grid;
  gap: 10px;
}

.directory-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  padding: 12px 14px;
  border: 1px solid var(--color-border-2);
  border-radius: 10px;
}

.directory-row__meta {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  min-width: 0;
}

.directory-row__text {
  min-width: 0;
}

.directory-row__name {
  color: var(--color-text-1);
  font-weight: 600;
}

.directory-row__path {
  color: var(--color-text-3);
  font-size: 12px;
  word-break: break-all;
}

@media (max-width: 900px) {
  .agent-modal-layout {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .agent-modal-layout::after {
    display: none;
  }

  .agent-modal-group--split {
    border-top: 1px solid var(--color-border-2);
    padding-left: 0;
    padding-top: 20px;
  }

  .directory-browser__toolbar {
    grid-template-columns: 1fr;
  }

  .directory-row {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
