import { apiRequest } from "src/lib/apiClient";

export async function addReaction(postId) {
  return apiRequest(`/api/posts/${postId}/reactions`, {
    method: "POST",
  });
}

export async function removeReaction(postId) {
  return apiRequest(`/api/posts/${postId}/reactions`, {
    method: "DELETE",
  });
}
