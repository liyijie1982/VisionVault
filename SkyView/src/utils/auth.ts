import { changePasswordRequest, fetchCurrentUser, loginRequest, logoutRequest } from "../api/auth";
import type { AuthUser } from "../types";

const AUTH_KEY = "skyview-auth";

export interface AuthState {
  user: AuthUser;
  displayName: string;
}

function toAuthState(user: AuthUser): AuthState {
  return {
    user,
    displayName: user.nickname || user.username
  };
}

export function isLoggedIn() {
  return Boolean(sessionStorage.getItem(AUTH_KEY));
}

export function getAuthState(): AuthState | null {
  const raw = sessionStorage.getItem(AUTH_KEY);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as AuthState;
  } catch {
    sessionStorage.removeItem(AUTH_KEY);
    return null;
  }
}

export function setAuthState(user: AuthUser) {
  sessionStorage.setItem(AUTH_KEY, JSON.stringify(toAuthState(user)));
}

export async function login(username: string, password: string, code: string, uuid: string) {
  const result = await loginRequest(username, password, code, uuid);
  setAuthState(result.user);
  return result.user;
}

export async function refreshAuthState() {
  const user = await fetchCurrentUser();
  setAuthState(user);
  return user;
}

export async function logout() {
  try {
    await logoutRequest();
  } finally {
    sessionStorage.removeItem(AUTH_KEY);
  }
}

export function clearAuthState() {
  sessionStorage.removeItem(AUTH_KEY);
}

export async function changePassword(currentPassword: string, newPassword: string) {
  const user = await changePasswordRequest({ currentPassword, newPassword });
  setAuthState(user);
  return user;
}

export function hasMenuAccess(menuKey?: string | null) {
  if (!menuKey) {
    return true;
  }
  const authState = getAuthState();
  if (!authState) {
    return false;
  }
  return authState.user.menuKeys?.includes(menuKey) ?? false;
}
