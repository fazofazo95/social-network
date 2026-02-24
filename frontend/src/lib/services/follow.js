import { apiRequest } from "src/lib/apiClient";

export async function followUser(targetUserId) {
  const payload = await apiRequest(`/api/users/${targetUserId}/follow`, {
    method: "POST",
  });

  return payload?.data || null;
}

export async function unfollowUser(targetUserId) {
  const payload = await apiRequest(`/api/users/${targetUserId}/unfollow`, {
    method: "DELETE",
  });

  return payload?.data || null;
}

export async function getFollowersByUser(userId) {
  const payload = await apiRequest(`/api/users/${userId}/followers`, {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function getFollowingByUser(userId) {
  const payload = await apiRequest(`/api/users/${userId}/following`, {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}
