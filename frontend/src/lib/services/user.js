import { apiRequest } from "src/lib/apiClient";

export async function fetchUserData(id) {
  const payload = await apiRequest(`/api/users/${id}`, {
    method: "GET",
  });
  return payload.data;
}

export async function fetchVisibilitySettings() {
  const payload = await apiRequest("/api/users/settings", {
    method: "GET",
  });
  return payload?.data || null;
}

export async function updateVisibilitySettings(data) {
  const payload = await apiRequest("/api/users/settings", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
  return payload?.data || null;
}

export async function fetchContentSettings() {
  const payload = await apiRequest("/api/users/settings/content", {
    method: "GET",
  });
  return payload?.data || null;
}

export async function updateContentSettings(data) {
  const payload = await apiRequest("/api/users/settings/content", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
  return payload?.data || null;
}

export async function updateUserAvatar(formData) {
  const payload = await apiRequest("/api/users/me", {
    method: "PUT",
    body: formData,
  });
  return payload?.data || null;
}

export async function updateUserCover(formData) {
  const payload = await apiRequest("/api/users/me", {
    method: "PUT",
    body: formData,
  });
  return payload?.data || null;
}
