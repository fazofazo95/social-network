import { getApiBaseUrl } from "src/lib/apiClient";

export function toUploadUrl(path, fallback = "") {
  if (!path) return fallback;
  if (path.startsWith("http://") || path.startsWith("https://") || path.startsWith("data:")) {
    return path;
  }
  if (path.startsWith("/uploads/")) {
    return `${getApiBaseUrl()}${path}`;
  }
  return fallback;
}

export function toCoverUrl(path, fallback = "/example_cover.png") {
  return toUploadUrl(path, fallback);
}
