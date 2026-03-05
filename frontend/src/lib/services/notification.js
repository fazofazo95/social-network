import { apiRequest } from "src/lib/apiClient";

export async function listNotifications(limit = 20, offset = 0) {
  const payload = await apiRequest(
    `/api/notifications?limit=${limit}&offset=${offset}`,
    { method: "GET" }
  );
  return payload?.data || { items: [], limit, offset, has_more: false };
}

export async function markNotificationRead(id) {
  return apiRequest(`/api/notifications/${id}/read`, { method: "POST" });
}

export async function markAllNotificationsRead() {
  return apiRequest(`/api/notifications/read-all`, { method: "POST" });
}

export async function respondToNotification(id, action) {
  return apiRequest(`/api/notifications/${id}/action`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ action }),
  });
}
