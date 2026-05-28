export type AppMenuItem = {
  key: string;
  label: string;
  path?: string;
  icon: string;
  children?: AppMenuItem[];
};

export const appMenus: AppMenuItem[] = [
  { key: "overview", label: "Overview", path: "/", icon: "dashboard" },
  {
    key: "agent-control",
    label: "Agent Control",
    icon: "computer",
    children: [
      { key: "agents", label: "Agents", path: "/agents", icon: "computer" },
      { key: "groups", label: "Groups", path: "/groups", icon: "apps" },
      { key: "sync-logs", label: "Sync Logs", path: "/sync-logs", icon: "list" },
      { key: "monitor", label: "Monitor", path: "/monitor", icon: "command" },
      { key: "extraction-rules", label: "Extraction Rules", path: "/extraction-rules", icon: "common" },
      { key: "versions", label: "Versions", path: "/versions", icon: "archive" }
    ]
  },
  {
    key: "file-control",
    label: "File Control",
    icon: "folder",
    children: [
      { key: "files", label: "Files", path: "/files", icon: "folder" },
      { key: "file-logs", label: "File Logs", path: "/file-logs", icon: "history" },
      { key: "file-permissions", label: "File Permissions", path: "/file-permissions", icon: "lock" },
      { key: "task-progress", label: "Task Progress", path: "/task-progress", icon: "list" }
    ]
  },
  {
    key: "alerts",
    label: "Alerts",
    icon: "safe",
    children: [
      { key: "alert-groups", label: "Alert Groups", path: "/alerts/groups", icon: "apps" },
      { key: "alert-logs", label: "Alert Logs", path: "/alerts/logs", icon: "history" },
      { key: "message-channels", label: "Message Channels", path: "/alerts/channels", icon: "common" },
      { key: "alert-policies", label: "Alert Policies", path: "/alerts/policies", icon: "safe" }
    ]
  },
  {
    key: "backup",
    label: "Backup",
    icon: "archive",
    children: [
      { key: "backup-devices", label: "Tape Devices", path: "/backup/devices", icon: "storage" },
      { key: "backup-tasks", label: "Task Management", path: "/backup/tasks", icon: "list" },
      { key: "backup-logs", label: "Backup Logs", path: "/backup/logs", icon: "history" }
    ]
  },
  {
    key: "system",
    label: "System",
    icon: "settings",
    children: [
      { key: "storage", label: "Storage", path: "/storage", icon: "storage" },
      { key: "system-users", label: "Users", path: "/system/users", icon: "user" },
      { key: "system-departments", label: "Departments", path: "/system/departments", icon: "user-group" },
      { key: "system-roles", label: "Roles", path: "/system/roles", icon: "safe" },
      { key: "system-login-logs", label: "Login Logs", path: "/system/login-logs", icon: "history" },
      { key: "system-parameters", label: "Settings", path: "/system/parameters", icon: "common" },
      { key: "system-license", label: "License", path: "/system/license", icon: "lock" }
    ]
  }
];

export function filterMenusByAccess(items: AppMenuItem[], allowedKeys: string[]): AppMenuItem[] {
  const allowed = new Set(allowedKeys);
  return items
    .map((item) => {
      if (item.children?.length) {
        const children: AppMenuItem[] = filterMenusByAccess(item.children, allowedKeys);
        return children.length ? { ...item, children } : null;
      }
      return allowed.has(item.key) ? item : null;
    })
    .filter((item): item is AppMenuItem => item !== null);
}

export function findFirstAccessiblePath(items: AppMenuItem[]): string {
  for (const item of items) {
    if (item.path) {
      return item.path;
    }
    if (item.children?.length) {
      const childPath = findFirstAccessiblePath(item.children);
      if (childPath) {
        return childPath;
      }
    }
  }
  return "/";
}
