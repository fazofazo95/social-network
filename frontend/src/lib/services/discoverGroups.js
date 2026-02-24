import { apiRequest } from "src/lib/apiClient";

export async function getDiscoveredGroups(page = 1) {
  const payload = await apiRequest(`/api/groups/discover?page=${page}`, {
    method: "GET",
  });

  if (!payload?.data?.items || !Array.isArray(payload.data.items)) {
    return [];
  }

  return payload.data.items.map((group) => ({
    id: group.id,
    name: group.name,
    group_picture: group.group_picture || "/groups_icon.svg",
    type: group.type,
  }));
}
