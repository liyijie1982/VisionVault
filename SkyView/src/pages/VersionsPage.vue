<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconArchive, IconCode, IconFile, IconRefresh } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { fetchVersions, uploadVersion } from "../api/skybase";
import type { VersionAgentSummary, VersionListResult, VersionRecord } from "../types";

const versions = ref<VersionRecord[]>([]);
const agentVersions = ref<VersionAgentSummary[]>([]);
const totalAgents = ref(0);
const publishedPackages = ref(0);
const onlineAgents = ref(0);
const currentPackageAgents = ref(0);
const loading = ref(false);
const uploadVisible = ref(false);
const uploadSubmitting = ref(false);
const uploadResetKey = ref(0);
const form = reactive({
  version: "",
  releaseNotes: "",
  activate: true
});
const uploadFileValue = ref<File | null>(null);

const activeVersion = computed(() => versions.value.find((item) => item.status === "Active") ?? versions.value[0] ?? null);
const latestAgentVersion = computed(() => agentVersions.value[0] ?? null);
const latestPackageRecord = computed(() =>
  latestAgentVersion.value ? versions.value.find((item) => item.version === latestAgentVersion.value?.version) ?? null : null
);
const selectedFileSummary = computed(() => {
  if (!uploadFileValue.value) {
    return "No package selected";
  }
  const fileSize = uploadFileValue.value.size / 1024 / 1024;
  return `${uploadFileValue.value.name} (${fileSize >= 1 ? `${fileSize.toFixed(2)} MB` : `${Math.max(fileSize * 1024, 1).toFixed(0)} KB`})`;
});
const versionDistribution = computed(() => {
  const versionDetails = new Map(versions.value.map((item) => [item.version, item]));
  return agentVersions.value.map((item) => {
    const detail = versionDetails.get(item.version);
    return {
      id: detail?.id ?? `agent-version-${item.version}`,
      version: item.version,
      filename: detail?.filename ?? "No package record in sync_agent_version",
      agentCount: item.agentCount,
      isActive: item.isLatest
    };
  });
});

async function loadVersions() {
  loading.value = true;
  try {
    const result: VersionListResult = await fetchVersions();
    versions.value = result.items;
    agentVersions.value = result.agentVersions;
    totalAgents.value = result.totalAgents;
    publishedPackages.value = result.publishedPackages;
    onlineAgents.value = result.onlineAgents;
    currentPackageAgents.value = result.currentPackageAgents;
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load versions");
  } finally {
    loading.value = false;
  }
}

function beforeUpload(fileItem: File) {
  uploadFileValue.value = fileItem;
  return false;
}

function removeUploadFile() {
  uploadFileValue.value = null;
  return true;
}

function openUploadModal() {
  form.version = activeVersion.value?.version ?? "";
  form.releaseNotes = "";
  form.activate = true;
  uploadFileValue.value = null;
  uploadResetKey.value += 1;
  uploadVisible.value = true;
}

function validateUploadForm() {
  const version = form.version.trim();
  if (!version) {
    Message.error("Version is required");
    return false;
  }
  if (!/^\d+(?:\.\d+)*$/.test(version)) {
    Message.error("Version format must be numeric, for example 1.9.3 or 1.9.3.4");
    return false;
  }
  if (!uploadFileValue.value) {
    Message.error("Package file is required");
    return false;
  }
  if (uploadFileValue.value.size <= 0) {
    Message.error("Selected package file is empty");
    return false;
  }
  return true;
}

async function submitUpload() {
  if (!validateUploadForm()) {
    return;
  }
  uploadSubmitting.value = true;
  try {
    await uploadVersion(form.version.trim(), uploadFileValue.value as File, form.releaseNotes.trim(), form.activate);
    Message.success("Package uploaded");
    uploadVisible.value = false;
    await loadVersions();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to upload package");
  } finally {
    uploadSubmitting.value = false;
  }
}

onMounted(() => {
  loadVersions();
});
</script>

<template>
  <div class="page-grid">
    <PageHeader
      title=""
      description=""
    >
      <a-space>
        <a-button :loading="loading" @click="loadVersions">
          <template #icon>
            <IconRefresh />
          </template>
          Refresh
        </a-button>
        <a-button type="primary" @click="openUploadModal">Upload Package</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Agent Total" :value="totalAgents" :hint="latestAgentVersion?.version ? `Latest agent version: ${latestAgentVersion.version}` : 'No agent version reported'" :icon="IconCode" />
      <StatCard label="Published Packages" :value="publishedPackages" hint="Package row count from sync_agent_version" :icon="IconArchive" />
      <StatCard label="Active Package" :value="onlineAgents" hint="Online agents from sync_agent" :icon="IconFile" />
      <StatCard label="Current Package Agents" :value="currentPackageAgents" hint="Agents on the latest package version" :icon="IconRefresh" />
    </div>

    <div class="content-grid content-grid--1-2">
      <a-card class="panel-card" title="Active Package">
        <div v-if="latestAgentVersion" class="stack-list">
          <div class="stack-list__item">
            <div class="stack-list__title">{{ latestAgentVersion.version }}</div>
            <div class="stack-list__text">{{ latestPackageRecord?.filename ?? "No package record in sync_agent_version" }}</div>
          </div>
          <div class="kv-list">
            <div class="kv-list__row"><span>Latest Version</span><span>{{ latestAgentVersion.version }}</span></div>
            <div class="kv-list__row"><span>Package ID</span><span>{{ latestPackageRecord?.id ?? "-" }}</span></div>
            <div class="kv-list__row"><span>MD5</span><span>{{ latestPackageRecord?.md5 ?? "-" }}</span></div>
            <div class="kv-list__row"><span>Status</span><span>{{ latestPackageRecord?.status ?? "From sync_agent" }}</span></div>
            <div class="kv-list__row"><span>Agent Count</span><span>{{ latestAgentVersion.agentCount }}</span></div>
          </div>
          <p class="panel-note">{{ latestPackageRecord?.releaseNotes || "Current version summary is coming from sync_agent. No matching package note was found in sync_agent_version." }}</p>
          <div class="version-distribution">
            <div class="version-distribution__header">
              <span>Agent Count by Version</span>
              <span>{{ totalAgents }} total</span>
            </div>
            <div
              v-for="item in versionDistribution"
              :key="item.id"
              class="version-distribution__row"
              :class="{ 'version-distribution__row--active': item.isActive }"
            >
              <div>
                <div class="version-distribution__version">{{ item.version }}</div>
                <div class="version-distribution__meta">{{ item.filename }}</div>
              </div>
              <a-tag :color="item.isActive ? 'green' : 'arcoblue'">{{ item.agentCount }} Agents</a-tag>
            </div>
          </div>
        </div>
        <a-empty v-else description="No agent version data" />
      </a-card>

      <a-card class="panel-card" title="Package History">
        <a-table :data="versions" :pagination="false" :loading="loading">
          <template #columns>
            <a-table-column title="Version" data-index="version" />
            <a-table-column title="Filename" data-index="filename" />
            <a-table-column title="Agent Count" data-index="agentCount" />
            <a-table-column title="Status">
              <template #cell="{ record }">
                <a-tag :color="record.status === 'Active' ? 'green' : 'gray'">{{ record.status }}</a-tag>
              </template>
            </a-table-column>
            <a-table-column title="Updated At" data-index="updatedAt" />
          </template>
        </a-table>
      </a-card>
    </div>

    <a-modal
      :visible="uploadVisible"
      title="Upload Package"
      :confirm-loading="uploadSubmitting"
      :ok-button-props="{ disabled: uploadSubmitting }"
      width="640px"
      @ok="submitUpload"
      @cancel="uploadVisible = false"
    >
      <a-form :model="form" layout="vertical">
        <a-alert class="upload-tip" type="info">
          Upload a new agent package, then optionally mark it as active immediately.
        </a-alert>
        <a-form-item field="version" label="Version" required>
          <a-input v-model="form.version" class="upload-field" placeholder="1.9.3.4" allow-clear />
        </a-form-item>
        <a-form-item field="releaseNotes" label="Release Notes">
          <a-textarea
            v-model="form.releaseNotes"
            class="upload-field"
            placeholder="Describe package changes, rollout notes, or special instructions"
            :auto-size="{ minRows: 4, maxRows: 6 }"
          />
        </a-form-item>
        <a-form-item field="activate" label="Activation">
          <div class="upload-activation">
            <a-switch v-model="form.activate" />
            <span>{{ form.activate ? "Set this package as active after upload" : "Only save this package record" }}</span>
          </div>
        </a-form-item>
        <a-form-item field="file" label="Package File" required>
          <a-upload
            :key="uploadResetKey"
            :auto-upload="false"
            :show-file-list="true"
            :limit="1"
            @before-upload="beforeUpload"
            @remove="removeUploadFile"
          >
            <template #upload-button>
              <a-button>Select Package</a-button>
            </template>
          </a-upload>
          <div class="upload-file-summary">{{ selectedFileSummary }}</div>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
:deep(.upload-field.arco-input-wrapper),
:deep(.upload-field.arco-textarea-wrapper),
:deep(.upload-field .arco-input-wrapper),
:deep(.upload-field .arco-textarea-wrapper) {
  border: 1px solid var(--color-border-3) !important;
  box-shadow: none !important;
  background-color: var(--color-bg-2);
}

:deep(.upload-field.arco-input-wrapper:hover),
:deep(.upload-field.arco-textarea-wrapper:hover),
:deep(.upload-field .arco-input-wrapper:hover),
:deep(.upload-field .arco-textarea-wrapper:hover) {
  border-color: rgb(var(--primary-5)) !important;
}

:deep(.upload-field.arco-input-wrapper-focus),
:deep(.upload-field.arco-textarea-focus),
:deep(.upload-field .arco-input-wrapper-focus),
:deep(.upload-field .arco-textarea-focus) {
  border-color: rgb(var(--primary-6)) !important;
  box-shadow: 0 0 0 2px rgba(var(--primary-6), 0.12) !important;
}

.upload-tip {
  margin-bottom: 16px;
}

.upload-activation {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--color-text-2);
}

.upload-file-summary {
  margin-top: 8px;
  width: 100%;
  min-height: 36px;
  padding: 8px 12px;
  border: 1px solid var(--color-border-3);
  border-radius: 6px;
  background-color: var(--color-fill-1);
  color: var(--color-text-2);
  font-size: 12px;
  box-sizing: border-box;
}

.version-distribution {
  margin-top: 12px;
  border-top: 1px solid var(--color-border-2);
  padding-top: 12px;
}

.version-distribution__header,
.version-distribution__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.version-distribution__header {
  margin-bottom: 10px;
  color: var(--color-text-2);
  font-size: 12px;
}

.version-distribution__row {
  padding: 10px 12px;
  border-radius: 10px;
  background: var(--color-fill-1);
}

.version-distribution__row + .version-distribution__row {
  margin-top: 8px;
}

.version-distribution__row--active {
  background: rgb(var(--green-1));
}

.version-distribution__version {
  font-weight: 600;
}

.version-distribution__meta {
  margin-top: 4px;
  color: var(--color-text-3);
  font-size: 12px;
  word-break: break-all;
}
</style>
