<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { IconCheckCircle, IconCloseCircle, IconFire, IconStorage } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import MetricBar from "../components/MetricBar.vue";
import { fetchAgents } from "../api/skybase";
import type { AgentRecord } from "../types";

const agents = ref<AgentRecord[]>([]);
const router = useRouter();

const online = computed(() => agents.value.filter((item) => item.status === 1));
const offline = computed(() => agents.value.filter((item) => item.status !== 1));
const avgCpu = computed(() =>
  online.value.length ? Math.round(online.value.reduce((sum, item) => sum + item.cpu, 0) / online.value.length) : 0
);
const avgMem = computed(() =>
  online.value.length ? Math.round(online.value.reduce((sum, item) => sum + item.mem, 0) / online.value.length) : 0
);
const diskPressure = computed(() => {
  const usage = online.value.flatMap((item) =>
    item.storage.map((storage) => (storage.total > 0 ? Math.round((storage.used / storage.total) * 100) : 0))
  );
  return usage.length ? Math.round(usage.reduce((sum, item) => sum + item, 0) / usage.length) : 0;
});
const versionDistribution = computed(() => {
  const counts = new Map<string, number>();
  for (const item of agents.value) {
    counts.set(item.version, (counts.get(item.version) ?? 0) + 1);
  }
  return Array.from(counts.entries()).map(([version, count]) => ({ version, count }));
});

async function loadSnapshot() {
  agents.value = await fetchAgents();
}

onMounted(() => {
  loadSnapshot();
});
</script>

<template>
  <div class="page-grid">
    <PageHeader title="" description="">
      <a-space>
        <a-button @click="loadSnapshot">Refresh Snapshot</a-button>
        <a-button type="primary" @click="router.push('/alerts/policies')">Open Alert Rules</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Online Agents" :value="online.length" hint="Heartbeat currently available" :icon="IconCheckCircle" />
      <StatCard label="Offline Agents" :value="offline.length" hint="Potential loss of visibility" :icon="IconCloseCircle" />
      <StatCard label="Average CPU" :value="`${avgCpu}%`" hint="Across online agents" :icon="IconFire" />
      <StatCard label="Average Memory" :value="`${avgMem}%`" hint="Across online agents" :icon="IconStorage" />
    </div>

    <div class="content-grid content-grid--equal">
      <a-card class="panel-card" title="Resource Pressure">
        <MetricBar label="CPU Trend Window" :value="avgCpu" tone="warning" />
        <MetricBar label="Memory Trend Window" :value="avgMem" tone="default" />
        <MetricBar label="Disk Pressure" :value="diskPressure" tone="danger" />
      </a-card>

      <a-card class="panel-card" title="Version Distribution">
        <div class="stack-list">
          <div v-for="item in versionDistribution" :key="item.version" class="stack-list__item">
            <div class="stack-list__title">{{ item.version }}</div>
            <div class="stack-list__text">{{ item.count }} agents</div>
          </div>
        </div>
      </a-card>
    </div>

    <a-card class="panel-card" title="Latest Heartbeats">
      <a-table :data="agents" :pagination="false">
        <template #columns>
          <a-table-column title="Host" data-index="hostName" />
          <a-table-column title="IP" data-index="ip" />
          <a-table-column title="Last Heartbeat" data-index="lastAccessTime" />
          <a-table-column title="CPU">
            <template #cell="{ record }">
              {{ record.cpu }}%
            </template>
          </a-table-column>
          <a-table-column title="Memory">
            <template #cell="{ record }">
              {{ record.mem }}%
            </template>
          </a-table-column>
          <a-table-column title="Status">
            <template #cell="{ record }">
              <a-tag :color="record.status === 1 ? 'green' : 'gray'">{{ record.status === 1 ? "Online" : "Offline" }}</a-tag>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>
  </div>
</template>
