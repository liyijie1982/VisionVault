<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconCloud, IconDelete, IconEdit, IconPauseCircle, IconStorage } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { createStorageTarget, deleteStorageTarget, fetchStorageTargets, updateStorageTarget } from "../api/skybase";
import type { StoragePayload, StorageRecord } from "../types";

const targets = ref<StorageRecord[]>([]);
const loading = ref(false);
const submitting = ref(false);
const modalVisible = ref(false);
const editingId = ref<number | null>(null);

const form = reactive<StoragePayload>({
  name: "",
  type: "local",
  endpoint: "",
  accessKey: "",
  secretKey: "",
  bucket: "",
  region: "",
  localPath: "",
  quota: 0,
  status: 1,
  remark: ""
});

const localCount = computed(() => targets.value.filter((item) => item.type === "local").length);
const s3Count = computed(() => targets.value.filter((item) => item.type === "s3").length);
const disabledCount = computed(() => targets.value.filter((item) => item.status !== 1).length);
const modalTitle = computed(() => (editingId.value ? "Edit Storage" : "Add Storage"));

function resetForm() {
  editingId.value = null;
  form.name = "";
  form.type = "local";
  form.endpoint = "";
  form.accessKey = "";
  form.secretKey = "";
  form.bucket = "";
  form.region = "";
  form.localPath = "";
  form.quota = 0;
  form.status = 1;
  form.remark = "";
}

function openCreateModal() {
  resetForm();
  modalVisible.value = true;
}

function openEditModal(record: StorageRecord) {
  editingId.value = record.id;
  form.name = record.name;
  form.type = record.type;
  form.endpoint = record.endpoint;
  form.accessKey = record.accessKey;
  form.secretKey = record.secretKey;
  form.bucket = record.bucket;
  form.region = record.region;
  form.localPath = record.localPath;
  form.quota = record.quota;
  form.status = record.status;
  form.remark = record.remark;
  modalVisible.value = true;
}

function closeModal() {
  modalVisible.value = false;
  resetForm();
}

async function loadData() {
  loading.value = true;
  try {
    targets.value = await fetchStorageTargets();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load storage");
  } finally {
    loading.value = false;
  }
}

async function submitStorage() {
  if (!form.name.trim()) {
    Message.error("Storage name is required");
    return;
  }
  if (form.type === "local" && !form.localPath.trim()) {
    Message.error("Local path is required");
    return;
  }
  if (form.type === "s3" && (!form.endpoint.trim() || !form.bucket.trim())) {
    Message.error("S3 endpoint and bucket are required");
    return;
  }

  submitting.value = true;
  try {
    const payload: StoragePayload = {
      ...form,
      name: form.name.trim(),
      endpoint: form.endpoint.trim(),
      accessKey: form.accessKey.trim(),
      secretKey: form.secretKey.trim(),
      bucket: form.bucket.trim(),
      region: form.region.trim(),
      localPath: form.localPath.trim(),
      remark: form.remark.trim()
    };

    if (editingId.value) {
      await updateStorageTarget(editingId.value, payload);
      Message.success("Storage updated");
    } else {
      await createStorageTarget(payload);
      Message.success("Storage created");
    }
    closeModal();
    await loadData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to save storage");
  } finally {
    submitting.value = false;
  }
}

async function removeStorage(record: StorageRecord) {
  try {
    await deleteStorageTarget(record.id);
    Message.success(`Deleted storage ${record.name}`);
    await loadData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to delete storage");
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
        <a-button type="primary" @click="openCreateModal">Add Target</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Total Targets" :value="targets.length" hint="Registered storage definitions" :icon="IconStorage" />
      <StatCard label="Local Targets" :value="localCount" hint="Filesystem storage endpoints" :icon="IconStorage" />
      <StatCard label="S3 Targets" :value="s3Count" hint="Object storage endpoints" :icon="IconCloud" />
      <StatCard label="Disabled" :value="disabledCount" hint="Temporarily unavailable targets" :icon="IconPauseCircle" />
    </div>

    <div class="storage-card-grid">
      <a-card v-for="item in targets" :key="item.id" class="panel-card storage-card">
        <div class="storage-card__head">
          <div>
            <div class="storage-card__title">{{ item.name }}</div>
            <div class="storage-card__subtitle">{{ item.type === "local" ? "Local storage" : "S3-compatible object storage" }}</div>
          </div>
          <a-space size="mini">
            <a-tag :color="item.status === 1 ? 'green' : 'gray'">{{ item.status === 1 ? "Active" : "Disabled" }}</a-tag>
            <a-button class="action-icon-button" type="text" @click="openEditModal(item)">
              <template #icon><IconEdit /></template>
            </a-button>
            <a-popconfirm content="Delete this storage target?" @ok="removeStorage(item)">
              <a-button class="action-icon-button action-icon-button--danger" type="text" status="danger">
                <template #icon><IconDelete /></template>
              </a-button>
            </a-popconfirm>
          </a-space>
        </div>
        <div class="kv-list">
          <div class="kv-list__row"><span>Created At</span><span>{{ item.createdAt }}</span></div>
          <div class="kv-list__row"><span>Quota</span><span>{{ (item.quota / 1_000_000_000_000).toFixed(1) }} TB</span></div>
          <div class="kv-list__row"><span>Region</span><span>{{ item.region || "Local host" }}</span></div>
          <div class="kv-list__row"><span>Endpoint</span><span>{{ item.endpoint || item.localPath }}</span></div>
          <div class="kv-list__row"><span>Reference</span><span>{{ item.bucket || item.localPath }}</span></div>
        </div>
        <p class="panel-note">{{ item.remark || "No remark" }}</p>
      </a-card>
    </div>

    <a-modal :visible="modalVisible" :title="modalTitle" :confirm-loading="submitting" unmount-on-close @ok="submitStorage" @cancel="closeModal">
      <a-form :model="form" layout="vertical" class="modal-bordered-form">
        <a-form-item field="name" label="Storage Name" required>
          <a-input v-model="form.name" placeholder="Enter storage name" />
        </a-form-item>
        <a-form-item field="type" label="Type">
          <a-radio-group v-model="form.type" type="button">
            <a-radio value="local">Local</a-radio>
            <a-radio value="s3">S3</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item v-if="form.type === 'local'" field="localPath" label="Local Path" required>
          <a-input v-model="form.localPath" placeholder="Enter local path" />
        </a-form-item>
        <template v-else>
          <a-form-item field="endpoint" label="Endpoint" required>
            <a-input v-model="form.endpoint" placeholder="Enter S3 endpoint" />
          </a-form-item>
          <a-form-item field="bucket" label="Bucket" required>
            <a-input v-model="form.bucket" placeholder="Enter bucket name" />
          </a-form-item>
          <a-form-item field="region" label="Region">
            <a-input v-model="form.region" placeholder="Enter region" />
          </a-form-item>
          <a-form-item field="accessKey" label="Access Key">
            <a-input v-model="form.accessKey" placeholder="Enter access key" />
          </a-form-item>
          <a-form-item field="secretKey" label="Secret Key">
            <a-input-password v-model="form.secretKey" placeholder="Enter secret key" />
          </a-form-item>
        </template>
        <a-form-item field="quota" label="Quota (bytes)">
          <a-input-number v-model="form.quota" :min="0" style="width: 100%" />
        </a-form-item>
        <a-form-item field="status" label="Status">
          <a-radio-group v-model="form.status" type="button">
            <a-radio :value="1">Active</a-radio>
            <a-radio :value="0">Disabled</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item field="remark" label="Remark">
          <a-textarea v-model="form.remark" :max-length="300" placeholder="Enter remark" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.storage-card__head {
  margin-bottom: 18px;
}
</style>
