export function formatFriendlyDateTime(value) {
  if (!value) {
    return "";
  }

  // Backend sends UTC timestamps without timezone marker (e.g. "2026-03-01 10:00:00").
  // Append "Z" so JS Date parses them as UTC, making relative-time diffs correct.
  let raw = value;
  if (typeof raw === "string" && !raw.endsWith("Z") && !raw.includes("+") && !raw.includes("T")) {
    raw = raw.replace(" ", "T") + "Z";
  }

  const date = raw instanceof Date ? raw : new Date(raw);
  if (Number.isNaN(date.getTime())) {
    return "";
  }

  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const minuteMs = 60 * 1000;
  const hourMs = 60 * minuteMs;
  const dayMs = 24 * hourMs;

  if (diffMs >= 0 && diffMs < minuteMs) {
    return "Just now";
  }

  if (diffMs >= minuteMs && diffMs < hourMs) {
    const minutes = Math.floor(diffMs / minuteMs);
    return `${minutes} min ago`;
  }

  if (diffMs >= hourMs && diffMs < dayMs) {
    const hours = Math.floor(diffMs / hourMs);
    return `${hours}h ago`;
  }

  if (diffMs >= dayMs && diffMs < 7 * dayMs) {
    const days = Math.floor(diffMs / dayMs);
    return `${days}d ago`;
  }

  return new Intl.DateTimeFormat(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}