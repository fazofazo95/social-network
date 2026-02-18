import { getApiBaseUrl } from "src/lib/apiClient";
import { FALLBACK_PROFILE_IMAGE } from "src/lib/constants/images";

export function parseProfileImage(profilePicture) {
  if (!profilePicture || typeof profilePicture !== "string") {
    return FALLBACK_PROFILE_IMAGE;
  }

  const trimmed = profilePicture.trim();
  if (!trimmed) {
    return FALLBACK_PROFILE_IMAGE;
  }

  if (trimmed.startsWith("http://") || trimmed.startsWith("https://") || trimmed.startsWith("data:")) {
    return trimmed;
  }

  if (trimmed.startsWith("/uploads/")) {
    return `${getApiBaseUrl()}${trimmed}`;
  }

  return FALLBACK_PROFILE_IMAGE;
}
