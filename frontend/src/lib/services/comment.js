import { apiRequest } from "src/lib/apiClient";

function normalizeMediaUrl(value) {
  if (typeof value !== "string") {
    return "";
  }

  const trimmed = value.trim();
  if (
    trimmed.startsWith("/uploads/") ||
    trimmed.startsWith("http://") ||
    trimmed.startsWith("https://") ||
    trimmed.startsWith("data:")
  ) {
    return trimmed;
  }

  return "";
}

function normalizeComment(comment) {
  if (!comment || typeof comment !== "object") {
    return comment;
  }

  return {
    ...comment,
    image: normalizeMediaUrl(comment.image || comment.extra_content),
  };
}

export async function getPostComments(postId) {
  const payload = await apiRequest(`/api/posts/${postId}/comments`, {
    method: "GET",
  });

  return Array.isArray(payload?.data) ? payload.data.map(normalizeComment) : [];
}

export async function createComment(formData) {
  return apiRequest("/api/comments", {
    method: "POST",
    body: formData,
  });
}

export async function updateComment(commentId, content) {
  return apiRequest(`/api/comments/${commentId}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ content }),
  });
}

export async function deleteComment(commentId) {
  return apiRequest(`/api/comments/${commentId}/delete`, {
    method: "PUT",
  });
}

export async function restoreComment(commentId) {
  return apiRequest(`/api/comments/${commentId}/restore`, {
    method: "PUT",
  });
}