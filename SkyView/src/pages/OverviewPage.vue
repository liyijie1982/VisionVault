<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { IconCode, IconCommon, IconDashboard, IconSettings } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { fetchCurrentVersion, fetchHealth, fetchModules } from "../api/skybase";
import type { HealthStatus, ModuleSummary, VersionRecord } from "../types";

const health = ref<HealthStatus | null>(null);
const modules = ref<ModuleSummary[]>([]);
const currentVersion = ref<VersionRecord | null>(null);

const readyModules = computed(() => modules.value.filter((item) => item.status !== "planned").length);

async function loadOverview() {
  [health.value, modules.value, currentVersion.value] = await Promise.all([
    fetchHealth(),
    fetchModules(),
    fetchCurrentVersion()
  ]);
}

function exportSummary() {
  const blob = new Blob(
    [
      JSON.stringify(
        {
          exportedAt: new Date().toISOString(),
          health: health.value,
          modules: modules.value,
          currentVersion: currentVersion.value
        },
        null,
        2
      )
    ],
    { type: "application/json" }
  );
  const url = URL.createObjectURL(blob);
  window.open(url, "_blank");
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

onMounted(() => {
  loadOverview();
});
</script>

<template>
  <div class="page-grid">
    <PageHeader
      title="Overview"
      description="Control-plane snapshot for backend health, module readiness, and current agent package delivery."
    >
      <a-button type="outline" @click="exportSummary">Export Summary</a-button>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="System Status" :value="health?.status?.toUpperCase() ?? 'LOADING'" :hint="health?.name ?? 'SkyBase'" :icon="IconDashboard" />
      <StatCard label="Environment" :value="health?.env ?? 'development'" hint="Active runtime profile" :icon="IconSettings" />
      <StatCard label="Enabled Modules" :value="readyModules" :hint="`${modules.length} modules defined`" :icon="IconCommon" />
      <StatCard
        label="Current Agent Version"
        :value="currentVersion?.version ?? '1.9.3.3'"
        :hint="currentVersion?.filename ?? 'Current package ready'"
        :icon="IconCode"
      />
    </div>

    <div class="content-grid content-grid--2-1">
      <a-card class="panel-card" title="Platform Relationship">
        <div class="relationship-grid">
          <div class="relationship-node">
            <div class="relationship-node__title">SkyView</div>
            <div class="relationship-node__text">Management UI for dashboards, inventory, and policy operations.</div>
          </div>
          <div class="relationship-arrow">HTTP / JSON</div>
          <div class="relationship-node">
            <div class="relationship-node__title">SkyBase</div>
            <div class="relationship-node__text">Control plane, orchestration server, and package delivery service.</div>
          </div>
          <div class="relationship-arrow">HTTP / JSON</div>
          <div class="relationship-node">
            <div class="relationship-node__title">SkyDrop</div>
            <div class="relationship-node__text">Distributed collection agent responsible for heartbeat, sync, and scan tasks.</div>
          </div>
        </div>
      </a-card>

      <a-card class="panel-card" title="Release Focus">
        <div class="stack-list">
          <div class="stack-list__item">
            <div class="stack-list__title">Calm operational UI</div>
            <div class="stack-list__text">Low-saturation colors and structured cards improve scanability for daily operations.</div>
          </div>
          <div class="stack-list__item">
            <div class="stack-list__title">Backend-aligned navigation</div>
            <div class="stack-list__text">Modules match the scope defined in SkyBase requirements and current endpoints.</div>
          </div>
          <div class="stack-list__item">
            <div class="stack-list__title">Phased delivery</div>
            <div class="stack-list__text">Core management modules are live, while backup workflows still remain in a demo-only delivery phase.</div>
          </div>
        </div>
      </a-card>
    </div>

    <div class="content-grid content-grid--2-1">
      <a-card class="panel-card" title="Module Readiness">
        <a-table :data="modules" :pagination="false" :bordered="false">
          <template #columns>
            <a-table-column title="Module" data-index="name" />
            <a-table-column title="Description" data-index="description" />
            <a-table-column title="Status">
              <template #cell="{ record }">
                <a-tag :color="record.status === 'planned' ? 'gray' : 'arcoblue'">{{ record.status }}</a-tag>
              </template>
            </a-table-column>
          </template>
        </a-table>
      </a-card>

      <a-card class="panel-card" title="Recent Signals">
        <div class="timeline-list">
          <div class="timeline-list__item">
            <div class="timeline-list__time">{{ health?.now ?? "Now" }}</div>
            <div class="timeline-list__title">Health endpoint reachable</div>
            <div class="timeline-list__text">SkyView can resolve the current platform health envelope successfully.</div>
          </div>
          <div class="timeline-list__item">
            <div class="timeline-list__time">{{ currentVersion?.updatedAt ?? "Package state" }}</div>
            <div class="timeline-list__title">Version package available</div>
            <div class="timeline-list__text">The latest package metadata is ready for rollout visibility and download actions.</div>
          </div>
          <div class="timeline-list__item">
            <div class="timeline-list__time">Design phase</div>
            <div class="timeline-list__title">Backup workflows still staged</div>
            <div class="timeline-list__text">Tape devices, backup tasks, and backup logs are available as UI prototypes but are not yet backed by SkyBase APIs.</div>
          </div>
        </div>
      </a-card>
    </div>
  </div>
</template>
