"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import IconButton from "./IconButton";
import SearchBar from "./SearchBar";
import NotificationDropdown from "./NotificationDropdown";
import { logoutUser } from "src/lib/services/auth";
import { listChats } from "src/lib/services/chat";
import { searchAll } from "src/lib/services/search";
import { getApiBaseUrl } from "src/lib/apiClient";

function toWebSocketUrl(path = "/ws") {
  const baseUrl = getApiBaseUrl();
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  if (baseUrl.startsWith("https://")) {
    return `${baseUrl.replace("https://", "wss://")}${normalizedPath}`;
  }
  return `${baseUrl.replace("http://", "ws://")}${normalizedPath}`;
}

const NavBar = () => {
  const router = useRouter();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const [unreadMessagesCount, setUnreadMessagesCount] = useState(0);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchUsers, setSearchUsers] = useState([]);
  const [searchGroups, setSearchGroups] = useState([]);
  const [isSearchLoading, setIsSearchLoading] = useState(false);
  const menuRef = useRef(null);
  const wsRef = useRef(null);
  const searchTimerRef = useRef(null);

  useEffect(() => {
    const handleOutsideClick = (event) => {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setIsMenuOpen(false);
      }
    };

    document.addEventListener("mousedown", handleOutsideClick);
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
    };
  }, []);

  // Handle search input changes with debounce
  const handleSearchChange = (e) => {
    const query = e.target.value;
    setSearchQuery(query);

    if (searchTimerRef.current) clearTimeout(searchTimerRef.current);

    if (!query.trim()) {
      setSearchUsers([]);
      setSearchGroups([]);
      setIsSearchLoading(false);
      return;
    }

    setIsSearchLoading(true);
    searchTimerRef.current = setTimeout(async () => {
      try {
        const { users, groups } = await searchAll(query.trim());
        setSearchUsers(users);
        setSearchGroups(groups);
      } catch (err) {
        console.error("Search failed:", err);
        setSearchUsers([]);
        setSearchGroups([]);
      } finally {
        setIsSearchLoading(false);
      }
    }, 300);
  };

  useEffect(() => {
    return () => {
      if (searchTimerRef.current) clearTimeout(searchTimerRef.current);
    };
  }, []);

  const handleSearchResultClick = () => {
    setSearchQuery("");
    setSearchUsers([]);
    setSearchGroups([]);
  };

  useEffect(() => {
    let disposed = false;
    let reconnectTimeout = null;

    async function refreshUnreadCount() {
      try {
        const chats = await listChats();
        if (disposed) return;
        const unread = (Array.isArray(chats) ? chats : []).filter((chatItem) => !chatItem?.seen).length;
        setUnreadMessagesCount(unread);
      } catch {
        if (!disposed) {
          setUnreadMessagesCount(0);
        }
      }
    }

    const connect = () => {
      if (disposed) return;

      try {
        const socket = new WebSocket(toWebSocketUrl("/ws"));
        wsRef.current = socket;

        socket.onopen = () => {
          refreshUnreadCount();
        };

        socket.onmessage = (event) => {
          let incoming = null;
          try {
            incoming = JSON.parse(event.data);
          } catch {
            refreshUnreadCount();
            return;
          }

          const incomingChatId = Number(incoming?.chat_id || 0);
          const incomingMessageId = Number(incoming?.id || 0);
          if (!incomingChatId || !incomingMessageId) {
            return;
          }

          refreshUnreadCount();
        };

        socket.onerror = () => {
          if (!disposed) socket.close();
        };

        socket.onclose = () => {
          if (disposed) return;
          reconnectTimeout = setTimeout(connect, 2000);
        };
      } catch {
        reconnectTimeout = setTimeout(connect, 2000);
      }
    };

    const handleUnreadRefresh = () => {
      refreshUnreadCount();
    };

    const handleVisibilityChange = () => {
      if (!document.hidden) {
        refreshUnreadCount();
      }
    };

    refreshUnreadCount();
    const poll = setInterval(refreshUnreadCount, 30000);
    window.addEventListener("messages:unread-refresh", handleUnreadRefresh);
    window.addEventListener("focus", handleUnreadRefresh);
    document.addEventListener("visibilitychange", handleVisibilityChange);
    connect();

    return () => {
      disposed = true;
      clearInterval(poll);
      window.removeEventListener("messages:unread-refresh", handleUnreadRefresh);
      window.removeEventListener("focus", handleUnreadRefresh);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      if (wsRef.current) {
        try {
          wsRef.current.onclose = null;
          wsRef.current.close();
        } catch {
          // no-op
        }
        wsRef.current = null;
      }
    };
  }, []);

  const handleLogout = async () => {
    try {
      setIsLoggingOut(true);
      await logoutUser();
    } catch (error) {
      console.error("Logout failed:", error?.message || error);
    } finally {
      setIsLoggingOut(false);
      setIsMenuOpen(false);
      router.replace("/login");
    }
  };

  return (
    <nav className="border-b border-purple-500/20 w-full h-10 bg-[#1a1a2e] relative flex flex-row items-center justify-between shadow-custom px-4">
      <div className="flex items-center gap-3 relative">
        <Link href="/" className="flex items-center gap-1.5 shrink-0">
          <Image src="/logo_icon.svg" alt="Logo" width={22} height={22} />
          <div className="relative">
            <span className="text-purple-500 text-lg font-semibold relative">
              Pulse
            </span>
            <span className="absolute top-0 left-0 text-lg text-purple-500 neon-glow">
              Pulse
            </span>
          </div>
        </Link>
        
        <SearchBar
          placeholder="Search..."
          value={searchQuery}
          onChange={handleSearchChange}
          users={searchUsers}
          groups={searchGroups}
          onResultClick={handleSearchResultClick}
          isLoading={isSearchLoading}
        />
      </div>

      <div className="flex items-center gap-3 pr-4">
        <NotificationDropdown />

        <Link href="/messages" className="relative inline-flex items-center justify-center w-6 h-6 rounded-full bg-transparent hover:bg-purple-900/30 transition" aria-label="Messages">
          <span className="text-sm">💬</span>
          {unreadMessagesCount > 0 ? (
            <span className="absolute -top-1 -right-1 min-w-4 h-4 px-1 rounded-full bg-red-500 text-white text-[10px] leading-4 text-center">
              {unreadMessagesCount > 99 ? "99+" : unreadMessagesCount}
            </span>
          ) : null}
        </Link>

        <div className="relative" ref={menuRef}>
          <IconButton
            emoji="👤"
            alt="Profile Icon"
            iconSize={20}
            onClick={() => setIsMenuOpen((prev) => !prev)}
          />

          {isMenuOpen && (
            <div className="absolute right-0 mt-2 w-36 rounded-lg border border-purple-500/30 bg-[#1a1a2e] shadow-custom z-50">
              <Link
                href="/profile"
                className="block px-3 py-2 text-sm text-purple-100 hover:bg-purple-900/30 rounded-t-lg transition"
                onClick={() => setIsMenuOpen(false)}
              >
                Profile
              </Link>
              <button
                type="button"
                className="w-full text-left px-3 py-2 text-sm text-purple-100 hover:bg-purple-900/30 rounded-b-lg disabled:opacity-50 transition cursor-pointer"
                onClick={handleLogout}
                disabled={isLoggingOut}
              >
                {isLoggingOut ? "Logging out..." : "Logout"}
              </button>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
};

export default NavBar;
