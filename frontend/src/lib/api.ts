const API_BASE = "/api";

type Tokens = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
};

function saveTokens(tokens: Tokens) {
  localStorage.setItem("access_token", tokens.access_token);
  localStorage.setItem("refresh_token", tokens.refresh_token);
}

function getAccessToken(): string | null {
  return localStorage.getItem("access_token");
}

function getRefreshToken(): string | null {
  return localStorage.getItem("refresh_token");
}

export function clearTokens() {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
}

export function isLoggedIn(): boolean {
  return !!getAccessToken();
}

async function refreshAccessToken(): Promise<boolean> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;

  const res = await fetch(`${API_BASE}/users/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!res.ok) return false;

  const tokens: Tokens = await res.json();
  saveTokens(tokens);
  return true;
}

async function apiFetch(
  path: string,
  options: RequestInit = {}
): Promise<Response> {
  const token = getAccessToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  let res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (res.status === 401 && token) {
    const refreshed = await refreshAccessToken();
    if (refreshed) {
      headers["Authorization"] = `Bearer ${getAccessToken()}`;
      res = await fetch(`${API_BASE}${path}`, { ...options, headers });
    } else {
      clearTokens();
      window.location.href = "/login";
    }
  }

  return res;
}

// Auth
export async function register(
  email: string,
  password: string,
  userName: string
): Promise<{ id: string }> {
  const res = await fetch(`${API_BASE}/users`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, user_name: userName }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.message || "Registration failed");
  }
  return res.json();
}

export async function login(
  email: string,
  password: string
): Promise<Tokens> {
  const res = await fetch(`${API_BASE}/users/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.message || "Login failed");
  }
  const tokens: Tokens = await res.json();
  saveTokens(tokens);
  return tokens;
}

// Tasks
export type Task = {
  id: string;
  title: string;
  description: string;
  status: string;
  due_date: string;
  version: number;
};

export type TaskListResponse = {
  items: Task[];
  next_cursor: string;
};

export async function listTasks(): Promise<TaskListResponse> {
  const res = await apiFetch("/tasks");
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.message || "Failed to fetch tasks");
  }
  return res.json();
}

export async function createTask(
  title: string,
  description: string,
  dueDate: number
): Promise<{ id: string }> {
  const res = await apiFetch("/tasks", {
    method: "POST",
    body: JSON.stringify({ title, description, due_date: dueDate }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.message || "Failed to create task");
  }
  return res.json();
}

export async function updateTask(
  id: string,
  version: number,
  fields: { title?: string; description?: string; due_date?: number }
): Promise<{ id: string }> {
  const res = await apiFetch(`/tasks/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ version, ...fields }),
  });
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.message || "Failed to update task");
  }
  return res.json();
}

export async function deleteTask(
  id: string,
  version: number
): Promise<void> {
  const res = await apiFetch(`/tasks/${id}`, {
    method: "DELETE",
    body: JSON.stringify({ version }),
  });
  if (res.status !== 204 && !res.ok) {
    const err = await res.json();
    throw new Error(err.message || "Failed to delete task");
  }
}
