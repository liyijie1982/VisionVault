<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconCalendarClock, IconCheckCircle, IconCloseCircle, IconHistory } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { fetchLoginLogs } from "../api/skybase";
import type { LoginLogRecord, PagedResult } from "../types";

const loading = ref(false);
const logsPage = ref<PagedResult<LoginLogRecord>>({
  items: [],
  total: 0,
  page: 1,
  pageSize: 20
});

const filters = reactive({
  username: "",
  startAt: "",
  endAt: ""
});

const pagination = reactive({
  page: 1,
  pageSize: 20
});

const successCount = computed(() => logsPage.value.items.filter((item) => item.loginStatus === 1).length);
const failedCount = computed(() => logsPage.value.items.filter((item) => item.loginStatus !== 1).length);
const latestEventAt = computed(() => logsPage.value.items[0]?.createdAt ?? "No records");

function userAgentOS(userAgent: string) {
  const source = userAgent.toLowerCase();
  if (source.includes("windows")) {
    return "Windows";
  }
  if (source.includes("mac os x") || source.includes("macintosh")) {
    return "macOS";
  }
  if (source.includes("android")) {
    return "Android";
  }
  if (source.includes("iphone") || source.includes("ipad") || source.includes("ios")) {
    return "iOS";
  }
  if (source.includes("linux")) {
    return "Linux";
  }
  return "Unknown";
}

async function loadLoginLogs(resetPage = false) {
  if (resetPage) {
    pagination.page = 1;
  }

  loading.value = true;
  try {
    logsPage.value = await fetchLoginLogs({
      username: filters.username,
      startAt: filters.startAt || undefined,
      endAt: filters.endAt || undefined,
      page: pagination.page,
      pageSize: pagination.pageSize
    });
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load login logs");
  } finally {
    loading.value = false;
  }
}

function resetFilters() {
  filters.username = "";
  filters.startAt = "";
  filters.endAt = "";
  loadLoginLogs(true);
}

function handlePageChange(page: number) {
  pagination.page = page;
  loadLoginLogs();
}

function handlePageSizeChange(pageSize: number) {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  loadLoginLogs();
}

onMounted(() => {
  loadLoginLogs(true);
});
</script>

<template>
  <div class="page-grid">
    <PageHeader title="" description="">
      <a-space>
        <a-button @click="loadLoginLogs()">Refresh</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Total Records" :value="logsPage.total" hint="Matched by current query" :icon="IconHistory" />
      <StatCard label="Success On Page" :value="successCount" hint="Login attempts accepted" :icon="IconCheckCircle" />
      <StatCard label="Failed On Page" :value="failedCount" hint="Rejected authentication attempts" :icon="IconCloseCircle" />
      <StatCard label="Latest Event" :value="latestEventAt" hint="Newest record in current result set" :icon="IconCalendarClock" />
    </div>

    <a-card class="panel-card">
      <div class="toolbar-row login-log-toolbar">
        <a-input v-model="filters.username" class="login-log-toolbar__field login-log-toolbar__field--username" placeholder="Username" allow-clear />
        <a-date-picker
          v-model="filters.startAt"
          class="login-log-toolbar__field login-log-toolbar__field--date"
          placeholder="Start date"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          allow-clear
        />
        <a-date-picker
          v-model="filters.endAt"
          class="login-log-toolbar__field login-log-toolbar__field--date"
          placeholder="End date"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          allow-clear
        />
        <a-space class="login-log-toolbar__actions">
          <a-button type="primary" @click="loadLoginLogs(true)">Search</a-button>
          <a-button @click="resetFilters">Reset</a-button>
        </a-space>
      </div>
    </a-card>

    <a-card class="panel-card" title="Login Log Records">
      <a-table
        :data="logsPage.items"
        :loading="loading"
        row-key="id"
        :pagination="{
          total: logsPage.total,
          current: pagination.page,
          pageSize: pagination.pageSize,
          showTotal: true,
          showPageSize: true,
          pageSizeOptions: [10, 20, 50, 100]
        }"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #columns>
          <a-table-column title="Username" data-index="username" :width="160" />
          <a-table-column title="Login IP" data-index="loginIp" :width="160" />
          <a-table-column title="Status" :width="120">
            <template #cell="{ record }">
              <a-tag :color="record.loginStatus === 1 ? 'green' : 'red'">
                {{ record.loginStatus === 1 ? "Success" : "Failed" }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Message" data-index="message" :width="220" />
          <a-table-column title="User Agent">
            <template #cell="{ record }">
              <span class="login-log-user-agent" :title="record.userAgent">{{ userAgentOS(record.userAgent) }}</span>
            </template>
          </a-table-column>
          <a-table-column title="Created At" data-index="createdAt" :width="180" />
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<style scoped>
.login-log-toolbar {
  flex-wrap: nowrap;
  align-items: center;
  justify-content: flex-start;
  gap: 12px;
}

.login-log-toolbar :deep(.arco-input-wrapper),
.login-log-toolbar :deep(.arco-select-view),
.login-log-toolbar :deep(.arco-picker) {
  width: 100%;
}

.login-log-toolbar__field {
  flex: 0 0 auto;
}

.login-log-toolbar__field--username {
  flex: 0 0 20%;
  width: 20%;
  min-width: 20%;
  max-width: 20%;
}

.login-log-toolbar__field--date {
  flex: 0 0 14%;
  width: 14%;
  min-width: 14%;
  max-width: 14%;
}

.login-log-toolbar__field--username :deep(.arco-input-wrapper) {
  width: 100%;
  min-width: 100%;
  max-width: 100%;
}

.login-log-toolbar__actions {
  flex: 0 0 auto;
  margin-left: auto;
}

@media (max-width: 1440px) {
  .login-log-toolbar {
    gap: 10px;
  }

  .login-log-toolbar__field--username {
    flex-basis: 20%;
    width: 20%;
    min-width: 20%;
    max-width: 20%;
  }

  .login-log-toolbar__field--date {
    width: 14%;
  }
}

.login-log-user-agent {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  color: var(--vv-text-secondary);
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
