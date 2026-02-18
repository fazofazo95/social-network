import { apiRequest } from "src/lib/apiClient";

export async function getDiscoveredUsers() {
  const payload = await apiRequest("/api/discover", {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}
