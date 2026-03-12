import { apiRequest } from "src/lib/apiClient";

export async function searchAll(query, limit = 10) {
  const params = new URLSearchParams({ q: query });
  if (limit !== 10) params.set("limit", String(limit));

  const payload = await apiRequest(`/api/search?${params}`, {
    method: "GET",
  });

  return {
    users: Array.isArray(payload?.data?.users) ? payload.data.users : [],
    groups: Array.isArray(payload?.data?.groups) ? payload.data.groups : [],
  };
}
