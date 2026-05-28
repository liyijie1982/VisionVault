<script setup lang="ts">
import { computed, ref } from "vue";
import { IconCheckCircle, IconExclamationCircle, IconHistory, IconPauseCircle } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";

type BackupLogRecord = {
  id: number;
  taskName: string;
  source: string;
  tapeTarget: string;
  status: "Success" | "Running" | "Paused" | "Failed";
  message: string;
  operator: string;
  createdAt: string;
};

const logs = ref<BackupLogRecord[]>([
  {
    id: 1,
    taskName: "Camera-A May Incremental Backup",
    source: "Cold Object Vault /visionvault-cold/2026/camera-a",
    tapeTarget: "LTO Library A / LTO9-0001",
    status: "Running",
    message: "Object enumeration finished and the fourth chunk is being written.",
    operator: "system",
    createdAt: "2026-05-26 10:28"
  },
  {
    id: 2,
    taskName: "Night Shift Weekly Archive",
    source: "Factory Archive /srv/archive/factory-a/night-shift",
    tapeTarget: "LTO Library B / LTO8-0013",
    status: "Paused",
    message: "The task was manually paused by operations and will resume in the night window.",
    operator: "ops_admin",
    createdAt: "2026-05-26 07:50"
  },
  {
    id: 3,
    taskName: "Monthly Cold Archive to Tape",
    source: "Cold Object Vault /visionvault-cold/monthly-archive",
    tapeTarget: "LTO Library A / LTO9-0002",
    status: "Success",
    message: "Backup completed and verification passed. The tape has been returned to the library.",
    operator: "system",
    createdAt: "2026-05-24 23:25"
  },
  {
    id: 4,
    taskName: "Production Line B Full Archive",
    source: "Cold Object Vault /visionvault-cold/2026/camera-b",
    tapeTarget: "LTO Library B / LTO8-0015",
    status: "Failed",
    message: "The tape drive reported a media alert. Remount the tape and retry.",
    operator: "ops_admin",
    createdAt: "2026-05-22 16:42"
  }
]);

const successCount = computed(() => logs.value.filter((item) => item.status === "Success").length);
const runningCount = computed(() => logs.value.filter((item) => item.status === "Running").length);
const pausedCount = computed(() => logs.value.filter((item) => item.status === "Paused").length);
const failedCount = computed(() => logs.value.filter((item) => item.status === "Failed").length);
</script>

<template>
  <div class="page-grid">
    <PageHeader title="Backup Logs" description="Track results, execution state, and exception details for object storage to tape backup operations.">
      <a-space>
        <a-button>Export Logs</a-button>
        <a-button type="primary">Refresh Logs</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Successful Records" :value="successCount" hint="Completed and verified successfully" :icon="IconCheckCircle" />
      <StatCard label="Running Records" :value="runningCount" hint="Currently still executing" :icon="IconHistory" />
      <StatCard label="Paused Records" :value="pausedCount" hint="Waiting for manual resume" :icon="IconPauseCircle" />
      <StatCard label="Failed Records" :value="failedCount" hint="Requires troubleshooting" :icon="IconExclamationCircle" />
    </div>

    <a-card class="panel-card" title="Log Details">
      <a-table :data="logs" :pagination="false">
        <template #columns>
          <a-table-column title="Time" data-index="createdAt" :width="160" />
          <a-table-column title="Task Name" data-index="taskName" :width="220" />
          <a-table-column title="Backup Source" data-index="source" />
          <a-table-column title="Target Tape" data-index="tapeTarget" />
          <a-table-column title="Status" :width="100">
            <template #cell="{ record }">
              <a-tag
                :color="record.status === 'Success' ? 'green' : record.status === 'Failed' ? 'red' : record.status === 'Paused' ? 'orange' : 'arcoblue'"
              >
                {{ record.status }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Details">
            <template #cell="{ record }">
              <div>{{ record.message }}</div>
              <div class="log-operator">Operator: {{ record.operator }}</div>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<style scoped>
.log-operator {
  margin-top: 4px;
  color: var(--text-2);
  font-size: 12px;
}
</style>
