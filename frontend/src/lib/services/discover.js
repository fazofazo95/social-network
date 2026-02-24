import { apiRequest } from "src/lib/apiClient";

export async function getDiscoveredUsers() {
  const payload = await apiRequest("/api/discover", {
    method: "GET",
  });

  if (!Array.isArray(payload?.data)) {
    return [];
  }

  return payload.data
    .map((user) => {
      if (!user || typeof user !== "object") {
        return null;
      }

      const id = user.id ?? user.user_id ?? null;
      return {
        ...user,
        id,
      };
    })
    .filter((user) => user && user.id !== null && user.id !== undefined);
}
