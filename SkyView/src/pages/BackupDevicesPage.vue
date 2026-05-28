<script setup lang="ts">
import { computed, ref } from "vue";
import { IconArchive, IconCheckCircle, IconStorage, IconTool } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";

type TapeItem = {
  id: string;
  label: string;
  slot: string;
  pool: string;
  capacity: string;
  used: string;
  status: "Mounted" | "Idle" | "Write Protected";
  lastWriteAt: string;
};

type TapeDevice = {
  id: string;
  name: string;
  vendor: string;
  host: string;
  path: string;
  status: "Online" | "Alert";
  mountedAt: string;
  robot: string;
  driveCount: number;
  tapes: TapeItem[];
};

const devices = ref<TapeDevice[]>([
  {
    id: "lib-01",
    name: "LTO Library A",
    vendor: "HPE MSL2024",
    host: "backup-node-01",
    path: "/dev/sg3",
    status: "Online",
    mountedAt: "2026-05-26 08:15",
    robot: "24 slots / 2 drives",
    driveCount: 2,
    tapes: [
      { id: "tape-001", label: "LTO9-0001", slot: "Slot-01", pool: "Production Archive", capacity: "18 TB", used: "11.2 TB", status: "Mounted", lastWriteAt: "2026-05-26 10:24" },
      { id: "tape-002", label: "LTO9-0002", slot: "Slot-02", pool: "Production Archive", capacity: "18 TB", used: "2.8 TB", status: "Idle", lastWriteAt: "2026-05-24 19:05" }
    ]
  },
  {
    id: "lib-02",
    name: "LTO Library B",
    vendor: "IBM TS4300",
    host: "backup-node-02",
    path: "/dev/sg5",
    status: "Online",
    mountedAt: "2026-05-25 21:40",
    robot: "40 slots / 4 drives",
    driveCount: 4,
    tapes: [
      { id: "tape-013", label: "LTO8-0013", slot: "Slot-08", pool: "Monthly Cold Backup", capacity: "12 TB", used: "8.1 TB", status: "Mounted", lastWriteAt: "2026-05-25 23:18" },
      { id: "tape-014", label: "LTO8-0014", slot: "Slot-09", pool: "Monthly Cold Backup", capacity: "12 TB", used: "0 TB", status: "Write Protected", lastWriteAt: "2026-05-01 09:00" }
    ]
  }
]);

const mountedTapeCount = computed(() => devices.value.reduce((sum, item) => sum + item.tapes.filter((tape) => tape.status === "Mounted").length, 0));
const totalTapeCount = computed(() => devices.value.reduce((sum, item) => sum + item.tapes.length, 0));
const onlineDeviceCount = computed(() => devices.value.filter((item) => item.status === "Online").length);
</script>

<template>
  <div class="page-grid">
    <PageHeader title="Backup Tape Devices" description="Review mounted tape libraries, drives, and media details used as backup targets for object storage to tape workflows.">
      <a-space>
        <a-button>Refresh Topology</a-button>
        <a-button type="primary">Rescan Mounts</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Mounted Devices" :value="devices.length" hint="Tape libraries available for backup writes" :icon="IconStorage" />
      <StatCard label="Online Devices" :value="onlineDeviceCount" hint="Connectivity and drive health are normal" :icon="IconCheckCircle" />
      <StatCard label="Mounted Tapes" :value="mountedTapeCount" hint="Media currently mounted or ready" :icon="IconArchive" />
      <StatCard label="Total Tapes" :value="totalTapeCount" hint="Managed tape media inventory" :icon="IconTool" />
    </div>

    <div class="device-grid">
      <a-card v-for="device in devices" :key="device.id" class="panel-card">
        <template #title>
          <div class="device-card__title-row">
            <div>
              <div class="device-card__title">{{ device.name }}</div>
              <div class="device-card__subtitle">{{ device.vendor }} · {{ device.host }}</div>
            </div>
            <a-tag :color="device.status === 'Online' ? 'green' : 'red'">{{ device.status }}</a-tag>
          </div>
        </template>

        <div class="kv-list">
          <div class="kv-list__row"><span>Device Path</span><span>{{ device.path }}</span></div>
          <div class="kv-list__row"><span>Mounted At</span><span>{{ device.mountedAt }}</span></div>
          <div class="kv-list__row"><span>Robot Spec</span><span>{{ device.robot }}</span></div>
          <div class="kv-list__row"><span>Drive Count</span><span>{{ device.driveCount }}</span></div>
        </div>

        <a-table :data="device.tapes" :pagination="false" size="small" class="device-card__table">
          <template #columns>
            <a-table-column title="Tape" data-index="label" />
            <a-table-column title="Slot" data-index="slot" />
            <a-table-column title="Pool" data-index="pool" />
            <a-table-column title="Capacity">
              <template #cell="{ record }">{{ record.used }} / {{ record.capacity }}</template>
            </a-table-column>
            <a-table-column title="Status">
              <template #cell="{ record }">
                <a-tag :color="record.status === 'Mounted' ? 'arcoblue' : record.status === 'Write Protected' ? 'orange' : 'gray'">
                  {{ record.status }}
                </a-tag>
              </template>
            </a-table-column>
            <a-table-column title="Last Write" data-index="lastWriteAt" />
          </template>
        </a-table>
      </a-card>
    </div>
  </div>
</template>

<style scoped>
.device-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.device-card__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.device-card__title {
  font-size: 16px;
  font-weight: 700;
}

.device-card__subtitle {
  margin-top: 4px;
  color: var(--text-2);
  font-size: 13px;
}

.device-card__table {
  margin-top: 18px;
}

@media (max-width: 1200px) {
  .device-grid {
    grid-template-columns: 1fr;
  }
}
</style>
