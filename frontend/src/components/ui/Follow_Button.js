"use client";

import { useState } from "react";
import { followUser, unfollowUser } from "src/lib/services/follow";

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

    try {
      setIsSubmitting(true);

      if (shouldFollow) {
        const data = await followUser(targetUserId);
        const nextStatus = data?.status;
        setStatus(toUiStatus(nextStatus));
      } else {
        await unfollowUser(targetUserId);
        setStatus("Follow");
      }
    } catch (error) {
      console.error("Follow request failed:", error?.message || error);
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
