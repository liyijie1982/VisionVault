import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import AppLayout from "../layouts/AppLayout.vue";
import LoginPage from "../pages/LoginPage.vue";
import OverviewPage from "../pages/OverviewPage.vue";
import AgentsPage from "../pages/AgentsPage.vue";
import GroupsPage from "../pages/GroupsPage.vue";
import BackupDevicesPage from "../pages/BackupDevicesPage.vue";
import BackupLogsPage from "../pages/BackupLogsPage.vue";
import BackupTasksPage from "../pages/BackupTasksPage.vue";
import StoragePage from "../pages/StoragePage.vue";
import FilesPage from "../pages/FilesPage.vue";
import SyncLogsPage from "../pages/SyncLogsPage.vue";
import MonitorPage from "../pages/MonitorPage.vue";
import VersionsPage from "../pages/VersionsPage.vue";
import SystemPage from "../pages/SystemPage.vue";
import SystemModulePage from "../pages/SystemModulePage.vue";
import LoginLogsPage from "../pages/LoginLogsPage.vue";
import UsersPage from "../pages/UsersPage.vue";
import DepartmentsPage from "../pages/DepartmentsPage.vue";
import RolesPage from "../pages/RolesPage.vue";
import { findFirstAccessiblePath, filterMenusByAccess, appMenus } from "../constants/menu";
import { clearAuthState, getAuthState, hasMenuAccess, isLoggedIn, refreshAuthState } from "../utils/auth";

const routes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "login",
    component: LoginPage,
    meta: { public: true, title: "Sign In" }
  },
  {
    path: "/",
    component: AppLayout,
    children: [
      { path: "", name: "overview", component: OverviewPage, meta: { title: "Overview", menuKey: "overview" } },
      { path: "backup/devices", name: "backup-devices", component: BackupDevicesPage, meta: { title: "Tape Devices", menuKey: "backup-devices" } },
      { path: "backup/tasks", name: "backup-tasks", component: BackupTasksPage, meta: { title: "Task Management", menuKey: "backup-tasks" } },
      { path: "backup/logs", name: "backup-logs", component: BackupLogsPage, meta: { title: "Backup Logs", menuKey: "backup-logs" } },
      { path: "agents", name: "agents", component: AgentsPage, meta: { title: "Agents", menuKey: "agents" } },
      { path: "groups", name: "groups", component: GroupsPage, meta: { title: "Groups", menuKey: "groups" } },
      { path: "storage", name: "storage", component: StoragePage, meta: { title: "Storage", menuKey: "storage" } },
      { path: "files", name: "files", component: FilesPage, meta: { title: "Files", menuKey: "files" } },
      { path: "sync-logs", name: "sync-logs", component: SyncLogsPage, meta: { title: "Sync Logs", menuKey: "sync-logs" } },
      { path: "monitor", name: "monitor", component: MonitorPage, meta: { title: "Monitor", menuKey: "monitor" } },
      { path: "extraction-rules", name: "extraction-rules", component: SystemModulePage, meta: { title: "Extraction Rules", menuKey: "extraction-rules" } },
      { path: "versions", name: "versions", component: VersionsPage, meta: { title: "Versions", menuKey: "versions" } },
      { path: "file-logs", name: "file-logs", component: SystemModulePage, meta: { title: "File Logs", menuKey: "file-logs" } },
      { path: "file-permissions", name: "file-permissions", component: SystemModulePage, meta: { title: "File Permissions", menuKey: "file-permissions" } },
      { path: "task-progress", name: "task-progress", component: SystemModulePage, meta: { title: "Task Progress", menuKey: "task-progress" } },
      { path: "alerts/groups", name: "alert-groups", component: SystemModulePage, meta: { title: "Alert Groups", menuKey: "alert-groups" } },
      { path: "alerts/logs", name: "alert-logs", component: SystemModulePage, meta: { title: "Alert Logs", menuKey: "alert-logs" } },
      { path: "alerts/channels", name: "message-channels", component: SystemModulePage, meta: { title: "Message Channels", menuKey: "message-channels" } },
      { path: "alerts/policies", name: "alert-policies", component: SystemModulePage, meta: { title: "Alert Policies", menuKey: "alert-policies" } },
      { path: "system", name: "system", component: SystemPage, meta: { title: "System", menuKey: "system-users" } },
      { path: "system/users", name: "system-users", component: UsersPage, meta: { title: "Users", menuKey: "system-users" } },
      { path: "system/departments", name: "system-departments", component: DepartmentsPage, meta: { title: "Departments", menuKey: "system-departments" } },
      { path: "system/roles", name: "system-roles", component: RolesPage, meta: { title: "Roles", menuKey: "system-roles" } },
      { path: "system/login-logs", name: "system-login-logs", component: LoginLogsPage, meta: { title: "Login Logs", menuKey: "system-login-logs" } },
      { path: "system/parameters", name: "system-parameters", component: SystemModulePage, meta: { title: "Settings", menuKey: "system-parameters" } },
      { path: "system/license", name: "system-license", component: SystemModulePage, meta: { title: "License", menuKey: "system-license" } }
    ]
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach(async (to) => {
  document.title = `VisionVault | ${String(to.meta.title ?? "Control Plane")}`;
  if (to.meta.public) {
    if (to.name === "login" && isLoggedIn()) {
      return { name: "overview" };
    }
    return true;
  }
  if (isLoggedIn()) {
    if (!hasMenuAccess(String(to.meta.menuKey ?? ""))) {
      const filteredMenus = filterMenusByAccess(appMenus, getAuthState()?.user.menuKeys ?? []);
      return findFirstAccessiblePath(filteredMenus);
    }
    return true;
  }

  try {
    await refreshAuthState();
    if (!hasMenuAccess(String(to.meta.menuKey ?? ""))) {
      const filteredMenus = filterMenusByAccess(appMenus, getAuthState()?.user.menuKeys ?? []);
      return findFirstAccessiblePath(filteredMenus);
    }
    return true;
  } catch {
    clearAuthState();
    return { name: "login", query: { redirect: to.fullPath } };
  }
});

export default router;
