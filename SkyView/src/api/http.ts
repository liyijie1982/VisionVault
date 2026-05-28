import type { ApiEnvelope } from "../types";

export const API_BASE = (import.meta.env.VITE_API_BASE as string | undefined) ?? "http://localhost:8080";

export function buildApiUrl(path: string) {
  return `${API_BASE}${path}`;
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(buildApiUrl(path), {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    },
    ...init
  });

  let payload: ApiEnvelope<T> | null = null;
  const contentType = response.headers.get("Content-Type") ?? "";
  if (contentType.includes("application/json")) {
    payload = (await response.json()) as ApiEnvelope<T>;
  }

  if (!response.ok) {
    throw new Error(payload?.msg || `Request failed with status ${response.status}`);
  }

  if (!payload) {
    throw new Error("Response payload is empty");
  }
  if (payload.code !== 0) {
    throw new Error(payload.msg || "Request failed");
  }
  return payload.data;
}
