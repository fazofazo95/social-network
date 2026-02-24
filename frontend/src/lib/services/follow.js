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

export async function getFollowers() {
  const payload = await apiRequest("/api/users/followers", {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function getFollowing() {
  const payload = await apiRequest("/api/users/following", {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function getPendingRequests() {
  const payload = await apiRequest("/api/users/pending", {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function acceptFollowRequest(userId) {
  const payload = await apiRequest(`/api/users/${userId}/follow/accept`, {
    method: "POST",
  });

  return payload?.data || null;
}

export async function getBlockedUsers() {
  const payload = await apiRequest("/api/users/blocked", {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function blockUser(targetUserId) {
  const payload = await apiRequest(`/api/users/${targetUserId}/block`, {
    method: "POST",
  });

  return payload?.data || null;
}

export async function unblockUser(targetUserId) {
  const payload = await apiRequest(`/api/users/${targetUserId}/unblock`, {
    method: "DELETE",
  });

  return payload?.data || null;
}
