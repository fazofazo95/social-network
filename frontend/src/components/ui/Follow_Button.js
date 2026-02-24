"use client";

import { useEffect, useState } from "react";
import { followUser, unfollowUser } from "src/lib/services/follow";

function toUiStatus(status) {
  const value = String(status || "").trim().toLowerCase();

  if (value === "accepted" || value === "following") {
    return "Following";
  }
  if (value === "unfollow") {
    return "Unfollow";
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

const Follow_Bottom = ({ status: initialStatus, targetUserId, onStatusChange }) => {
  const [status, setStatus] = useState(toUiStatus(initialStatus));
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    setStatus(toUiStatus(initialStatus));
  }, [initialStatus]);

  const buttonLabel = status === "Following" ? "Unfollow" : status;

  const handleFollow = async (event) => {
    event.preventDefault();

    if (!targetUserId) {
      console.error("Missing targetUserId for follow action");
      return;
    }

    const shouldFollow = status === "Follow" || status === "Follow Back";
    const shouldUnfollow = status === "Following" || status === "Pending" || status === "Unfollow";

    try {
      setIsSubmitting(true);

      if (shouldFollow) {
        const data = await followUser(targetUserId);
        const nextStatus = data?.status;
        const nextUiStatus = toUiStatus(nextStatus);
        setStatus(nextUiStatus);
        if (typeof onStatusChange === "function") {
          onStatusChange(nextUiStatus);
        }
      } else if (shouldUnfollow) {
        await unfollowUser(targetUserId);
        setStatus("Follow");
        if (typeof onStatusChange === "function") {
          onStatusChange("Follow");
        }
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
