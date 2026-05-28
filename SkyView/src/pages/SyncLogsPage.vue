<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconCalendarClock, IconCloseCircle, IconFile, IconStorage } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { fetchAgents, fetchSyncLogs } from "../api/skybase";
import type { AgentRecord, SyncLogRecord } from "../types";

const logs = ref<SyncLogRecord[]>([]);
const agents = ref<AgentRecord[]>([]);
const loading = ref(false);
const keyword = ref("");
const resultFilter = ref("all");
const currentPage = ref(1);
const pageSize = 10;

const hostNameByIp = computed(() => {
  const pairs = agents.value.map((item) => [item.ip, item.hostName] as const);
  return new Map<string, string>(pairs);
});

const sortedLogs = computed(() =>
  [...logs.value]
    .map((item) => ({
      ...item,
      hostName: item.hostName || hostNameByIp.value.get(item.agentIp) || "",
      formattedStartTime: formatDateTime(item.startTime),
      formattedCommitTime: formatDateTime(item.commitTime),
      formattedFileSize: formatSizeInGb(parseSizeToBytes(item.fileSize))
    }))
    .sort((left, right) => parseDateValue(right.commitTime) - parseDateValue(left.commitTime))
);

const filteredLogs = computed(() => {
  const query = keyword.value.trim().toLowerCase();
  if (!query) {
    return sortedLogs.value;
  }
  return sortedLogs.value.filter((item) => item.agentIp.toLowerCase().includes(query) || item.hostName.toLowerCase().includes(query));
});

const todayLogs = computed(() => filteredLogs.value.filter((item) => isTodayByCommitTime(item.commitTime)));
const pagedLogs = computed(() => {
  const start = (currentPage.value - 1) * pageSize;
  return filteredLogs.value.slice(start, start + pageSize);
});

const runsToday = computed(() => todayLogs.value.length);
const transferredFiles = computed(() => todayLogs.value.reduce((sum, item) => sum + item.fileCount, 0));
const transferredSize = computed(() => formatSizeInGb(todayLogs.value.reduce((sum, item) => sum + parseSizeToBytes(item.fileSize), 0)));
const errorCount = computed(() => todayLogs.value.reduce((sum, item) => sum + item.errorCount, 0));

async function loadLogs() {
  loading.value = true;
  try {
    const [syncLogs, agentRecords] = await Promise.all([fetchSyncLogs(resultFilter.value), fetchAgents()]);
    agents.value = agentRecords;
    logs.value = syncLogs;
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load sync logs");
  } finally {
    loading.value = false;
  }
}

function exportLogs() {
  const rows = [
    ["Host Name", "Agent IP", "Start Time", "File Count", "File Size", "Error Count", "Log Path", "Commit Time"],
    ...filteredLogs.value.map((item) => [
      item.hostName || hostNameByIp.value.get(item.agentIp) || "",
      item.agentIp,
      formatDateTime(item.startTime),
      String(item.fileCount),
      formatSizeInGb(parseSizeToBytes(item.fileSize)),
      String(item.errorCount),
      item.logPath,
      formatDateTime(item.commitTime)
    ])
  ];
  const csv = rows.map((row) => row.map(escapeCsvCell).join(",")).join("\n");
  const blob = new Blob([`\uFEFF${csv}`], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  const dateStamp = new Date().toISOString().slice(0, 10);
  link.href = url;
  link.download = `sync-logs-${dateStamp}.csv`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

function escapeCsvCell(value: string) {
  const normalized = value.replace(/"/g, "\"\"");
  return `"${normalized}"`;
}

function parseDateValue(value: string) {
  const trimmed = value.trim();
  if (!trimmed) {
    return 0;
  }
  const match = trimmed.match(
    /^(\d{4})-(\d{2})-(\d{2})(?:[ T](\d{2})(?::(\d{2}))?(?::(\d{2}))?)?$/
  );
  if (match) {
    const [, year, month, day, hour = "00", minute = "00", second = "00"] = match;
    return new Date(
      Number(year),
      Number(month) - 1,
      Number(day),
      Number(hour),
      Number(minute),
      Number(second)
    ).getTime();
  }
  const parsed = new Date(trimmed).getTime();
  return Number.isNaN(parsed) ? 0 : parsed;
}

function isTodayByCommitTime(value: string) {
  const timestamp = parseDateValue(value);
  if (!timestamp) {
    return false;
  }
  const date = new Date(timestamp);
  const now = new Date();
  return (
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate()
  );
}

function formatDateTime(value: string) {
  const timestamp = parseDateValue(value);
  if (!timestamp) {
    return value || "-";
  }
  const date = new Date(timestamp);
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  const hours = `${date.getHours()}`.padStart(2, "0");
  const minutes = `${date.getMinutes()}`.padStart(2, "0");
  const seconds = `${date.getSeconds()}`.padStart(2, "0");
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}

function parseSizeToBytes(value: string) {
  const normalized = value.trim().toUpperCase();
  const match = normalized.match(/^([\d.]+)\s*([KMGT]?B)$/);
  if (!match) {
    return 0;
  }
  const size = Number(match[1]);
  if (Number.isNaN(size)) {
    return 0;
  }
  const unitMap: Record<string, number> = {
    B: 1,
    KB: 1024,
    MB: 1024 ** 2,
    GB: 1024 ** 3,
    TB: 1024 ** 4
  };
  return size * (unitMap[match[2]] ?? 1);
}

function formatSizeInGb(bytes: number) {
  if (!bytes) {
    return "0.00 GB";
  }
  return `${(bytes / 1024 ** 3).toFixed(2)} GB`;
}

function handlePageChange(page: number) {
  currentPage.value = page;
}

onMounted(() => {
  loadLogs();
});

watch(resultFilter, () => {
  currentPage.value = 1;
  loadLogs();
});
</script>

<template>
  <div class="page-grid">
    <PageHeader
      title=""
      description=""
    >
      <a-space>
        <a-button :loading="loading" @click="loadLogs">Refresh</a-button>
        <a-button type="primary" @click="exportLogs">Export Logs</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Runs Today" :value="runsToday" hint="Today's task count" :icon="IconCalendarClock" />
      <StatCard label="Transferred Files" :value="transferredFiles" hint="Today's transferred file count" :icon="IconFile" />
      <StatCard label="Transferred Size" :value="transferredSize" hint="Today's transferred size" :icon="IconStorage" />
      <StatCard label="Error Count" :value="errorCount" hint="Today's failed file count" :icon="IconCloseCircle" />
    </div>

    <a-card class="panel-card">
      <div class="toolbar-row sync-logs-toolbar">
        <a-input-search
          v-model="keyword"
          placeholder="Search by IP or Host Name"
          allow-clear
          @search="currentPage = 1"
          @clear="currentPage = 1"
          @input="currentPage = 1"
        />
        <a-select v-model="resultFilter" style="width: 220px">
          <a-option value="all">All tasks</a-option>
          <a-option value="failed">Failed tasks</a-option>
          <a-option value="success">Successful tasks</a-option>
        </a-select>
      </div>
    </a-card>

    <a-card class="panel-card" title="Commit Records">
      <a-table
        :data="pagedLogs"
        :loading="loading"
        :pagination="{
          total: filteredLogs.length,
          current: currentPage,
          pageSize,
          showTotal: true
        }"
        @page-change="handlePageChange"
      >
        <template #columns>
          <a-table-column title="Host Name" data-index="hostName" />
          <a-table-column title="Agent IP" data-index="agentIp" />
          <a-table-column title="File Count" data-index="fileCount" />
          <a-table-column title="File Size" data-index="formattedFileSize" />
          <a-table-column title="Error Count" data-index="errorCount" />
          <a-table-column title="Start Time" data-index="formattedStartTime" />
          <a-table-column title="Commit Time" data-index="formattedCommitTime" />
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<style scoped>
.sync-logs-toolbar {
  flex-wrap: nowrap;
  align-items: center;
}

.sync-logs-toolbar :deep(.arco-input-search),
.sync-logs-toolbar :deep(.arco-input-wrapper) {
  flex: 1;
  min-width: 0;
}
</style>
