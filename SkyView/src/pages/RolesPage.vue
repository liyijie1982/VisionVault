<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { Message, type TreeNodeData } from "@arco-design/web-vue";
import { IconApps, IconCheckCircle, IconDelete, IconEdit, IconMenu, IconSafe } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { createRole, deleteRole, fetchMenus, fetchRoles, updateRole } from "../api/skybase";
import type { MenuRecord, RolePayload, RoleRecord } from "../types";

const roles = ref<RoleRecord[]>([]);
const menus = ref<MenuRecord[]>([]);
const loading = ref(false);
const submitting = ref(false);
const modalVisible = ref(false);
const editingId = ref<number | null>(null);
const keyword = ref("");
const statusFilter = ref("all");

const form = reactive<RolePayload>({
  name: "",
  key: "",
  dataScope: "custom",
  sort: 0,
  status: 1,
  description: "",
  menuKeys: []
});

const dataScopeOptions = [
  { label: "All Data", value: "all" },
  { label: "Department & Children", value: "dept_and_children" },
  { label: "Department Only", value: "dept_only" },
  { label: "Custom", value: "custom" },
  { label: "Self Only", value: "self" }
];

const filteredRoles = computed(() =>
  roles.value.filter((item) => {
    const query = keyword.value.trim().toLowerCase();
    const matchesKeyword =
      !query ||
      item.name.toLowerCase().includes(query) ||
      item.key.toLowerCase().includes(query) ||
      item.description.toLowerCase().includes(query);
    const matchesStatus =
      statusFilter.value === "all" || String(item.status) === statusFilter.value;
    return matchesKeyword && matchesStatus;
  })
);

const activeCount = computed(() => roles.value.filter((item) => item.status === 1).length);
const disabledCount = computed(() => roles.value.filter((item) => item.status !== 1).length);
const mappedMenuCount = computed(() => new Set(roles.value.flatMap((item) => item.menuKeys)).size);

const modalTitle = computed(() => (editingId.value ? "Edit Role" : "Create Role"));
const menuTreeData = computed(() => buildMenuTree(menus.value));

function dataScopeLabel(value: string) {
  return dataScopeOptions.find((item) => item.value === value)?.label ?? value;
}

function menuName(routeName: string) {
  return menus.value.find((item) => item.routeName === routeName)?.name ?? routeName;
}

function resetForm() {
  editingId.value = null;
  form.name = "";
  form.key = "";
  form.dataScope = "custom";
  form.sort = 0;
  form.status = 1;
  form.description = "";
  form.menuKeys = [];
}

function openCreateModal() {
  resetForm();
  modalVisible.value = true;
}

function openEditModal(record: RoleRecord) {
  editingId.value = record.id;
  form.name = record.name;
  form.key = record.key;
  form.dataScope = record.dataScope;
  form.sort = record.sort;
  form.status = record.status;
  form.description = record.description;
  form.menuKeys = [...record.menuKeys];
  modalVisible.value = true;
}

function closeModal() {
  modalVisible.value = false;
  resetForm();
}

async function loadRoles() {
  loading.value = true;
  try {
    [roles.value, menus.value] = await Promise.all([fetchRoles(), fetchMenus()]);
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load roles");
  } finally {
    loading.value = false;
  }
}

async function submitRole() {
  if (!form.name.trim() || !form.key.trim()) {
    Message.error("Role name and key are required");
    return;
  }
  if (form.menuKeys.length === 0) {
    Message.error("Select at least one accessible menu");
    return;
  }

  submitting.value = true;
  try {
    if (editingId.value) {
      await updateRole(editingId.value, { ...form, name: form.name.trim(), key: form.key.trim(), description: form.description.trim() });
      Message.success("Role updated");
    } else {
      await createRole({ ...form, name: form.name.trim(), key: form.key.trim(), description: form.description.trim() });
      Message.success("Role created");
    }
    closeModal();
    await loadRoles();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to save role");
  } finally {
    submitting.value = false;
  }
}

async function removeRole(record: RoleRecord) {
  try {
    await deleteRole(record.id);
    Message.success(`Deleted role ${record.name}`);
    await loadRoles();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to delete role");
  }
}

onMounted(() => {
  loadRoles();
});

function buildMenuTree(items: MenuRecord[]): TreeNodeData[] {
  const nodeMap = new Map<number, TreeNodeData>();
  const roots: TreeNodeData[] = [];

  items
    .filter((item) => item.status === 1 && item.visible === 1)
    .sort((left, right) => left.sort - right.sort || left.id - right.id)
    .forEach((item) => {
      const node = {
        key: item.routeName,
        title: item.name,
        disableCheckbox: item.menuType === "directory",
        children: [] as TreeNodeData[]
      } satisfies TreeNodeData;
      nodeMap.set(item.id, node);
      if (item.parentId === 0) {
        roots.push(node);
        return;
      }
      const parent = nodeMap.get(item.parentId);
      if (parent) {
        (parent.children ??= []).push(node);
      } else {
        roots.push(node);
      }
    });

  return roots;
}
</script>

<template>
  <div class="page-grid">
    <PageHeader title="" description="">
      <a-space>
        <a-button @click="loadRoles">Refresh</a-button>
        <a-button type="primary" @click="openCreateModal">Create Role</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Total Roles" :value="roles.length" hint="Current visible role definitions" :icon="IconApps" />
      <StatCard label="Active Roles" :value="activeCount" hint="Roles available for assignment" :icon="IconCheckCircle" />
      <StatCard label="Disabled Roles" :value="disabledCount" hint="Temporarily unavailable roles" :icon="IconSafe" />
      <StatCard label="Mapped Menus" :value="mappedMenuCount" hint="Distinct pages already assigned to roles" :icon="IconMenu" />
    </div>

    <a-card class="panel-card">
      <div class="toolbar-row role-toolbar">
        <a-input-search v-model="keyword" placeholder="Search by role name, key, or description" allow-clear />
        <a-select v-model="statusFilter" style="width: 220px">
          <a-option value="all">All statuses</a-option>
          <a-option value="1">Active</a-option>
          <a-option value="0">Disabled</a-option>
        </a-select>
      </div>
    </a-card>

    <a-card class="panel-card" title="Role Inventory">
      <a-table :data="filteredRoles" :loading="loading" :pagination="{ pageSize: 8 }" row-key="id">
        <template #columns>
          <a-table-column title="Role Name" data-index="name" />
          <a-table-column title="Role Key" data-index="key" />
          <a-table-column title="Data Scope">
            <template #cell="{ record }">
              {{ dataScopeLabel(record.dataScope) }}
            </template>
          </a-table-column>
          <a-table-column title="Status" :width="110">
            <template #cell="{ record }">
              <a-tag :color="record.status === 1 ? 'green' : 'gray'">{{ record.status === 1 ? "Active" : "Disabled" }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Description">
            <template #cell="{ record }">
              <span class="role-description">{{ record.description || "No description" }}</span>
            </template>
          </a-table-column>
          <a-table-column title="Accessible Menus" :width="280">
            <template #cell="{ record }">
              <a-space wrap size="mini">
                <a-tag v-for="menuKey in record.menuKeys" :key="menuKey" color="arcoblue">
                  {{ menuName(menuKey) }}
                </a-tag>
              </a-space>
            </template>
          </a-table-column>
          <a-table-column title="Updated At" data-index="updatedAt" :width="170" />
          <a-table-column title="Actions" :width="110" fixed="right">
            <template #cell="{ record }">
              <a-space size="mini">
                <a-tooltip content="Edit">
                  <a-button class="action-icon-button" type="text" @click="openEditModal(record)">
                    <template #icon><IconEdit /></template>
                  </a-button>
                </a-tooltip>
                <a-popconfirm content="Delete this role?" @ok="removeRole(record)">
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
      @ok="submitRole"
      @cancel="closeModal"
    >
      <a-form :model="form" layout="vertical" class="role-form">
        <a-form-item field="name" label="Role Name" required>
          <a-input v-model="form.name" placeholder="Enter role name" />
        </a-form-item>
        <a-form-item field="key" label="Role Key" required>
          <a-input v-model="form.key" placeholder="Enter role key" />
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
        <a-form-item field="menuKeys" required>
          <template #label>
            <span class="role-form-label">
              Accessible Menus
              <a-tooltip content="Choose the pages this role can see in the left navigation and open directly by route.">
                <IconInfoCircle class="role-form-label__icon" />
              </a-tooltip>
            </span>
          </template>
          <a-tree
            v-model:checked-keys="form.menuKeys"
            :data="menuTreeData"
            checkable
            checked-strategy="child"
          />
        </a-form-item>
        <a-form-item field="description" label="Description">
          <a-textarea v-model="form.description" :max-length="500" placeholder="Describe the role intent and permission boundary" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.role-description {
  color: var(--vv-text-secondary);
}

.role-toolbar {
  flex-wrap: nowrap;
  align-items: center;
}

.role-toolbar :deep(.arco-input-search) {
  flex: 1 1 auto;
  min-width: 0;
}

.role-toolbar :deep(.arco-select) {
  flex: 0 0 220px;
}

.role-form :deep(.arco-input-wrapper),
.role-form :deep(.arco-select-view),
.role-form :deep(.arco-textarea-wrapper),
.role-form :deep(.arco-input-number) {
  background: #fff;
  border: 1px solid #c9d2de;
  box-shadow: none;
}

.role-form :deep(.arco-input-wrapper:hover),
.role-form :deep(.arco-select-view:hover),
.role-form :deep(.arco-textarea-wrapper:hover),
.role-form :deep(.arco-input-number:hover) {
  border-color: #9fb2c8;
}

.role-form :deep(.arco-input-wrapper.arco-input-focus),
.role-form :deep(.arco-select-view.arco-select-view-focus),
.role-form :deep(.arco-textarea-wrapper-focus),
.role-form :deep(.arco-input-number.arco-input-number-focus) {
  border-color: #3c6ea8;
  box-shadow: 0 0 0 2px rgba(60, 110, 168, 0.12);
}

.role-form-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.role-form-label__icon {
  color: var(--color-text-3);
  cursor: help;
}
</style>
