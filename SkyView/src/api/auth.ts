import { request } from "./http";
import type { AuthUser, CaptchaResponse, ChangePasswordPayload, LoginResponse } from "../types";

export function fetchCaptcha() {
  return request<CaptchaResponse>("/api/v1/auth/captcha");
}

export function loginRequest(username: string, password: string, code: string, uuid: string) {
  return request<LoginResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password, code, uuid })
  });
}

export function fetchCurrentUser() {
  return request<AuthUser>("/api/v1/auth/me");
}

export function logoutRequest() {
  return request<{ signedOut: boolean }>("/api/v1/auth/logout", {
    method: "POST"
  });
}

export function changePasswordRequest(payload: ChangePasswordPayload) {
  return request<AuthUser>("/api/v1/auth/change-password", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
