export function parseProfileImage(profilePicture) {
  if (!profilePicture || typeof profilePicture !== "string") {
    return null;
  }

  const trimmed = profilePicture.trim();
  if (!trimmed) {
    return null;
  }

  if (
    trimmed.startsWith("http://") ||
    trimmed.startsWith("https://") ||
    trimmed.startsWith("data:")
  ) {
    return trimmed;
  }

  if (trimmed.startsWith("/uploads/")) {
    return trimmed;
  }

  return null;
}
