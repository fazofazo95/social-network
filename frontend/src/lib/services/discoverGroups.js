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
    description: group.description,
    group_picture: group.group_picture || null,
    members_count: group.group_members,
    owner_first_name: group.owner_first_name || "",
    owner_last_name: group.owner_last_name || "",
    type: group.type,
  }));
}
