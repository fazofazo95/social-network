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

function normalizePost(post) {
  if (!post || typeof post !== "object") {
    return post;
  }

  return {
    ...post,
    image: normalizeMediaUrl(post.image || post.extra_content),
  };
}

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

  return Array.isArray(payload?.data) ? payload.data.map(normalizePost) : [];
}

export async function getFeedPosts(page = 1) {
  const payload = await apiRequest(`/api/feed?page=${page}`, {
    method: "GET",
  });

  const posts = Array.isArray(payload?.data?.posts) ? payload.data.posts : [];
  return posts.map(normalizePost);
}