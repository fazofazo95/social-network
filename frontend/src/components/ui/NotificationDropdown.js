"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import Image from "next/image";
import {
  listNotifications,
  markAllNotificationsRead,
  respondToNotification,
} from "src/lib/services/notification";
import { parseProfileImage } from "src/lib/utils/profileImage";
import { formatFriendlyDateTime } from "src/lib/utils/dateTime";
import { getApiBaseUrl } from "src/lib/apiClient";

function toSSEUrl(path) {
  return `${getApiBaseUrl()}${path}`;
}

const TYPE_LABELS = {
  follow_request: "Follow Request",
  group_invite: "Group Invite",
  group_join_request: "Join Request",
  group_event_created: "New Event",
};

export default function NotificationDropdown() {
  const [open, setOpen] = useState(false);
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState({});
  const dropdownRef = useRef(null);
  const sseRef = useRef(null);

  // Fetch notifications
  const fetchNotifications = useCallback(async () => {
    try {
      setLoading(true);
      const data = await listNotifications(50, 0);
      const items = Array.isArray(data.items) ? data.items : [];
      setNotifications(items);
      setUnreadCount(items.filter((n) => !n.seen).length);
    } catch (err) {
      console.error("Failed to fetch notifications:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Initial load
  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  // SSE for real-time notifications
  useEffect(() => {
    let disposed = false;
    let retryTimeout = null;

    function connect() {
      if (disposed) return;
      try {
        const es = new EventSource(toSSEUrl("/api/notifications/stream"), {
          withCredentials: true,
        });
        sseRef.current = es;

        es.addEventListener("notification:new", (e) => {
          try {
            const payload = JSON.parse(e.data);
            if (payload?.notification) {
              setNotifications((prev) => {
                const exists = prev.some((n) => n.id === payload.notification.id);
                if (exists) return prev;
                return [payload.notification, ...prev];
              });
              setUnreadCount((c) => c + 1);
            }
          } catch {
            // ignore parse errors
          }
        });

        es.onerror = () => {
          es.close();
          if (!disposed) {
            retryTimeout = setTimeout(connect, 5000);
          }
        };
      } catch {
        if (!disposed) {
          retryTimeout = setTimeout(connect, 5000);
        }
      }
    }

    connect();

    return () => {
      disposed = true;
      if (retryTimeout) clearTimeout(retryTimeout);
      if (sseRef.current) {
        sseRef.current.close();
        sseRef.current = null;
      }
    };
  }, []);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClick(e) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  // When dropdown opens, mark all as read
  const handleToggle = async () => {
    const willOpen = !open;
    setOpen(willOpen);

    if (willOpen && unreadCount > 0) {
      try {
        await markAllNotificationsRead();
        setNotifications((prev) => prev.map((n) => ({ ...n, seen: true })));
        setUnreadCount(0);
      } catch (err) {
        console.error("Failed to mark notifications read:", err);
      }
    }
  };

  // Handle accept / reject
  const handleAction = async (notifId, action) => {
    setActionLoading((prev) => ({ ...prev, [notifId]: action }));
    try {
      await respondToNotification(notifId, action);
      setNotifications((prev) =>
        prev.map((n) =>
          n.id === notifId
            ? { ...n, status: action === "accept" ? "accepted" : "rejected" }
            : n
        )
      );
    } catch (err) {
      console.error(`Failed to ${action} notification:`, err);
    } finally {
      setActionLoading((prev) => {
        const next = { ...prev };
        delete next[notifId];
        return next;
      });
    }
  };

  const isActionable = (n) =>
    (n.status === "pending" || n.status === "read") &&
    ["follow_request", "group_invite", "group_join_request"].includes(n.type);

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Bell button */}
      <button
        type="button"
        onClick={handleToggle}
        className="relative w-6 h-6 flex items-center justify-center rounded-full bg-transparent hover:bg-purple-900/30 transition cursor-pointer"
        aria-label="Notifications"
      >
        <Image
          src="/notif-icon.svg"
          alt="Notification Icon"
          width={17}
          height={17}
        />
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 min-w-4 h-4 px-1 rounded-full bg-red-500 text-white text-[10px] leading-4 text-center shadow-[0_0_8px_rgba(239,68,68,0.6)]">
            {unreadCount > 99 ? "99+" : unreadCount}
          </span>
        )}
      </button>

      {/* Dropdown */}
      {open && (
        <div className="absolute right-0 mt-2 w-80 max-h-96 overflow-y-auto rounded-lg border border-purple-500/30 bg-[#1a1a2e] shadow-custom z-50">
          <div className="sticky top-0 bg-[#1a1a2e] border-b border-purple-500/20 px-4 py-2 flex items-center justify-between">
            <h3 className="text-sm font-semibold text-purple-100">Notifications</h3>
          </div>

          {loading && notifications.length === 0 ? (
            <div className="px-4 py-6 text-center text-sm text-purple-400">
              Loading...
            </div>
          ) : notifications.length === 0 ? (
            <div className="px-4 py-6 text-center text-sm text-purple-400">
              No notifications
            </div>
          ) : (
            <ul className="divide-y divide-purple-500/20">
              {notifications.map((n) => (
                <li
                  key={n.id}
                  className={`px-4 py-3 flex gap-3 items-start hover:bg-purple-900/20 transition ${
                    !n.seen ? "bg-purple-900/30" : ""
                  }`}
                >
                  {/* Actor avatar */}
                  <div className="shrink-0 w-8 h-8 rounded-full overflow-hidden bg-[#0d0d1a] relative">
                    <Image
                      src={parseProfileImage(n.actor_picture)}
                      alt={`${n.actor_first_name || "User"}'s avatar`}
                      fill
                      className="object-cover"
                    />
                  </div>

                  <div className="flex-1 min-w-0">
                    {/* Name + content */}
                    <p className="text-sm text-purple-100 leading-snug">
                      <span className="font-semibold">
                        {n.actor_first_name || ""} {n.actor_last_name || ""}
                      </span>{" "}
                      <span className="text-purple-300">{n.content}</span>
                    </p>

                    {/* Type badge + time */}
                    <div className="flex items-center gap-2 mt-1">
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-purple-900/50 text-purple-300 font-medium border border-purple-500/20">
                        {TYPE_LABELS[n.type] || n.type}
                      </span>
                      <span className="text-[11px] text-purple-400/60">
                        {formatFriendlyDateTime(n.created_at)}
                      </span>
                    </div>

                    {/* Status badge for resolved */}
                    {n.status === "accepted" && (
                      <span className="inline-block mt-1 text-[10px] px-1.5 py-0.5 rounded bg-green-900/30 text-green-400 font-medium border border-green-500/20">
                        Accepted
                      </span>
                    )}
                    {n.status === "rejected" && (
                      <span className="inline-block mt-1 text-[10px] px-1.5 py-0.5 rounded bg-red-900/30 text-red-400 font-medium border border-red-500/20">
                        Rejected
                      </span>
                    )}

                    {/* Action buttons */}
                    {isActionable(n) && (
                      <div className="flex gap-2 mt-2">
                        <button
                          type="button"
                          disabled={!!actionLoading[n.id]}
                          onClick={() => handleAction(n.id, "accept")}
                          className="px-3 py-1 text-xs rounded-md bg-purple-600 text-white hover:bg-purple-500 disabled:opacity-50 transition cursor-pointer shadow-[0_0_8px_rgba(168,85,247,0.3)]"
                        >
                          {actionLoading[n.id] === "accept" ? "..." : "Accept"}
                        </button>
                        <button
                          type="button"
                          disabled={!!actionLoading[n.id]}
                          onClick={() => handleAction(n.id, "reject")}
                          className="px-3 py-1 text-xs rounded-md bg-purple-900/30 text-purple-300 border border-purple-500/30 hover:bg-purple-900/50 disabled:opacity-50 transition cursor-pointer"
                        >
                          {actionLoading[n.id] === "reject" ? "..." : "Reject"}
                        </button>
                      </div>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
