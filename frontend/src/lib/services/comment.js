import { apiRequest } from "src/lib/apiClient";

export async function getPostComments(postId) {
  const payload = await apiRequest(`/api/posts/${postId}/comments`, {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data : [];
}

export async function createComment(formData) {
  return apiRequest("/api/comments", {
    method: "POST",
    body: formData,
  });
}