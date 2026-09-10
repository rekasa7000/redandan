// Typed fetch wrapper around the Go API (apps/server). Every call attaches the
// bearer token from lib/session.ts — this app has no server-side data fetching
// to fall back to, so every page that needs data calls through here.

import { getToken } from "@/lib/session";
import type { components } from "@/lib/types.gen";

export type Task = components["schemas"]["Task"];
export type CreateTaskRequest = components["schemas"]["CreateTaskRequest"];
export type UpdateTaskRequest = components["schemas"]["UpdateTaskRequest"];
export type Context = components["schemas"]["Context"];
export type CreateContextRequest = components["schemas"]["CreateContextRequest"];

/** Shared shape used by TaskForm — assignable to both CreateTaskRequest and UpdateTaskRequest. */
export type TaskPayload = {
  title: string;
  description?: string;
  context_id?: string;
  priority?: "low" | "medium" | "high" | "urgent";
  status?: "todo" | "in_progress" | "done" | "archived";
  due_date?: string;
  tags?: string[];
};

const API = process.env.NEXT_PUBLIC_API_URL;

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getToken() ?? ""}`,
      ...init?.headers,
    },
  });

  if (res.status === 204) return undefined as T;

  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new ApiError(res.status, data?.error ?? `Request failed (${res.status})`);
  }
  return data as T;
}

// ── Tasks ──────────────────────────────────────────────────────

export async function listTasks(filters?: { status?: string; context_id?: string }) {
  const params = new URLSearchParams();
  if (filters?.status) params.set("status", filters.status);
  if (filters?.context_id) params.set("context_id", filters.context_id);
  const qs = params.toString();
  // Go serializes an empty (nil) slice as JSON `null`, not `[]`.
  return (await request<Task[] | null>(`/api/v1/tasks${qs ? `?${qs}` : ""}`)) ?? [];
}

export function getTask(id: string) {
  return request<Task>(`/api/v1/tasks/${id}`);
}

export function createTask(body: CreateTaskRequest) {
  return request<Task>("/api/v1/tasks", { method: "POST", body: JSON.stringify(body) });
}

export function updateTask(id: string, body: UpdateTaskRequest) {
  return request<Task>(`/api/v1/tasks/${id}`, { method: "PATCH", body: JSON.stringify(body) });
}

export function deleteTask(id: string) {
  return request<void>(`/api/v1/tasks/${id}`, { method: "DELETE" });
}

// ── Contexts ───────────────────────────────────────────────────

export async function listContexts() {
  return (await request<Context[] | null>("/api/v1/contexts")) ?? [];
}

export function createContext(body: CreateContextRequest) {
  return request<Context>("/api/v1/contexts", { method: "POST", body: JSON.stringify(body) });
}
