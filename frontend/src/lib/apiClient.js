const API_BASE_URL = (process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080").replace(/\/$/, "");

export function getApiBaseUrl() {
  return API_BASE_URL;
}

export function toApiUrl(path) {
  if (!path) return API_BASE_URL;
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  return `${API_BASE_URL}${path.startsWith("/") ? "" : "/"}${path}`;
}

export async function apiRequest(path, options = {}) {
  const response = await fetch(toApiUrl(path), {
    credentials: "include",
    ...options,
  });

  const payload = await response.json().catch(() => ({}));

  if (!response.ok) {
    const message = payload?.message || response.statusText || "Request failed";
    const error = new Error(message);
    error.status = response.status;
    error.payload = payload;
    throw error;
  }

  return payload;
}
