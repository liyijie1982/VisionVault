<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  IconApps,
  IconArchive,
  IconBarChart,
  IconCommon,
  IconCommand,
  IconComputer,
  IconDashboard,
  IconFile,
  IconFolder,
  IconHistory,
  IconList,
  IconLock,
  IconSafe,
  IconSettings,
  IconStorage,
  IconUser,
  IconUserGroup
} from "@arco-design/web-vue/es/icon";
import { appMenus, filterMenusByAccess, type AppMenuItem } from "../constants/menu";
import { changePassword, getAuthState, logout, refreshAuthState } from "../utils/auth";
import { Message } from "@arco-design/web-vue";

const route = useRoute();
const router = useRouter();
const collapsed = ref(false);
const authState = ref(getAuthState());
const changingPassword = ref(false);
const passwordForm = reactive({
  currentPassword: "",
  newPassword: "",
  confirmPassword: ""
});
const forcePasswordChange = computed(() => authState.value?.user.passwordResetRequired === 1);

function syncAuthState() {
  authState.value = getAuthState();
}

onMounted(async () => {
  if (!authState.value) {
    return;
  }
  try {
    await refreshAuthState();
    syncAuthState();
  } catch {
    // Ignore refresh failures here and let route guards handle expired sessions.
  }
});

type MenuItem = {
  key: string;
  label: string;
  path?: string;
  icon: string;
  children?: MenuItem[];
};

const iconMap: Record<string, unknown> = {
  apps: IconApps,
  archive: IconArchive,
  "bar-chart": IconBarChart,
  common: IconCommon,
  command: IconCommand,
  computer: IconComputer,
  dashboard: IconDashboard,
  file: IconFile,
  folder: IconFolder,
  history: IconHistory,
  list: IconList,
  lock: IconLock,
  safe: IconSafe,
  settings: IconSettings,
  storage: IconStorage,
  user: IconUser,
  "user-group": IconUserGroup
};

const menus = computed<MenuItem[]>(() => filterMenusByAccess(appMenus, authState.value?.user.menuKeys ?? []) as AppMenuItem[]);

function resolveIcon(name: string) {
  return iconMap[name] ?? IconFile;
}

function findParentKey(items: MenuItem[], targetKey: string): string | null {
  for (const item of items) {
    if (!item.children) {
      continue;
    }
    if (item.children.some((child) => child.key === targetKey)) {
      return item.key;
    }
  }
  return null;
}

const selectedMenuKey = computed(() => String(route.name ?? "overview"));
const openMenuKeys = computed(() => {
  const parentKey = findParentKey(menus.value, selectedMenuKey.value);
  return parentKey ? [parentKey] : [];
});

function handleMenuClick(path: string) {
  router.push(path);
}

async function handleLogout() {
  await logout();
  syncAuthState();
  Message.success("Signed out");
  router.push("/login");
}

async function handleForcedPasswordChange() {
  if (!passwordForm.currentPassword.trim() || !passwordForm.newPassword.trim()) {
    Message.error("Current password and new password are required");
    return;
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    Message.error("Passwords do not match");
    return;
  }

  changingPassword.value = true;
  try {
    await changePassword(passwordForm.currentPassword, passwordForm.newPassword);
    syncAuthState();
    passwordForm.currentPassword = "";
    passwordForm.newPassword = "";
    passwordForm.confirmPassword = "";
    Message.success("Password updated");
  } catch (error) {
    Message.error(error instanceof Error ? error.message : "Failed to change password");
  } finally {
    changingPassword.value = false;
  }
}
</script>

<template>
  <a-layout class="app-shell">
    <a-layout-sider
      :collapsed="collapsed"
      :width="220"
      :collapsed-width="72"
      class="app-shell__sider"
    >
      <div class="brand-block">
        <img class="brand-mark brand-mark--image" src="/logo.png" alt="SkyView logo" />
        <div v-if="!collapsed" class="brand-copy">
          <div class="brand-copy__title">VisionVault</div>
          <div class="brand-copy__subtitle">Control Plane</div>
        </div>
      </div>

      <a-menu :selected-keys="[selectedMenuKey]" :default-open-keys="openMenuKeys" auto-open :collapsed="collapsed">
        <template v-for="item in menus" :key="item.key">
          <a-sub-menu v-if="item.children?.length" :key="item.key">
            <template #icon>
              <component :is="resolveIcon(item.icon)" />
            </template>
            <template #title>{{ item.label }}</template>
            <a-menu-item v-for="child in item.children" :key="child.key" @click="child.path && handleMenuClick(child.path)">
              <component :is="resolveIcon(child.icon)" />
              <span>{{ child.label }}</span>
            </a-menu-item>
          </a-sub-menu>
          <a-menu-item v-else :key="item.key" @click="item.path && handleMenuClick(item.path)">
            <component :is="resolveIcon(item.icon)" />
            <span>{{ item.label }}</span>
          </a-menu-item>
        </template>
      </a-menu>
    </a-layout-sider>

    <a-layout class="app-shell__main">
      <a-layout-header class="app-shell__header">
        <div class="header-leading">
          <div>
            <div class="header-title">{{ route.meta.title }}</div>
            <div class="header-subtitle">VisionVault operational visibility and management</div>
          </div>
        </div>

        <div class="header-actions">
          <a-tag color="arcoblue" bordered>Development</a-tag>
          <a-input-search placeholder="Search agents, files, or versions" allow-clear class="header-search" />
          <a-dropdown>
            <a-button type="text" class="header-user">
              <IconUser />
              <span>{{ authState?.displayName ?? "Operator" }}</span>
            </a-button>
            <template #content>
              <a-doption>Profile</a-doption>
              <a-doption @click="handleLogout">Sign Out</a-doption>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <a-layout-content class="app-shell__content">
        <router-view />
      </a-layout-content>
    </a-layout>

    <a-modal
      :visible="forcePasswordChange"
      title="Change Password"
      modal-class="force-password-modal"
      :confirm-loading="changingPassword"
      :closable="false"
      :mask-closable="false"
      :esc-to-close="false"
      :hide-cancel="true"
      @ok="handleForcedPasswordChange"
    >
      <a-alert type="warning" show-icon>
        Your password was just created or reset. Change it now before continuing.
      </a-alert>
      <a-form :model="passwordForm" layout="vertical" class="force-password-form">
        <a-form-item field="currentPassword" label="Current Password" required>
          <a-input-password v-model="passwordForm.currentPassword" placeholder="Enter current password" />
        </a-form-item>
        <a-form-item field="newPassword" label="New Password" required>
          <a-input-password v-model="passwordForm.newPassword" placeholder="Enter new password" />
        </a-form-item>
        <a-form-item field="confirmPassword" label="Confirm New Password" required>
          <a-input-password v-model="passwordForm.confirmPassword" placeholder="Re-enter new password" />
        </a-form-item>
      </a-form>
    </a-modal>
  </a-layout>
</template>

<style scoped>
.force-password-form {
  margin-top: 16px;
}
</style>

<style>
.force-password-modal .arco-input-wrapper,
.force-password-modal .arco-input-wrapper.arco-input-password {
  background: #fff;
  border: 1px solid #b7c3d1;
  box-shadow: none;
}

.force-password-modal .arco-input-wrapper:hover,
.force-password-modal .arco-input-wrapper.arco-input-password:hover {
  border-color: #8fa4bc;
}

.force-password-modal .arco-input-wrapper.arco-input-focus,
.force-password-modal .arco-input-wrapper.arco-input-password.arco-input-focus {
  border-color: #315f92;
  box-shadow: 0 0 0 2px rgba(60, 110, 168, 0.12);
}
</style>
