<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { IconCalendarClock, IconDashboard, IconSettings, IconStorage } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { fetchHealth, fetchModules } from "../api/skybase";
import type { HealthStatus, ModuleSummary } from "../types";

const health = ref<HealthStatus | null>(null);
const modules = ref<ModuleSummary[]>([]);
const router = useRouter();

async function loadSystem() {
  [health.value, modules.value] = await Promise.all([fetchHealth(), fetchModules()]);
}

onMounted(() => {
  loadSystem();
});
</script>

<template>
  <div class="page-grid">
    <PageHeader
      title="System"
      description="Service runtime metadata, module surface, and quick access into platform configuration."
    >
      <a-space>
        <a-button @click="loadSystem">Refresh Health</a-button>
        <a-button type="primary" @click="router.push('/system/parameters')">Open Settings</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Service Name" :value="health?.name ?? 'SkyBase'" hint="Backend service identity" :icon="IconStorage" />
      <StatCard label="Environment" :value="health?.env ?? 'development'" hint="Current runtime environment" :icon="IconSettings" />
      <StatCard label="Status" :value="health?.status ?? 'ok'" hint="Health envelope state" :icon="IconDashboard" />
      <StatCard label="Started At" :value="health?.startedAt ?? 'Unavailable'" hint="Service start timestamp" :icon="IconCalendarClock" />
    </div>

    <div class="content-grid content-grid--equal">
      <a-card class="panel-card" title="Runtime Metadata">
        <div class="kv-list">
          <div class="kv-list__row"><span>Service</span><span>{{ health?.name ?? "SkyBase" }}</span></div>
          <div class="kv-list__row"><span>Environment</span><span>{{ health?.env ?? "development" }}</span></div>
          <div class="kv-list__row"><span>Status</span><span>{{ health?.status ?? "ok" }}</span></div>
          <div class="kv-list__row"><span>Current Time</span><span>{{ health?.now ?? "Unavailable" }}</span></div>
        </div>
      </a-card>

      <a-card class="panel-card" title="Module Surface">
        <div class="stack-list">
          <div v-for="module in modules" :key="module.key" class="stack-list__item">
            <div class="stack-list__title">{{ module.name }}</div>
            <div class="stack-list__text">{{ module.description }}</div>
          </div>
        </div>
      </a-card>
    </div>
  </div>
</template>
