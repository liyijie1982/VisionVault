<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconDelete, IconDownload, IconEdit, IconFile, IconFolder } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import { buildFileDownloadUrl, deleteFile, fetchFiles, fetchStorageTargets, updateFile, uploadFile } from "../api/skybase";
import type { FilePayload, FileRecord, StorageRecord } from "../types";

type FileInventoryRow =
  | { kind: "folder"; id: string; name: string; path: string; type: string; size: string; tags: string[]; modifiedAt: string; storage: string }
  | (FileRecord & { kind: "file" });

const files = ref<FileRecord[]>([]);
const storageTargets = ref<StorageRecord[]>([]);
const loading = ref(false);
const submitting = ref(false);
const modalVisible = ref(false);
const editingId = ref<string | null>(null);
const search = ref("");
const selectedStorage = ref<string>("all");
const currentPath = ref("/");
const tagInput = ref("");
const uploadFileValue = ref<File | null>(null);

const form = reactive<FilePayload>({
  name: "",
  path: "",
  type: "",
  size: "",
  tags: [],
  modifiedAt: "",
  storageId: 0
});

const storageFilteredFiles = computed(() =>
  files.value.filter((item) => selectedStorage.value === "all" || String(item.storageId) === selectedStorage.value)
);

const inventoryRows = computed<FileInventoryRow[]>(() => {
  const folderMap = new Map<string, FileInventoryRow>();
  const currentSegments = getPathSegments(currentPath.value);
  const rows: FileInventoryRow[] = [];

  storageFilteredFiles.value.forEach((item) => {
    const itemSegments = getPathSegments(item.path);
    const isUnderCurrentPath = currentSegments.every((segment, index) => itemSegments[index] === segment);
    if (!isUnderCurrentPath || itemSegments.length <= currentSegments.length) {
      return;
    }

    const nextSegment = itemSegments[currentSegments.length];
    const isDirectFile = itemSegments.length === currentSegments.length + 1;
    if (isDirectFile) {
      rows.push({ ...item, kind: "file" });
      return;
    }

    const folderPath = buildPath([...currentSegments, nextSegment]);
    if (!folderMap.has(folderPath)) {
      folderMap.set(folderPath, {
        kind: "folder",
        id: `folder:${folderPath}`,
        name: nextSegment,
        path: folderPath,
        type: "Folder",
        size: "--",
        tags: [],
        modifiedAt: "",
        storage: ""
      });
    }
  });

  const keyword = search.value.trim().toLowerCase();
  return [...folderMap.values(), ...rows]
    .filter((item) => !keyword || item.name.toLowerCase().includes(keyword) || item.path.toLowerCase().includes(keyword))
    .sort((a, b) => {
      if (a.kind !== b.kind) {
        return a.kind === "folder" ? -1 : 1;
      }
      return a.name.localeCompare(b.name);
    });
});

const pathBreadcrumbs = computed(() => {
  const segments = getPathSegments(currentPath.value);
  return segments.map((segment, index) => ({
    name: segment,
    path: buildPath(segments.slice(0, index + 1))
  }));
});

const modalTitle = computed(() => (editingId.value ? "Edit File" : "Upload File"));

function resetForm() {
  editingId.value = null;
  form.name = "";
  form.path = "";
  form.type = "";
  form.size = "";
  form.tags = [];
  form.modifiedAt = "";
  form.storageId = storageTargets.value[0]?.id ?? 0;
  tagInput.value = "";
  uploadFileValue.value = null;
}

function openCreateModal() {
  resetForm();
  modalVisible.value = true;
}

function openEditModal(record: FileRecord) {
  editingId.value = record.id;
  form.name = record.name;
  form.path = record.path;
  form.type = record.type;
  form.size = record.size;
  form.tags = [...record.tags];
  form.modifiedAt = record.modifiedAt;
  form.storageId = record.storageId;
  tagInput.value = record.tags.join(", ");
  modalVisible.value = true;
}

function closeModal() {
  modalVisible.value = false;
  resetForm();
}

async function loadData() {
  loading.value = true;
  try {
    [files.value, storageTargets.value] = await Promise.all([fetchFiles(), fetchStorageTargets()]);
    if (!form.storageId) {
      form.storageId = storageTargets.value[0]?.id ?? 0;
    }
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load files");
  } finally {
    loading.value = false;
  }
}

async function submitFile() {
  if (!form.storageId) {
    Message.error("Storage is required");
    return;
  }

  submitting.value = true;
  try {
    if (editingId.value) {
      if (!form.name.trim() || !form.path.trim()) {
        Message.error("File name and path are required");
        return;
      }
      const payload: FilePayload = {
        ...form,
        name: form.name.trim(),
        path: form.path.trim(),
        type: form.type.trim(),
        size: form.size.trim(),
        modifiedAt: form.modifiedAt.trim(),
        tags: tagInput.value.split(",").map((item) => item.trim()).filter(Boolean)
      };
      await updateFile(editingId.value, payload);
      Message.success("File updated");
    } else {
      if (!uploadFileValue.value) {
        Message.error("Select a file to upload");
        return;
      }
      await uploadFile(form.storageId, uploadFileValue.value, tagInput.value.split(",").map((item) => item.trim()).filter(Boolean));
      Message.success("File uploaded");
    }
    closeModal();
    await loadData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to save file");
  } finally {
    submitting.value = false;
  }
}

async function removeFile(record: FileRecord) {
  try {
    await deleteFile(record.id);
    Message.success(`Deleted file ${record.name}`);
    await loadData();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to delete file");
  }
}

function beforeUpload(fileItem: File) {
  uploadFileValue.value = fileItem;
  form.name = fileItem.name;
  return false;
}

function downloadFile(record: FileRecord) {
  window.open(buildFileDownloadUrl(record.id), "_blank");
}

function getPathSegments(path: string) {
  return path.split("/").filter(Boolean);
}

function buildPath(segments: string[]) {
  return segments.length ? `/${segments.join("/")}` : "/";
}

function openFolder(path: string) {
  currentPath.value = path;
}

function goToParentFolder() {
  currentPath.value = buildPath(getPathSegments(currentPath.value).slice(0, -1));
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
        <a-button type="primary" @click="openCreateModal">Upload File</a-button>
      </a-space>
    </PageHeader>

    <a-card class="panel-card">
      <div class="toolbar-row files-toolbar">
        <a-select v-model="selectedStorage" class="files-toolbar__storage">
          <a-option value="all">All storage targets</a-option>
          <a-option v-for="item in storageTargets" :key="item.id" :value="String(item.id)">{{ item.name }}</a-option>
        </a-select>
        <a-input-search v-model="search" class="files-toolbar__search" placeholder="Search by name or path" allow-clear />
      </div>
    </a-card>

    <a-card class="panel-card" title="File Inventory">
      <div class="file-path-bar">
        <a-button size="small" :disabled="currentPath === '/'" @click="goToParentFolder">Up</a-button>
        <a-breadcrumb>
          <a-breadcrumb-item>
            <a-button type="text" size="small" @click="openFolder('/')">Root</a-button>
          </a-breadcrumb-item>
          <a-breadcrumb-item v-for="item in pathBreadcrumbs" :key="item.path">
            <a-button type="text" size="small" @click="openFolder(item.path)">{{ item.name }}</a-button>
          </a-breadcrumb-item>
        </a-breadcrumb>
      </div>

      <a-table :data="inventoryRows" :loading="loading" :pagination="{ pageSize: 8 }" row-key="id">
        <template #columns>
          <a-table-column title="Name" data-index="name">
            <template #cell="{ record }">
              <a-button v-if="record.kind === 'folder'" class="file-name-button" type="text" @click="openFolder(record.path)">
                <template #icon><IconFolder /></template>
                {{ record.name }}
              </a-button>
              <span v-else class="file-name-cell">
                <IconFile />
                {{ record.name }}
              </span>
            </template>
          </a-table-column>
          <a-table-column title="Path" data-index="path" />
          <a-table-column title="Type" data-index="type" />
          <a-table-column title="Size" data-index="size" />
          <a-table-column title="Tags">
            <template #cell="{ record }">
              <a-space v-if="record.kind === 'file'" wrap>
                <a-tag v-for="tag in record.tags" :key="tag" color="gray">{{ tag }}</a-tag>
              </a-space>
            </template>
          </a-table-column>
          <a-table-column title="Modified At" data-index="modifiedAt" />
          <a-table-column title="Storage" data-index="storage" />
          <a-table-column title="Actions" :width="110">
            <template #cell="{ record }">
              <a-space v-if="record.kind === 'file'" size="mini">
                <a-button class="action-icon-button" type="text" @click="downloadFile(record)">
                  <template #icon><IconDownload /></template>
                </a-button>
                <a-button class="action-icon-button" type="text" @click="openEditModal(record)">
                  <template #icon><IconEdit /></template>
                </a-button>
                <a-popconfirm content="Delete this file?" @ok="removeFile(record)">
                  <a-button class="action-icon-button action-icon-button--danger" type="text" status="danger">
                    <template #icon><IconDelete /></template>
                  </a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-modal :visible="modalVisible" :title="modalTitle" :confirm-loading="submitting" unmount-on-close @ok="submitFile" @cancel="closeModal">
      <a-form :model="form" layout="vertical" class="modal-bordered-form">
        <a-form-item field="name" label="File Name" required>
          <a-input v-model="form.name" :disabled="!editingId" placeholder="Enter file name" />
        </a-form-item>
        <a-form-item v-if="editingId" field="path" label="Path" required>
          <a-input v-model="form.path" placeholder="Enter file path" />
        </a-form-item>
        <a-form-item v-if="editingId" field="type" label="Type">
          <a-input v-model="form.type" placeholder="jpg / png / zip" />
        </a-form-item>
        <a-form-item v-if="editingId" field="size" label="Size">
          <a-input v-model="form.size" placeholder="18.4 MB" />
        </a-form-item>
        <a-form-item v-if="!editingId" field="upload" label="Select File" required>
          <a-upload :auto-upload="false" :show-file-list="true" :limit="1" @before-upload="beforeUpload">
            <template #upload-button>
              <a-button>Select File</a-button>
            </template>
          </a-upload>
        </a-form-item>
        <a-form-item field="storageId" label="Storage" required>
          <a-select v-model="form.storageId">
            <a-option v-for="item in storageTargets" :key="item.id" :value="item.id">{{ item.name }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item v-if="editingId" field="modifiedAt" label="Modified At">
          <a-input v-model="form.modifiedAt" placeholder="YYYY-MM-DD HH:mm:ss" />
        </a-form-item>
        <a-form-item field="tags" label="Tags">
          <a-input v-model="tagInput" placeholder="Comma separated tags" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.files-toolbar {
  flex-wrap: nowrap;
  justify-content: flex-start;
}

.files-toolbar :deep(.arco-select) {
  flex: 0 0 200px;
  width: 200px;
  min-width: 0;
}

.files-toolbar :deep(.arco-input-search) {
  flex: 1 1 auto;
  min-width: 0;
}

.file-path-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.file-name-button {
  padding: 0;
  color: var(--text-1);
}

.file-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
</style>
