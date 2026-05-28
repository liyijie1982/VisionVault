<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconArchive, IconPauseCircle, IconPlayCircle, IconStorage } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";

type SourceDirectory = {
  label: string;
  value: string;
};

type StorageSource = {
  id: number;
  name: string;
  type: "Object Storage" | "NAS";
  directories: SourceDirectory[];
};

type TapeOption = {
  id: string;
  label: string;
  free: string;
  status: "Writable" | "Mounted";
};

type TapeDeviceOption = {
  id: string;
  name: string;
  location: string;
  tapes: TapeOption[];
};

type BackupTask = {
  id: number;
  name: string;
  sourceStorageId: number;
  sourceStorageName: string;
  sourcePath: string;
  tapeDeviceId: string;
  tapeDeviceName: string;
  tapeId: string;
  tapeLabel: string;
  progress: number;
  transferred: string;
  totalSize: string;
  status: "Running" | "Paused" | "Completed";
  startedAt: string;
  updatedAt: string;
};

const storageOptions = ref<StorageSource[]>([
  {
    id: 1,
    name: "Cold Object Vault",
    type: "Object Storage",
    directories: [
      { label: "visionvault-cold / 2026 / camera-a", value: "/visionvault-cold/2026/camera-a" },
      { label: "visionvault-cold / 2026 / camera-b", value: "/visionvault-cold/2026/camera-b" },
      { label: "visionvault-cold / monthly-archive", value: "/visionvault-cold/monthly-archive" }
    ]
  },
  {
    id: 2,
    name: "Factory Archive",
    type: "NAS",
    directories: [
      { label: "/srv/archive/factory-a/day-shift", value: "/srv/archive/factory-a/day-shift" },
      { label: "/srv/archive/factory-a/night-shift", value: "/srv/archive/factory-a/night-shift" }
    ]
  }
]);

const tapeDevices = ref<TapeDeviceOption[]>([
  {
    id: "lib-01",
    name: "LTO Library A",
    location: "机房 A / 机柜 07",
    tapes: [
      { id: "tape-001", label: "LTO9-0001", free: "6.8 TB", status: "Mounted" },
      { id: "tape-002", label: "LTO9-0002", free: "15.2 TB", status: "Writable" }
    ]
  },
  {
    id: "lib-02",
    name: "LTO Library B",
    location: "机房 B / 机柜 02",
    tapes: [
      { id: "tape-013", label: "LTO8-0013", free: "3.9 TB", status: "Mounted" },
      { id: "tape-015", label: "LTO8-0015", free: "12.0 TB", status: "Writable" }
    ]
  }
]);

const tasks = ref<BackupTask[]>([
  {
    id: 101,
    name: "Camera-A May Incremental Backup",
    sourceStorageId: 1,
    sourceStorageName: "Cold Object Vault",
    sourcePath: "/visionvault-cold/2026/camera-a",
    tapeDeviceId: "lib-01",
    tapeDeviceName: "LTO Library A",
    tapeId: "tape-001",
    tapeLabel: "LTO9-0001",
    progress: 68,
    transferred: "8.6 TB",
    totalSize: "12.7 TB",
    status: "Running",
    startedAt: "2026-05-26 09:15",
    updatedAt: "2026-05-26 10:28"
  },
  {
    id: 102,
    name: "Night Shift Weekly Archive",
    sourceStorageId: 2,
    sourceStorageName: "Factory Archive",
    sourcePath: "/srv/archive/factory-a/night-shift",
    tapeDeviceId: "lib-02",
    tapeDeviceName: "LTO Library B",
    tapeId: "tape-013",
    tapeLabel: "LTO8-0013",
    progress: 31,
    transferred: "2.5 TB",
    totalSize: "8.0 TB",
    status: "Paused",
    startedAt: "2026-05-25 23:40",
    updatedAt: "2026-05-26 07:50"
  },
  {
    id: 103,
    name: "Monthly Cold Archive to Tape",
    sourceStorageId: 1,
    sourceStorageName: "Cold Object Vault",
    sourcePath: "/visionvault-cold/monthly-archive",
    tapeDeviceId: "lib-01",
    tapeDeviceName: "LTO Library A",
    tapeId: "tape-002",
    tapeLabel: "LTO9-0002",
    progress: 100,
    transferred: "15.2 TB",
    totalSize: "15.2 TB",
    status: "Completed",
    startedAt: "2026-05-24 18:00",
    updatedAt: "2026-05-24 23:25"
  }
]);

const modalVisible = ref(false);
const creating = ref(false);

const form = reactive({
  name: "",
  sourceStorageId: 0,
  sourcePath: "",
  tapeDeviceId: "",
  tapeId: ""
});

const runningTasks = computed(() => tasks.value.filter((item) => item.status === "Running"));
const pausedTasks = computed(() => tasks.value.filter((item) => item.status === "Paused"));
const finishedTasks = computed(() => tasks.value.filter((item) => item.status === "Completed"));
const selectedStorage = computed(() => storageOptions.value.find((item) => item.id === form.sourceStorageId) ?? null);
const selectedTapeDevice = computed(() => tapeDevices.value.find((item) => item.id === form.tapeDeviceId) ?? null);

watch(
  () => form.sourceStorageId,
  () => {
    form.sourcePath = "";
  }
);

watch(
  () => form.tapeDeviceId,
  () => {
    form.tapeId = "";
  }
);

function openCreateModal() {
  form.name = "";
  form.sourceStorageId = storageOptions.value[0]?.id ?? 0;
  form.sourcePath = "";
  form.tapeDeviceId = tapeDevices.value[0]?.id ?? "";
  form.tapeId = "";
  modalVisible.value = true;
}

function closeModal() {
  modalVisible.value = false;
}

function pauseTask(task: BackupTask) {
  if (task.status !== "Running") {
    return;
  }
  task.status = "Paused";
  task.updatedAt = "Just now";
  Message.success(`Task paused: ${task.name}`);
}

function resumeTask(task: BackupTask) {
  if (task.status !== "Paused") {
    return;
  }
  task.status = "Running";
  task.updatedAt = "Just now";
  Message.success(`Task resumed: ${task.name}`);
}

function deleteTask(task: BackupTask) {
  tasks.value = tasks.value.filter((item) => item.id !== task.id);
  Message.success(`Task deleted: ${task.name}`);
}

function createTask() {
  if (!form.name.trim() || !form.sourceStorageId || !form.sourcePath || !form.tapeDeviceId || !form.tapeId) {
    Message.error("Select a backup source, directory, tape library, and tape");
    return;
  }

  const storage = storageOptions.value.find((item) => item.id === form.sourceStorageId);
  const device = tapeDevices.value.find((item) => item.id === form.tapeDeviceId);
  const tape = device?.tapes.find((item) => item.id === form.tapeId);
  if (!storage || !device || !tape) {
    Message.error("The selected backup resources are invalid");
    return;
  }

  creating.value = true;
  tasks.value.unshift({
    id: Date.now(),
    name: form.name.trim(),
    sourceStorageId: storage.id,
    sourceStorageName: storage.name,
    sourcePath: form.sourcePath,
    tapeDeviceId: device.id,
    tapeDeviceName: device.name,
    tapeId: tape.id,
    tapeLabel: tape.label,
    progress: 0,
    transferred: "0 GB",
    totalSize: "Pending scan",
    status: "Running",
    startedAt: "Created just now",
    updatedAt: "Just now"
  });
  creating.value = false;
  modalVisible.value = false;
  Message.success("Backup task created");
}
</script>

<template>
  <div class="page-grid">
    <PageHeader title="Backup Task Management" description="Manage object storage to tape backup jobs, including progress tracking, stop, delete, resume, and new task creation from a modal.">
      <a-space>
        <a-button>Refresh Tasks</a-button>
        <a-button type="primary" @click="openCreateModal">New Backup</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Running Tasks" :value="runningTasks.length" hint="Currently writing to tape" :icon="IconPlayCircle" />
      <StatCard label="Paused Tasks" :value="pausedTasks.length" hint="Waiting to resume" :icon="IconPauseCircle" />
      <StatCard label="Total Tasks" :value="tasks.length" hint="Includes completed archives" :icon="IconArchive" />
      <StatCard label="Backup Sources" :value="storageOptions.length" hint="Available storages and directories" :icon="IconStorage" />
    </div>

    <a-card class="panel-card" title="Running Backups">
      <div v-if="runningTasks.length" class="running-grid">
        <div v-for="task in runningTasks" :key="task.id" class="running-task">
          <div class="running-task__top">
            <div>
              <div class="running-task__title">{{ task.name }}</div>
              <div class="running-task__meta">{{ task.sourceStorageName }} · {{ task.sourcePath }}</div>
            </div>
            <a-tag color="arcoblue">{{ task.tapeDeviceName }} / {{ task.tapeLabel }}</a-tag>
          </div>
          <a-progress :percent="task.progress" :show-text="true" />
          <div class="running-task__foot">
            <span>{{ task.transferred }} / {{ task.totalSize }}</span>
            <a-space size="mini">
              <a-button size="mini" @click="pauseTask(task)">Stop Task</a-button>
              <a-button size="mini" status="danger" @click="deleteTask(task)">Delete Task</a-button>
            </a-space>
          </div>
        </div>
      </div>
      <a-empty v-else description="There are no running backup tasks right now" />
    </a-card>

    <a-card class="panel-card" title="All Backup Records">
      <a-table :data="tasks" :pagination="false">
        <template #columns>
          <a-table-column title="Task Name" data-index="name" />
          <a-table-column title="Backup Source">
            <template #cell="{ record }">
              <div>{{ record.sourceStorageName }}</div>
              <div class="cell-subtext">{{ record.sourcePath }}</div>
            </template>
          </a-table-column>
          <a-table-column title="Target Tape">
            <template #cell="{ record }">
              <div>{{ record.tapeDeviceName }}</div>
              <div class="cell-subtext">{{ record.tapeLabel }}</div>
            </template>
          </a-table-column>
          <a-table-column title="Progress">
            <template #cell="{ record }">
              <div class="progress-cell">
                <a-progress :percent="record.progress" :stroke-width="14" :show-text="false" />
                <span>{{ record.progress }}%</span>
              </div>
            </template>
          </a-table-column>
          <a-table-column title="Status">
            <template #cell="{ record }">
              <a-tag :color="record.status === 'Running' ? 'arcoblue' : record.status === 'Paused' ? 'orange' : 'green'">
                {{ record.status }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Updated At" data-index="updatedAt" />
          <a-table-column title="Actions" :width="220">
            <template #cell="{ record }">
              <a-space size="mini">
                <a-button v-if="record.status === 'Running'" size="mini" @click="pauseTask(record)">Stop Task</a-button>
                <a-button v-if="record.status === 'Paused'" size="mini" type="outline" @click="resumeTask(record)">Resume Backup</a-button>
                <a-button size="mini" status="danger" @click="deleteTask(record)">Delete Task</a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-modal :visible="modalVisible" title="New Backup" :confirm-loading="creating" unmount-on-close @ok="createTask" @cancel="closeModal">
      <a-form :model="form" layout="vertical" class="modal-bordered-form">
        <a-form-item field="name" label="Task Name" required>
          <a-input v-model="form.name" placeholder="Example: Camera-A June Full Backup" />
        </a-form-item>
        <a-form-item field="sourceStorageId" label="Source Storage" required>
          <a-select v-model="form.sourceStorageId" placeholder="Select a backup source">
            <a-option v-for="item in storageOptions" :key="item.id" :value="item.id">
              {{ item.name }} ({{ item.type }})
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="sourcePath" label="Directory in Storage" required>
          <a-select v-model="form.sourcePath" placeholder="Select a directory" :disabled="!selectedStorage">
            <a-option v-for="dir in selectedStorage?.directories ?? []" :key="dir.value" :value="dir.value">
              {{ dir.label }}
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="tapeDeviceId" label="Tape Library" required>
          <a-select v-model="form.tapeDeviceId" placeholder="Select a tape library">
            <a-option v-for="item in tapeDevices" :key="item.id" :value="item.id">
              {{ item.name }} ({{ item.location }})
            </a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="tapeId" label="Tape" required>
          <a-select v-model="form.tapeId" placeholder="Select a tape" :disabled="!selectedTapeDevice">
            <a-option v-for="tape in selectedTapeDevice?.tapes ?? []" :key="tape.id" :value="tape.id">
              {{ tape.label }} · Free {{ tape.free }} · {{ tape.status }}
            </a-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.running-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.running-task {
  padding: 18px;
  border-radius: 14px;
  background: rgba(240, 244, 247, 0.88);
  border: 1px solid var(--border-soft);
}

.running-task__top,
.running-task__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.running-task__title {
  font-size: 15px;
  font-weight: 700;
}

.running-task__meta,
.cell-subtext {
  color: var(--text-2);
  font-size: 12px;
  margin-top: 4px;
}

.running-task__foot {
  margin-top: 12px;
}

.progress-cell {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 10px;
  align-items: center;
}

@media (max-width: 1200px) {
  .running-grid {
    grid-template-columns: 1fr;
  }
}
</style>
