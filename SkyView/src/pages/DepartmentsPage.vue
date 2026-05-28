<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconApps, IconCheckCircle, IconDelete, IconEdit, IconUser, IconUserGroup } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { createDepartment, deleteDepartment, fetchDepartments, updateDepartment } from "../api/skybase";
import type { DepartmentPayload, DepartmentRecord } from "../types";

const departments = ref<DepartmentRecord[]>([]);
const loading = ref(false);
const submitting = ref(false);
const modalVisible = ref(false);
const editingId = ref<number | null>(null);
const keyword = ref("");
const statusFilter = ref("all");

const form = reactive<DepartmentPayload>({
  parentId: 0,
  name: "",
  leader: "",
  phone: "",
  email: "",
  sort: 0,
  status: 1
});

const filteredDepartments = computed(() =>
  departments.value.filter((item) => {
    const query = keyword.value.trim().toLowerCase();
    const matchesKeyword =
      !query ||
      item.name.toLowerCase().includes(query) ||
      item.leader.toLowerCase().includes(query) ||
      item.phone.toLowerCase().includes(query);
    const matchesStatus = statusFilter.value === "all" || String(item.status) === statusFilter.value;
    return matchesKeyword && matchesStatus;
  })
);

const activeCount = computed(() => departments.value.filter((item) => item.status === 1).length);
const rootCount = computed(() => departments.value.filter((item) => item.parentId === 0).length);
const leaderCoverage = computed(() => departments.value.filter((item) => item.leader.trim()).length);
const modalTitle = computed(() => (editingId.value ? "Edit Department" : "Create Department"));

const parentOptions = computed(() => departments.value.filter((item) => item.id !== editingId.value));

function parentDepartmentName(parentId: number) {
  if (parentId === 0) {
    return "Top Level";
  }
  return departments.value.find((item) => item.id === parentId)?.name ?? "Unknown";
}

function resetForm() {
  editingId.value = null;
  form.parentId = 0;
  form.name = "";
  form.leader = "";
  form.phone = "";
  form.email = "";
  form.sort = 0;
  form.status = 1;
}

function openCreateModal() {
  resetForm();
  modalVisible.value = true;
}

function openEditModal(record: DepartmentRecord) {
  editingId.value = record.id;
  form.parentId = record.parentId;
  form.name = record.name;
  form.leader = record.leader;
  form.phone = record.phone;
  form.email = record.email;
  form.sort = record.sort;
  form.status = record.status;
  modalVisible.value = true;
}

function closeModal() {
  modalVisible.value = false;
  resetForm();
}

async function loadDepartments() {
  loading.value = true;
  try {
    departments.value = await fetchDepartments();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load departments");
  } finally {
    loading.value = false;
  }
}

async function submitDepartment() {
  if (!form.name.trim()) {
    Message.error("Department name is required");
    return;
  }

  submitting.value = true;
  try {
    const payload = {
      ...form,
      name: form.name.trim(),
      leader: form.leader.trim(),
      phone: form.phone.trim(),
      email: form.email.trim()
    };

    if (editingId.value) {
      await updateDepartment(editingId.value, payload);
      Message.success("Department updated");
    } else {
      await createDepartment(payload);
      Message.success("Department created");
    }
    closeModal();
    await loadDepartments();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to save department");
  } finally {
    submitting.value = false;
  }
}

async function removeDepartment(record: DepartmentRecord) {
  try {
    await deleteDepartment(record.id);
    Message.success(`Deleted department ${record.name}`);
    await loadDepartments();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to delete department");
  }
}

onMounted(() => {
  loadDepartments();
});
</script>

<template>
  <div class="page-grid">
    <PageHeader title="" description="">
      <a-space>
        <a-button @click="loadDepartments">Refresh</a-button>
        <a-button type="primary" @click="openCreateModal">Create Department</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Departments" :value="departments.length" hint="Current department records" :icon="IconApps" />
      <StatCard label="Active" :value="activeCount" hint="Departments available for assignment" :icon="IconCheckCircle" />
      <StatCard label="Root Nodes" :value="rootCount" hint="Top-level structures" :icon="IconUserGroup" />
      <StatCard label="Leader Coverage" :value="`${leaderCoverage}/${departments.length || 0}`" hint="Departments with assigned leaders" :icon="IconUser" />
    </div>

    <a-card class="panel-card">
      <div class="toolbar-row department-toolbar">
        <a-input-search v-model="keyword" placeholder="Search by name, leader, or phone" allow-clear />
        <a-select v-model="statusFilter" style="width: 220px">
          <a-option value="all">All statuses</a-option>
          <a-option value="1">Active</a-option>
          <a-option value="0">Disabled</a-option>
        </a-select>
      </div>
    </a-card>

    <a-card class="panel-card" title="Department Directory">
      <a-table :data="filteredDepartments" :loading="loading" :pagination="{ pageSize: 8 }" row-key="id">
        <template #columns>
          <a-table-column title="Department Name" data-index="name" />
          <a-table-column title="Parent">
            <template #cell="{ record }">
              {{ parentDepartmentName(record.parentId) }}
            </template>
          </a-table-column>
          <a-table-column title="Leader" data-index="leader" />
          <a-table-column title="Phone" data-index="phone" />
          <a-table-column title="Email" data-index="email" />
          <a-table-column title="Sort" data-index="sort" :width="90" />
          <a-table-column title="Status" :width="110">
            <template #cell="{ record }">
              <a-tag :color="record.status === 1 ? 'green' : 'gray'">{{ record.status === 1 ? "Active" : "Disabled" }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Updated At" data-index="updatedAt" :width="180" />
          <a-table-column title="Actions" :width="110" fixed="right">
            <template #cell="{ record }">
              <a-space size="mini">
                <a-tooltip content="Edit">
                  <a-button class="action-icon-button" type="text" @click="openEditModal(record)">
                    <template #icon><IconEdit /></template>
                  </a-button>
                </a-tooltip>
                <a-popconfirm content="Delete this department?" @ok="removeDepartment(record)">
                  <a-tooltip content="Delete">
                    <a-button class="action-icon-button action-icon-button--danger" type="text" status="danger">
                      <template #icon><IconDelete /></template>
                    </a-button>
                  </a-tooltip>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <a-modal
      :visible="modalVisible"
      :title="modalTitle"
      :confirm-loading="submitting"
      unmount-on-close
      @ok="submitDepartment"
      @cancel="closeModal"
    >
      <a-form :model="form" layout="vertical" class="department-form">
        <a-form-item field="name" label="Department Name" required>
          <a-input v-model="form.name" placeholder="Enter department name" />
        </a-form-item>
        <a-form-item field="parentId" label="Parent Department">
          <a-select v-model="form.parentId">
            <a-option :value="0">Top Level</a-option>
            <a-option v-for="item in parentOptions" :key="item.id" :value="item.id">{{ item.name }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="leader" label="Leader">
          <a-input v-model="form.leader" placeholder="Enter leader name" />
        </a-form-item>
        <a-form-item field="phone" label="Phone">
          <a-input v-model="form.phone" placeholder="Enter department phone" />
        </a-form-item>
        <a-form-item field="email" label="Email">
          <a-input v-model="form.email" placeholder="Enter department email" />
        </a-form-item>
        <a-form-item field="sort" label="Sort Order">
          <a-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%" />
        </a-form-item>
        <a-form-item field="status" label="Status">
          <a-radio-group v-model="form.status" type="button">
            <a-radio :value="1">Active</a-radio>
            <a-radio :value="0">Disabled</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.department-toolbar {
  flex-wrap: nowrap;
  align-items: center;
}

.department-toolbar :deep(.arco-input-search) {
  flex: 1 1 auto;
  min-width: 0;
}

.department-toolbar :deep(.arco-select) {
  flex: 0 0 220px;
}

.department-form :deep(.arco-input-wrapper),
.department-form :deep(.arco-select-view),
.department-form :deep(.arco-input-number) {
  background: #fff;
  border: 1px solid #c9d2de;
  box-shadow: none;
}

.department-form :deep(.arco-input-wrapper:hover),
.department-form :deep(.arco-select-view:hover),
.department-form :deep(.arco-input-number:hover) {
  border-color: #9fb2c8;
}

.department-form :deep(.arco-input-wrapper.arco-input-focus),
.department-form :deep(.arco-select-view.arco-select-view-focus),
.department-form :deep(.arco-input-number.arco-input-number-focus) {
  border-color: #3c6ea8;
  box-shadow: 0 0 0 2px rgba(60, 110, 168, 0.12);
}
</style>
