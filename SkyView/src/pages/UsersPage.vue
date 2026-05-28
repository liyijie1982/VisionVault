<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { Message } from "@arco-design/web-vue";
import { IconCheckCircle, IconDelete, IconEdit, IconLock, IconSafe, IconUser } from "@arco-design/web-vue/es/icon";
import PageHeader from "../components/PageHeader.vue";
import StatCard from "../components/StatCard.vue";
import { createUser, deleteUser, fetchDepartments, fetchRoles, fetchUsers, resetUserPassword, updateUser } from "../api/skybase";
import type { DepartmentRecord, RoleRecord, UserPayload, UserRecord } from "../types";

const users = ref<UserRecord[]>([]);
const departments = ref<DepartmentRecord[]>([]);
const roles = ref<RoleRecord[]>([]);
const loading = ref(false);
const submitting = ref(false);
const modalVisible = ref(false);
const editingId = ref<number | null>(null);
const resetModalVisible = ref(false);
const resetSubmitting = ref(false);
const resettingUser = ref<UserRecord | null>(null);
const keyword = ref("");
const statusFilter = ref("all");

const form = reactive<UserPayload>({
  deptId: 0,
  username: "",
  nickname: "",
  realName: "",
  phone: "",
  email: "",
  password: "",
  status: 1,
  roleIds: [],
  passwordResetRequired: 0
});

const resetForm = reactive({
  password: "",
  confirmPassword: ""
});

const filteredUsers = computed(() =>
  users.value.filter((item) => {
    const query = keyword.value.trim().toLowerCase();
    const matchesKeyword =
      !query ||
      item.username.toLowerCase().includes(query) ||
      item.nickname.toLowerCase().includes(query) ||
      item.realName.toLowerCase().includes(query) ||
      item.deptName.toLowerCase().includes(query);
    const matchesStatus = statusFilter.value === "all" || String(item.status) === statusFilter.value;
    return matchesKeyword && matchesStatus;
  })
);

const activeCount = computed(() => users.value.filter((item) => item.status === 1).length);
const resetRequiredCount = computed(() => users.value.filter((item) => item.passwordResetRequired === 1).length);
const assignedRoleCount = computed(() => new Set(users.value.flatMap((item) => item.roleIds)).size);
const modalTitle = computed(() => (editingId.value ? "Edit User" : "Create User"));
const isCreating = computed(() => editingId.value === null);

function resetUserForm() {
  editingId.value = null;
  form.deptId = departments.value[0]?.id ?? 0;
  form.username = "";
  form.nickname = "";
  form.realName = "";
  form.phone = "";
  form.email = "";
  form.password = "";
  form.status = 1;
  form.roleIds = [];
  form.passwordResetRequired = 0;
}

function openCreateModal() {
  resetUserForm();
  modalVisible.value = true;
}

function openEditModal(record: UserRecord) {
  editingId.value = record.id;
  form.deptId = record.deptId;
  form.username = record.username;
  form.nickname = record.nickname;
  form.realName = record.realName;
  form.phone = record.phone;
  form.email = record.email;
  form.password = "";
  form.status = record.status;
  form.roleIds = [...record.roleIds];
  form.passwordResetRequired = record.passwordResetRequired;
  modalVisible.value = true;
}

function closeModal() {
  modalVisible.value = false;
  resetUserForm();
}

function openResetPasswordModal(record: UserRecord) {
  resettingUser.value = record;
  resetForm.password = "";
  resetForm.confirmPassword = "";
  resetModalVisible.value = true;
}

function closeResetPasswordModal() {
  resetModalVisible.value = false;
  resettingUser.value = null;
  resetForm.password = "";
  resetForm.confirmPassword = "";
}

async function loadUsers() {
  loading.value = true;
  try {
    [users.value, departments.value, roles.value] = await Promise.all([fetchUsers(), fetchDepartments(), fetchRoles()]);
    if (!form.deptId) {
      form.deptId = departments.value[0]?.id ?? 0;
    }
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to load users");
  } finally {
    loading.value = false;
  }
}

async function submitUser() {
  if (!form.username.trim() || !form.nickname.trim() || !form.phone.trim() || form.deptId <= 0) {
    Message.error("Username, nickname, phone, and department are required");
    return;
  }
  if (isCreating.value && !form.password?.trim()) {
    Message.error("Initial password is required");
    return;
  }

  submitting.value = true;
  try {
    const payload: UserPayload = {
      ...form,
      username: form.username.trim(),
      nickname: form.nickname.trim(),
      realName: form.realName.trim(),
      phone: form.phone.trim(),
      email: form.email.trim(),
      password: form.password?.trim() || undefined
    };

    if (editingId.value) {
      await updateUser(editingId.value, payload);
      Message.success("User updated");
    } else {
      await createUser(payload);
      Message.success("User created");
    }
    closeModal();
    await loadUsers();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to save user");
  } finally {
    submitting.value = false;
  }
}

async function submitResetPassword() {
  if (!resettingUser.value) {
    return;
  }
  if (!resetForm.password.trim()) {
    Message.error("Reset password is required");
    return;
  }
  if (resetForm.password !== resetForm.confirmPassword) {
    Message.error("Passwords do not match");
    return;
  }

  resetSubmitting.value = true;
  try {
    await resetUserPassword(resettingUser.value.id, resetForm.password.trim());
    Message.success(`Password reset for ${resettingUser.value.username}`);
    closeResetPasswordModal();
    await loadUsers();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to reset password");
  } finally {
    resetSubmitting.value = false;
  }
}

async function removeUser(record: UserRecord) {
  try {
    await deleteUser(record.id);
    Message.success(`Deleted user ${record.username}`);
    await loadUsers();
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to delete user");
  }
}

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <div class="page-grid">
    <PageHeader title="" description="">
      <a-space>
        <a-button @click="loadUsers">Refresh</a-button>
        <a-button type="primary" @click="openCreateModal">Create User</a-button>
      </a-space>
    </PageHeader>

    <div class="stats-grid">
      <StatCard label="Managed Users" :value="users.length" hint="Current user records" :icon="IconUser" />
      <StatCard label="Active Users" :value="activeCount" hint="Enabled operator accounts" :icon="IconCheckCircle" />
      <StatCard label="Password Reset Required" :value="resetRequiredCount" hint="Users who must change password on next sign-in" :icon="IconLock" />
      <StatCard label="Assigned Roles" :value="assignedRoleCount" hint="Distinct roles already granted to users" :icon="IconSafe" />
    </div>

    <a-card class="panel-card">
      <div class="toolbar-row user-toolbar">
        <a-input-search v-model="keyword" placeholder="Search by username, nickname, real name, or department" allow-clear />
        <a-select v-model="statusFilter" style="width: 220px">
          <a-option value="all">All statuses</a-option>
          <a-option value="1">Active</a-option>
          <a-option value="0">Disabled</a-option>
        </a-select>
      </div>
    </a-card>

    <a-card class="panel-card" title="User Directory">
      <a-table :data="filteredUsers" :loading="loading" :pagination="{ pageSize: 8 }" row-key="id">
        <template #columns>
          <a-table-column title="Username" data-index="username" :width="150" />
          <a-table-column title="Nickname" data-index="nickname" :width="150" />
          <a-table-column title="Department" data-index="deptName" :width="180" />
          <a-table-column title="Roles" :width="220">
            <template #cell="{ record }">
              <a-space wrap size="mini">
                <a-tag v-for="roleName in record.roleNames" :key="roleName" color="arcoblue">{{ roleName }}</a-tag>
                <span v-if="record.roleNames.length === 0" class="user-role-empty">Unassigned</span>
              </a-space>
            </template>
          </a-table-column>
          <a-table-column title="Reset Password" :width="130">
            <template #cell="{ record }">
              <a-tag :color="record.passwordResetRequired === 1 ? 'orange' : 'green'">
                {{ record.passwordResetRequired === 1 ? "Required" : "Not Required" }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Status" :width="110">
            <template #cell="{ record }">
              <a-tag :color="record.status === 1 ? 'green' : 'gray'">{{ record.status === 1 ? "Active" : "Disabled" }}</a-tag>
            </template>
          </a-table-column>
          <a-table-column title="Updated At" data-index="updatedAt" :width="180" />
          <a-table-column title="Actions" :width="150" fixed="right">
            <template #cell="{ record }">
              <a-space size="mini">
                <a-tooltip content="Edit">
                  <a-button class="action-icon-button" type="text" @click="openEditModal(record)">
                    <template #icon><IconEdit /></template>
                  </a-button>
                </a-tooltip>
                <a-tooltip content="Reset Password">
                  <a-button class="action-icon-button" type="text" @click="openResetPasswordModal(record)">
                    <template #icon><IconLock /></template>
                  </a-button>
                </a-tooltip>
                <a-popconfirm content="Delete this user?" @ok="removeUser(record)">
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
      @ok="submitUser"
      @cancel="closeModal"
    >
      <a-form :model="form" layout="vertical" class="user-form">
        <a-form-item field="username" label="Username" required>
          <a-input v-model="form.username" placeholder="Enter username" />
        </a-form-item>
        <a-form-item field="nickname" label="Nickname" required>
          <a-input v-model="form.nickname" placeholder="Enter nickname" />
        </a-form-item>
        <a-form-item field="realName" label="Real Name">
          <a-input v-model="form.realName" placeholder="Enter real name" />
        </a-form-item>
        <a-form-item field="deptId" label="Department" required>
          <a-select v-model="form.deptId">
            <a-option v-for="item in departments" :key="item.id" :value="item.id">{{ item.name }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item field="phone" label="Phone" required>
          <a-input v-model="form.phone" placeholder="Enter phone" />
        </a-form-item>
        <a-form-item field="email" label="Email">
          <a-input v-model="form.email" placeholder="Enter email" />
        </a-form-item>
        <a-form-item v-if="isCreating" field="password" label="Initial Password" required>
          <a-input-password v-model="form.password" placeholder="Enter initial password" />
        </a-form-item>
        <a-form-item field="status" label="Status">
          <a-radio-group v-model="form.status" type="button">
            <a-radio :value="1">Active</a-radio>
            <a-radio :value="0">Disabled</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item field="roleIds" label="Roles">
          <a-select v-model="form.roleIds" multiple allow-clear placeholder="Select one or more roles">
            <a-option v-for="role in roles.filter((item) => item.status === 1)" :key="role.id" :value="role.id">
              {{ role.name }}
            </a-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <a-modal
      :visible="resetModalVisible"
      :title="resettingUser ? `Reset Password · ${resettingUser.username}` : 'Reset Password'"
      :confirm-loading="resetSubmitting"
      unmount-on-close
      @ok="submitResetPassword"
      @cancel="closeResetPasswordModal"
    >
      <a-form :model="resetForm" layout="vertical" class="user-form">
        <a-form-item field="password" label="New Temporary Password" required>
          <a-input-password v-model="resetForm.password" placeholder="Enter temporary password" />
        </a-form-item>
        <a-form-item field="confirmPassword" label="Confirm Password" required>
          <a-input-password v-model="resetForm.confirmPassword" placeholder="Re-enter temporary password" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.user-role-empty {
  color: var(--vv-text-secondary);
}

.user-toolbar {
  flex-wrap: nowrap;
  align-items: center;
}

.user-toolbar :deep(.arco-input-search) {
  flex: 1 1 auto;
  min-width: 0;
}

.user-toolbar :deep(.arco-select) {
  flex: 0 0 220px;
}

.user-form :deep(.arco-input-wrapper),
.user-form :deep(.arco-input-password),
.user-form :deep(.arco-select-view) {
  background: #fff;
  border: 1px solid #c9d2de;
  box-shadow: none;
}

.user-form :deep(.arco-input-wrapper:hover),
.user-form :deep(.arco-input-password:hover),
.user-form :deep(.arco-select-view:hover) {
  border-color: #9fb2c8;
}

.user-form :deep(.arco-input-wrapper.arco-input-focus),
.user-form :deep(.arco-input-password.arco-input-focus),
.user-form :deep(.arco-select-view.arco-select-view-focus) {
  border-color: #3c6ea8;
  box-shadow: 0 0 0 2px rgba(60, 110, 168, 0.12);
}
</style>
