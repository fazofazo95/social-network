"use client";

import { useState } from "react";

const API_BASE_URL = "http://localhost:8080";

function toUiStatus(status) {
  const value = String(status || "").trim().toLowerCase();

  if (value === "accepted" || value === "following") {
    return "Following";
  }
  if (value === "pending") {
    return "Pending";
  }
  if (value === "follow back") {
    return "Follow Back";
  }
  if (value === "follow") {
    return "Follow";
  }

  return "Follow";
}

const Follow_Bottom = ({ status: initialStatus, targetUserId }) => {
  const [status, setStatus] = useState(toUiStatus(initialStatus));
  const [isSubmitting, setIsSubmitting] = useState(false);

  const buttonLabel = status === "Following" ? "Unfollow" : status;

  const handleFollow = async (event) => {
    event.preventDefault();

    if (!targetUserId) {
      console.error("Missing targetUserId for follow action");
      return;
    }

    const shouldFollow = status === "Follow" || status === "Follow Back";
    const url = shouldFollow
      ? `${API_BASE_URL}/api/users/${targetUserId}/follow`
      : `${API_BASE_URL}/api/users/${targetUserId}/unfollow`;
    const method = shouldFollow ? "POST" : "DELETE";

    try {
      setIsSubmitting(true);

      const response = await fetch(url, {
        method,
        credentials: "include",
      });

      const payload = await response.json().catch(() => ({}));

      if (!response.ok) {
        console.error("Follow request failed:", payload?.message || response.statusText);
        return;
      }

      if (shouldFollow) {
        const nextStatus = payload?.data?.status;
        setStatus(toUiStatus(nextStatus));
      } else {
        setStatus("Follow");
      }
    } catch (error) {
      console.error("Error during follow action:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <button
      type="button"
      className="text-pink-500 ml-auto cursor-pointer hover:text-pink-400 disabled:opacity-50"
      onClick={handleFollow}
      disabled={isSubmitting}
    >
      {isSubmitting ? "..." : buttonLabel}
    </button>
  );
};

export default Follow_Bottom;
