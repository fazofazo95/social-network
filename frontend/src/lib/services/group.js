import { apiRequest } from "src/lib/apiClient";
import { getDiscoveredUsers } from "src/lib/services/discover";

export async function createGroup(formData) {
  return apiRequest("/api/groups", {
    method: "POST",
    body: formData,
  });
}

export async function getActiveGroups() {
  const payload = await apiRequest("/api/groups/active", {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function getPendingGroupInvites() {
  const payload = await apiRequest("/api/groups/invites/pending", {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function getPendingGroupRequests() {
  const payload = await apiRequest("/api/groups/requests/pending", {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function getGroupPage(groupId) {
  const payload = await apiRequest(`/api/groups/${groupId}`, {
    method: "GET",
  });

  return payload?.data || null;
}

export async function requestToJoinGroup(groupId) {
  return apiRequest(`/api/groups/${groupId}/join`, {
    method: "POST",
  });
}

export async function acceptGroupInvite(groupId) {
  return apiRequest(`/api/groups/${groupId}/invite/accept`, {
    method: "POST",
  });
}

export async function rejectGroupInvite(groupId) {
  return apiRequest(`/api/groups/${groupId}/invite/reject`, {
    method: "POST",
  });
}

export async function removeOwnGroupRequest(groupId) {
  return apiRequest(`/api/groups/${groupId}/requests/me`, {
    method: "DELETE",
  });
}

export async function leaveGroup(groupId) {
  return apiRequest(`/api/groups/${groupId}/leave`, {
    method: "POST",
  });
}

// ===== MODERATOR/OWNER ENDPOINTS =====

export async function getGroupMembers(groupId) {
  const payload = await apiRequest(`/api/groups/${groupId}/members`, {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function getGroupPendingRequests(groupId) {
  const payload = await apiRequest(`/api/groups/${groupId}/requests/pending`, {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function getGroupPendingInvites(groupId) {
  const payload = await apiRequest(`/api/groups/${groupId}/invites/pending`, {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function acceptJoinRequest(groupId, userId) {
  return apiRequest(`/api/groups/${groupId}/requests/${userId}/accept`, {
    method: "POST",
  });
}

export async function rejectJoinRequest(groupId, userId) {
  return apiRequest(`/api/groups/${groupId}/requests/${userId}/reject`, {
    method: "POST",
  });
}

export async function removeInvite(groupId, userId) {
  return apiRequest(`/api/groups/${groupId}/invites/${userId}`, {
    method: "DELETE",
  });
}

export async function kickMember(groupId, userId) {
  return apiRequest(`/api/groups/${groupId}/members/${userId}/kick`, {
    method: "POST",
  });
}

export async function promoteMember(groupId, userId) {
  return apiRequest(`/api/groups/${groupId}/members/${userId}/promote`, {
    method: "POST",
  });
}

export async function demoteMember(groupId, userId) {
  return apiRequest(`/api/groups/${groupId}/members/${userId}/demote`, {
    method: "POST",
  });
}

export async function deleteGroup(groupId) {
  return apiRequest(`/api/groups/${groupId}`, {
    method: "DELETE",
  });
}

export async function inviteToGroup(groupId, userId) {
  return apiRequest(`/api/groups/${groupId}/invite/${userId}`, {
    method: "POST",
  });
}

// ===== GROUP SETTINGS (OWNER ONLY) =====

export async function getGroupSettings(groupId) {
  const payload = await apiRequest(`/api/groups/${groupId}/settings`, {
    method: "GET",
  });
  return payload?.data || null;
}

export async function updateGroupSettings(groupId, data) {
  const payload = await apiRequest(`/api/groups/${groupId}/settings`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  return payload?.data || null;
}

// ===== GROUP POSTS =====

export async function createGroupPost(groupId, formData) {
  return apiRequest(`/api/groups/${groupId}/posts`, {
    method: "POST",
    body: formData,
  });
}

export async function getGroupPosts(groupId, page = 1) {
  const payload = await apiRequest(`/api/groups/${groupId}/posts?page=${page}`, {
    method: "GET",
  });

  return {
    page: payload?.data?.page ?? page,
    posts: Array.isArray(payload?.data?.posts) ? payload.data.posts : [],
  };
}

export async function deleteGroupPost(groupId, postId) {
  return apiRequest(`/api/groups/${groupId}/posts/${postId}`, {
    method: "DELETE",
  });
}
