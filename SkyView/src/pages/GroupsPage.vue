<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconApps, IconCheckCircle, IconDelete, IconEdit, IconEye, IconInfoCircle, IconMinusCircle, IconSchedule, IconThunderbolt } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { createGroup, deleteGroup, fetchGroups, fetchStorageTargets, updateGroup } from "../api/skybase";
import type { GroupPayload, GroupRecord, StorageRecord } from "../types";

const groups = ref<GroupRecord[]>([]);
const storageTargets = ref<StorageRecord[]>([]);
const loading = ref(false);
const submitting = ref(false);
const modalVisible = ref(false);
const drawerVisible = ref(false);
const editingId = ref<number | null>(null);
const selectedGroup = ref<GroupRecord | null>(null);
const keyword = ref("");
const statusFilter = ref("all");

function createDefaultWorkWindow() {
  return {
    startTime: "00:00",
    endTime: "23:59"
  };
}

const form = reactive<GroupPayload>({
  name: "",
  storageId: 0,
  ipRange: "",
  pathPrefix: "",
  intervalTime: 300,
  delTimeDays: 30,
  transferSpeedLimit: 4,
  workWindows: [createDefaultWorkWindow()],
  fileFilterId: 0,
  regexId: 0,
  imageProcessId: 0,
  alarmGroupId: 0,
  logEnabled: true,
  status: 1
});

const averageSpeedLimit = computed(() =>
  groups.value.length
    ? Math.round(groups.value.reduce((sum, item) => sum + item.transferSpeedLimit, 0) / groups.value.length)
    : 0
);
const alwaysOnCount = computed(
  () =>
    groups.value.filter(
      (item) =>
        item.workWindows.length === 1 &&
        item.workWindows[0]?.startTime === "00:00" &&
        item.workWindows[0]?.endTime === "23:59"
    ).length
);
const alarmGroupOptions = computed(() => {
  const ids = new Set<number>();
  for (const item of groups.value) {
    if (item.alarmGroupId > 0) {
      ids.add(item.alarmGroupId);
    }
  }
  if (form.alarmGroupId > 0) {
    ids.add(form.alarmGroupId);
  }
  return Array.from(ids)
    .sort((left, right) => left - right)
    .map((id) => ({
      id,
      name: `Alarm Group #${id}`
    }));
});
const modalTitle = computed(() => (editingId.value ? "Edit Group" : "Create Group"));
const filteredGroups = computed(() =>
  groups.value.filter((item) => {
    const query = keyword.value.trim().toLowerCase();
    const matchesKeyword =
      !query ||
      item.name.toLowerCase().includes(query) ||
      item.ipRange.toLowerCase().includes(query) ||
      item.pathPrefix.toLowerCase().includes(query) ||
      storageName(item.storageId).toLowerCase().includes(query);
    const matchesStatus = statusFilter.value === "all" || String(item.status) === statusFilter.value;
    return matchesKeyword && matchesStatus;
  })
);

function storageName(storageId: number) {
  return storageTargets.value.find((item) => item.id === storageId)?.name ?? "Unknown";
}

function formatWorkWindows(workWindows: GroupRecord["workWindows"]) {
  return workWindows.map((item) => `${item.startTime} - ${item.endTime}`).join(" / ");
}

function groupStatusMeta(status: number) {
  if (status === 1) {
    return { label: "Active", color: "green" as const };
  }
  return { label: "Disabled", color: "gray" as const };
}

function normalizeWorkWindows() {
  return form.workWindows
    .map((item) => ({
      startTime: item.startTime.trim(),
      endTime: item.endTime.trim()
    }))
    .filter((item) => item.startTime && item.endTime);
}

function addWorkWindow() {
  form.workWindows.push(createDefaultWorkWindow());
}

function removeWorkWindow(index: number) {
  if (form.workWindows.length === 1) {
    form.workWindows.splice(0, 1, createDefaultWorkWindow());
    return;
  }
  form.workWindows.splice(index, 1);
}

function resetForm() {
  editingId.value = null;
  form.name = "";
  form.storageId = storageTargets.value[0]?.id ?? 0;
  form.ipRange = "";
  form.pathPrefix = "";
  form.intervalTime = 300;
  form.delTimeDays = 30;
  form.transferSpeedLimit = 4;
  form.workWindows = [createDefaultWorkWindow()];
  form.fileFilterId = 0;
  form.regexId = 0;
  form.imageProcessId = 0;
  form.alarmGroupId = 0;
  form.logEnabled = true;
  form.status = 1;
}

function openCreateModal() {
  resetForm();
  modalVisible.value = true;
}

function openEditModal(record: GroupRecord) {
  editingId.value = record.id;
  form.name = record.name;
  form.storageId = record.storageId;
  form.ipRange = record.ipRange;
  form.pathPrefix = record.pathPrefix;
  form.intervalTime = record.intervalTime;
  form.delTimeDays = record.delTimeDays;
  form.transferSpeedLimit = record.transferSpeedLimit;
  form.workWindows = record.workWindows.length ? record.workWindows.map((item) => ({ ...item })) : [createDefaultWorkWindow()];
  form.fileFilterId = record.fileFilterId;
  form.regexId = record.regexId;
  form.imageProcessId = record.imageProcessId;
  form.alarmGroupId = record.alarmGroupId;
  form.logEnabled = record.logEnabled;
  form.status = record.status;
  modalVisible.value = true;
}

function openDetail(record: GroupRecord) {
  selectedGroup.value = record;
  drawerVisible.value = true;
}

function closeModal() {
  modalVisible.value = false;
  resetForm();
}

async function loadData() {
  loading.value = true;
  try {
    [groups.value, storageTargets.value] = await Promise.all([fetchGroups(), fetchStorageTargets()]);
    if (!form.storageId) {
      form.storageId = storageTargets.value[0]?.id ?? 0;
    }
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load groups");
  } finally {
    loading.value = false;
  }
}

async function submitGroup() {
  if (!form.name.trim()) {
    Message.error("Group name is required");
    return;
  }

  const workWindows = normalizeWorkWindows();
  if (!workWindows.length || workWindows.length !== form.workWindows.length) {
    Message.error("Each work window must include both start and end time");
    return;
  }

  submitting.value = true;
  try {
    const payload: GroupPayload = {
      ...form,
      name: form.name.trim(),
      ipRange: form.ipRange.trim(),
      pathPrefix: form.pathPrefix.trim(),
      workWindows
    };

    if (editingId.value) {
      await updateGroup(editingId.value, payload);
      Message.success("Group updated");
    } else {
      await createGroup(payload);
      Message.success("Group created");
    }
    closeModal();
    await loadData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to save group");
  } finally {
    submitting.value = false;
  }
}

async function removeGroup(record: GroupRecord) {
  try {
    await deleteGroup(record.id);
    Message.success(`Deleted group ${record.name}`);
    if (selectedGroup.value?.id === record.id) {
      drawerVisible.value = false;
      selectedGroup.value = null;
    }
    await loadData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to delete group");
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
        <a-button @click="loadData">Refresh</a-button>
        <a-button type="primary" @click="openCreateModal">Create Group</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Groups" :value="groups.length" hint="Policy aggregation units" :icon="IconApps" />
      <StatCard label="Logging Enabled" :value="groups.filter((item) => item.logEnabled).length" hint="Indexed sync logging" :icon="IconCheckCircle" />
      <StatCard label="Average Speed Limit" :value="`${averageSpeedLimit} MB/s`" hint="Configured transfer ceiling" :icon="IconThunderbolt" />
      <StatCard label="Always-On Schedules" :value="alwaysOnCount" hint="Single full-day work window" :icon="IconSchedule" />
    </div>

    <a-card class="panel-card">
      <div class="toolbar-row group-toolbar">
        <a-input-search
          v-model="keyword"
          placeholder="Search by group name, path prefix, or storage"
          allow-clear
        />
        <a-select v-model="statusFilter" style="width: 220px">
          <a-option value="all">All statuses</a-option>
          <a-option value="1">Active</a-option>
          <a-option value="0">Disabled</a-option>
        </a-select>
      </div>
    </a-card>

    <a-card class="panel-card" title="Group List">
      <a-table :data="filteredGroups" :loading="loading" :pagination="{ pageSize: 8, showTotal: true }" row-key="id">
        <template #columns>
          <a-table-column title="Group Name" data-index="name" :width="180" />
          <a-table-column title="Storage" :width="180">
            <template #cell="{ record }">{{ storageName(record.storageId) }}</template>
          </a-table-column>
          <a-table-column title="Path Prefix" data-index="pathPrefix" :width="180" />
          <a-table-column title="Interval Time" :width="120">
            <template #cell="{ record }">{{ record.intervalTime }} s</template>
          </a-table-column>
          <a-table-column title="Retention" :width="120">
            <template #cell="{ record }">{{ record.delTimeDays }} days</template>
          </a-table-column>
          <a-table-column title="Speed Limit" :width="130">
            <template #cell="{ record }">{{ record.transferSpeedLimit }} MB/s</template>
          </a-table-column>
          <a-table-column title="Alarm Group" :width="130">
            <template #cell="{ record }">{{ record.alarmGroupId ? `#${record.alarmGroupId}` : "None" }}</template>
          </a-table-column>
          <a-table-column title="Log Indexing" :width="120">
            <template #cell="{ record }">
              <a-tag :color="record.logEnabled ? 'green' : 'gray'">{{ record.logEnabled ? "Enabled" : "Disabled" }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Status" :width="110">
            <template #cell="{ record }">
              <a-tag :color="record.status === 1 ? 'green' : 'gray'">{{ record.status === 1 ? "Active" : "Disabled" }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Updated At" data-index="updatedAt" :width="180" />
          <a-table-column title="Actions" :width="140" fixed="right">
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
                <a-popconfirm content="Delete this group?" @ok="removeGroup(record)">
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

    <a-modal :visible="modalVisible" :title="modalTitle" :confirm-loading="submitting" :width="980" unmount-on-close @ok="submitGroup" @cancel="closeModal">
      <a-form :model="form" layout="vertical" class="modal-bordered-form group-modal-form">
        <div class="group-modal-layout">
          <section class="group-modal-group">
            <div class="group-modal-group__title">Basic Information</div>
            <a-form-item field="name" required>
              <template #label>
                <span class="group-form-label">
                  <span>Group Name</span>
                  <a-tooltip content="Human-readable group name used to identify this policy set in the UI and when assigning agents.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-input v-model="form.name" placeholder="Enter group name" />
            </a-form-item>
            <a-form-item field="storageId" required>
              <template #label>
                <span class="group-form-label">
                  <span>Storage</span>
                  <a-tooltip content="Default storage target used by agents in this group when no agent-level override is specified.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-select v-model="form.storageId">
                <a-option v-for="item in storageTargets" :key="item.id" :value="item.id">{{ item.name }}</a-option>
              </a-select>
            </a-form-item>
            <a-form-item field="ipRange">
              <template #label>
                <span class="group-form-label">
                  <span>IP Range</span>
                  <a-tooltip content="CIDR or range hint used to associate matching agents with this group during registration and heartbeat.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-input v-model="form.ipRange" placeholder="Example: 10.16.1.0/24" />
            </a-form-item>
            <a-form-item field="pathPrefix">
              <template #label>
                <span class="group-form-label">
                  <span>Path Prefix</span>
                  <a-tooltip content="Prefix appended to files or tasks from this group to keep output paths organized by business domain.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-input v-model="form.pathPrefix" placeholder="Enter path prefix" />
            </a-form-item>
            <a-form-item field="alarmGroupId">
              <template #label>
                <span class="group-form-label">
                  <span>Alarm Group</span>
                  <a-tooltip content="Select the alert recipient group that should receive notifications for this group's runtime and transfer issues.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-select v-model="form.alarmGroupId" placeholder="Select an alarm group">
                <a-option :value="0">None</a-option>
                <a-option v-for="item in alarmGroupOptions" :key="item.id" :value="item.id">{{ item.name }}</a-option>
              </a-select>
            </a-form-item>
          </section>

          <section class="group-modal-group group-modal-group--split">
            <div class="group-modal-group__title">Policy Settings</div>
            <a-form-item field="intervalTime">
              <template #label>
                <span class="group-form-label">
                  <span>Interval Time</span>
                  <a-tooltip content="Execution interval in seconds for scheduled work generated from this group policy.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-input-number v-model="form.intervalTime" :min="1" style="width: 100%" />
              <span class="group-form-unit">seconds</span>
            </a-form-item>
            <a-form-item field="delTimeDays">
              <template #label>
                <span class="group-form-label">
                  <span>Retention Days</span>
                  <a-tooltip content="How many days data should be retained before cleanup rules are allowed to remove it.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-input-number v-model="form.delTimeDays" :min="0" style="width: 100%" />
              <span class="group-form-unit">days</span>
            </a-form-item>
            <a-form-item field="transferSpeedLimit">
              <template #label>
                <span class="group-form-label">
                  <span>Transfer Speed Limit</span>
                  <a-tooltip content="Maximum transfer throughput for this group. The current UI treats the value as MB/s.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-input-number v-model="form.transferSpeedLimit" :min="0" style="width: 100%" />
              <span class="group-form-unit">MB/s</span>
            </a-form-item>
            <a-form-item field="workWindows" required>
              <template #label>
                <span class="group-form-label">
                  <span>Work Windows</span>
                  <a-tooltip content="Allowed execution periods for this group. Add multiple start/end pairs to define separate operating windows in one day.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <div class="stack-list">
                <div v-for="(item, index) in form.workWindows" :key="index" class="group-work-window-row">
                  <a-input v-model="item.startTime" class="group-work-window-row__field" placeholder="Work Start, e.g. 08:00" />
                  <a-input v-model="item.endTime" class="group-work-window-row__field" placeholder="Work End, e.g. 20:00" />
                  <a-tooltip content="Remove this work window">
                    <a-button class="action-icon-button action-icon-button--danger" type="text" status="danger" @click="removeWorkWindow(index)">
                      <template #icon><IconMinusCircle /></template>
                    </a-button>
                  </a-tooltip>
                </div>
                <a-button @click="addWorkWindow">Add Work Window</a-button>
              </div>
            </a-form-item>
            <a-form-item field="status">
              <template #label>
                <span class="group-form-label">
                  <span>Status</span>
                  <a-tooltip content="Controls whether this group policy is active and available for agent matching and scheduling.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-radio-group v-model="form.status" type="button">
                <a-radio :value="1">Active</a-radio>
                <a-radio :value="0">Disabled</a-radio>
              </a-radio-group>
            </a-form-item>
            <a-form-item field="logEnabled">
              <template #label>
                <span class="group-form-label">
                  <span>Log Indexing</span>
                  <a-tooltip content="Enables indexed logging for transfer and execution activity generated by this group.">
                    <IconInfoCircle class="group-form-label__icon" />
                  </a-tooltip>
                </span>
              </template>
              <a-switch v-model="form.logEnabled" />
            </a-form-item>
          </section>
        </div>
      </a-form>
    </a-modal>

    <a-drawer :visible="drawerVisible" width="480px" @cancel="drawerVisible = false" @ok="drawerVisible = false">
      <template #title>Group Detail</template>
      <div v-if="selectedGroup" class="drawer-stack">
        <div class="detail-meta">
          <div class="detail-meta__title">{{ selectedGroup.name }}</div>
          <div class="detail-meta__subtitle">{{ storageName(selectedGroup.storageId) }} · {{ selectedGroup.updatedAt }}</div>
        </div>

        <a-card class="panel-card panel-card--subtle" title="Policy Summary">
          <div class="kv-list">
            <div class="kv-list__row"><span>Storage</span><span>{{ storageName(selectedGroup.storageId) }}</span></div>
            <div class="kv-list__row"><span>IP Range</span><span>{{ selectedGroup.ipRange || "Not set" }}</span></div>
            <div class="kv-list__row"><span>Path Prefix</span><span>{{ selectedGroup.pathPrefix || "Not set" }}</span></div>
            <div class="kv-list__row"><span>Alarm Group</span><span>{{ selectedGroup.alarmGroupId ? `#${selectedGroup.alarmGroupId}` : "None" }}</span></div>
            <div class="kv-list__row">
              <span>Status</span>
              <span><a-tag :color="groupStatusMeta(selectedGroup.status).color">{{ groupStatusMeta(selectedGroup.status).label }}</a-tag></span>
            </div>
          </div>
        </a-card>

        <a-card class="panel-card panel-card--subtle" title="Execution Settings">
          <div class="kv-list">
            <div class="kv-list__row"><span>Interval Time</span><span>{{ selectedGroup.intervalTime }} s</span></div>
            <div class="kv-list__row"><span>Retention Days</span><span>{{ selectedGroup.delTimeDays }} days</span></div>
            <div class="kv-list__row"><span>Transfer Speed Limit</span><span>{{ selectedGroup.transferSpeedLimit }} MB/s</span></div>
            <div class="kv-list__row"><span>Work Windows</span><span>{{ formatWorkWindows(selectedGroup.workWindows) || "Not set" }}</span></div>
            <div class="kv-list__row"><span>Log Indexing</span><span>{{ selectedGroup.logEnabled ? "Enabled" : "Disabled" }}</span></div>
          </div>
        </a-card>
      </div>
    </a-drawer>
  </div>
</template>

<style scoped>
.group-toolbar {
  flex-wrap: nowrap;
  align-items: center;
}

.group-toolbar :deep(.arco-input-search),
.group-toolbar :deep(.arco-input-wrapper) {
  flex: 1;
  min-width: 0;
}

.group-modal-form {
  padding-top: 8px;
}

.group-modal-layout {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 32px;
}

.group-modal-layout::after {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 50%;
  width: 1px;
  background: var(--color-border-2);
  content: "";
  transform: translateX(-0.5px);
}

.group-modal-group {
  min-width: 0;
}

.group-modal-group--split {
  padding-left: 16px;
}

.group-modal-group__title {
  margin-bottom: 20px;
  color: var(--color-text-1);
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.group-form-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.group-form-label__icon {
  color: var(--color-text-3);
  cursor: help;
}

.group-form-unit {
  margin-left: 8px;
  color: var(--color-text-3);
  font-size: 12px;
}

.group-work-window-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.group-work-window-row__field {
  flex: 1;
}

.drawer-stack {
  display: grid;
  gap: 16px;
}

.detail-meta {
  display: grid;
  gap: 4px;
}

.detail-meta__title {
  color: var(--color-text-1);
  font-size: 20px;
  font-weight: 700;
}

.detail-meta__subtitle {
  color: var(--color-text-3);
  font-size: 13px;
}

.kv-list {
  display: grid;
  gap: 12px;
}

.kv-list__row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
}

.kv-list__row span:first-child {
  color: var(--color-text-3);
}

.kv-list__row span:last-child {
  color: var(--color-text-1);
  font-weight: 500;
  text-align: right;
  word-break: break-word;
}

@media (max-width: 900px) {
  .group-modal-layout {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .group-modal-layout::after {
    display: none;
  }

  .group-modal-group--split {
    border-top: 1px solid var(--color-border-2);
    padding-left: 0;
    padding-top: 20px;
  }

  .group-work-window-row {
    align-items: stretch;
    flex-direction: column;
  }

  .kv-list__row {
    flex-direction: column;
    gap: 6px;
  }

  .kv-list__row span:last-child {
    text-align: left;
  }
}
</style>
