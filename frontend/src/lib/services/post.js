import { apiRequest } from "src/lib/apiClient";

export async function createPost(formData) {
  return apiRequest("/api/posts", {
    method: "POST",
    body: formData,
  });
}

export async function getUserPosts(userId, page = 1, limit = 10) {
  const payload = await apiRequest(`/api/users/${userId}/posts?page=${page}&limit=${limit}`, {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}