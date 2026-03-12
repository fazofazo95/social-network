"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import { fetchUserData } from "src/lib/services/user";
import { getFollowing } from "src/lib/services/follow";
import { getChatMessages, listChats, markChatRead, sendDirectMessage, sendGroupMessage } from "src/lib/services/chat";
import { getApiBaseUrl } from "src/lib/apiClient";
import { parseProfileImage } from "src/lib/utils/profileImage";

const DIRECT_SEND_RULE_MESSAGE = "You can send a direct message only if you follow this user or their account is public.";
const DIRECT_VIEW_RULE_MESSAGE = "Your account is private. You can receive/view direct messages only from users you follow back.";

function isPublicProfile(profileType) {
  return String(profileType || "public").toLowerCase() === "public";
}

function formatChatTitle(chat) {
  if (!chat) return "Messages";
  if (chat.type === "direct") {
    const first = chat.other_user_first_name || "";
    const last = chat.other_user_last_name || "";
    return `${first} ${last}`.trim() || "Direct Chat";
  }
  return chat.group_id ? `Group #${chat.group_id}` : "Group Chat";
}

function formatSmallTime(value) {
  const parsedDate = parseChatDate(value);
  if (Number.isNaN(parsedDate.getTime())) return "";
  return parsedDate.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

function formatListTime(value, nowTs = Date.now()) {
  const parsedDate = parseChatDate(value);
  if (Number.isNaN(parsedDate.getTime())) return "";

  const diffMs = nowTs - parsedDate.getTime();
  const hourMs = 60 * 60 * 1000;

  if (diffMs < hourMs) {
    const minutes = Math.max(1, Math.floor(diffMs / (60 * 1000)));
    return `${minutes}m ago`;
  }

  if (diffMs < 24 * hourMs) {
    const hours = Math.floor(diffMs / hourMs);
    return `${hours}h ago`;
  }

  return parsedDate.toLocaleDateString();
}

function parseChatDate(value) {
  if (!value) {
    return new Date("");
  }

  if (value instanceof Date) {
    return value;
  }

  const raw = String(value).trim();
  if (!raw) {
    return new Date("");
  }

  const hasTimezone = /([zZ]|[+-]\d{2}:?\d{2})$/.test(raw);
  if (hasTimezone) {
    return new Date(raw);
  }

  const normalized = raw.replace(" ", "T");
  if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2}(\.\d{1,3})?)?$/.test(normalized)) {
    return new Date(`${normalized}Z`);
  }

  return new Date(raw);
}

function normalizeUserItem(user) {
  if (!user || typeof user !== "object") {
    return null;
  }

  const id = user.id ?? user.user_id ?? null;
  if (!id) {
    return null;
  }

  const firstName = user.first_name || user.firstname || "";
  const lastName = user.last_name || user.lastname || "";
  const displayName = `${firstName} ${lastName}`.trim() || user.nickname || user.username || "Unknown User";

  return {
    id,
    first_name: firstName,
    last_name: lastName,
    display_name: displayName,
    username: user.username || user.nickname || "",
    profile_picture: user.profile_picture || "",
    profile_type: user.profile_type || "",
  };
}

function toWebSocketUrl(path = "/ws") {
  const baseUrl = getApiBaseUrl();
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  if (baseUrl.startsWith("https://")) {
    return `${baseUrl.replace("https://", "wss://")}${normalizedPath}`;
  }
  return `${baseUrl.replace("http://", "ws://")}${normalizedPath}`;
}

const MessagesPage = () => {
  const [currentUser, setCurrentUser] = useState(null);
  const [chats, setChats] = useState([]);
  const [selectedChatId, setSelectedChatId] = useState(null);
  const [messages, setMessages] = useState([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [newChatQuery, setNewChatQuery] = useState("");
  const [newChatCandidates, setNewChatCandidates] = useState([]);
  const [newChatTarget, setNewChatTarget] = useState(null);
  const [followingIds, setFollowingIds] = useState([]);
  const [isOwnProfilePublic, setIsOwnProfilePublic] = useState(false);
  const [profilePublicByUserId, setProfilePublicByUserId] = useState({});
  const [draftMessage, setDraftMessage] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isMessagesLoading, setIsMessagesLoading] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [error, setError] = useState("");
  const [accessNotice, setAccessNotice] = useState("");
  const [isSocketConnected, setIsSocketConnected] = useState(false);
  const [onlineUserIds, setOnlineUserIds] = useState([]);
  const [timeTick, setTimeTick] = useState(Date.now());
  const socketRef = useRef(null);
  const selectedChatIdRef = useRef(null);

  const emitUnreadRefresh = () => {
    if (typeof window !== "undefined") {
      window.dispatchEvent(new Event("messages:unread-refresh"));
    }
  };

  const selectedChat = useMemo(
    () => chats.find((chatItem) => chatItem.chat_id === selectedChatId) || null,
    [chats, selectedChatId]
  );

  const headerActivityText = useMemo(() => {
    if (newChatTarget) {
      return "Start new chat";
    }
    if (!selectedChat) {
      return "No chat selected";
    }
    const lastActivity = formatListTime(selectedChat.last_message_at, timeTick);
    return lastActivity ? `Last message ${lastActivity}` : "No messages yet";
  }, [newChatTarget, selectedChat, timeTick]);

  const filteredChats = useMemo(() => {
    const query = searchTerm.trim().toLowerCase();
    if (!query) {
      return chats;
    }

    return chats.filter((chatItem) => {
      const title = formatChatTitle(chatItem).toLowerCase();
      const preview = String(chatItem.last_message_preview || "").toLowerCase();
      return title.includes(query) || preview.includes(query);
    });
  }, [chats, searchTerm]);

  const filteredNewChatCandidates = useMemo(() => {
    const query = newChatQuery.trim().toLowerCase();
    if (!query) {
      return newChatCandidates.slice(0, 6);
    }

    return newChatCandidates
      .filter((item) => {
        const name = String(item.display_name || "").toLowerCase();
        const username = String(item.username || "").toLowerCase();
        return name.includes(query) || username.includes(query);
      })
      .slice(0, 6);
  }, [newChatCandidates, newChatQuery]);

  const followingIdSet = useMemo(
    () => new Set((followingIds || []).map((id) => Number(id))),
    [followingIds]
  );

  const activeDirectTargetId = useMemo(() => {
    if (selectedChat?.type === "direct" && selectedChat?.other_user_id) {
      return Number(selectedChat.other_user_id);
    }
    if (newChatTarget?.id) {
      return Number(newChatTarget.id);
    }
    return 0;
  }, [newChatTarget, selectedChat]);

  const onlineUserIdSet = useMemo(
    () => new Set((onlineUserIds || []).map((id) => Number(id))),
    [onlineUserIds]
  );

  const isDirectTargetOnline = useMemo(() => {
    if (!activeDirectTargetId) {
      return false;
    }
    return onlineUserIdSet.has(Number(activeDirectTargetId));
  }, [activeDirectTargetId, onlineUserIdSet]);

  const canSendInCurrentContext = useMemo(() => {
    if (selectedChat?.type === "group") {
      return true;
    }

    if (!activeDirectTargetId) {
      return false;
    }

    if (followingIdSet.has(activeDirectTargetId) || isOwnProfilePublic) {
      return true;
    }

    if (newChatTarget?.id && newChatTarget?.profile_type) {
      return isPublicProfile(newChatTarget.profile_type);
    }

    return profilePublicByUserId[activeDirectTargetId] === true;

  }, [activeDirectTargetId, followingIdSet, isOwnProfilePublic, newChatTarget, profilePublicByUserId, selectedChat]);

  const canViewCurrentDirectContext = useMemo(() => {
    if (selectedChat?.type === "group") {
      return true;
    }

    if (!activeDirectTargetId) {
      return false;
    }

    if (isOwnProfilePublic) {
      return true;
    }

    return followingIdSet.has(activeDirectTargetId);
  }, [activeDirectTargetId, followingIdSet, isOwnProfilePublic, selectedChat]);

  async function ensureTargetPublicStatus(userId) {
    const numericId = Number(userId);
    if (!Number.isInteger(numericId) || numericId <= 0) {
      return false;
    }

    if (profilePublicByUserId[numericId] !== undefined) {
      return profilePublicByUserId[numericId] === true;
    }

    try {
      const profile = await fetchUserData(numericId);
      const isPublic = isPublicProfile(profile?.profile_type);
      setProfilePublicByUserId((prev) => ({
        ...prev,
        [numericId]: isPublic,
      }));
      return isPublic;
    } catch {
      setProfilePublicByUserId((prev) => ({
        ...prev,
        [numericId]: false,
      }));
      return false;
    }
  }

  async function canDirectMessageTarget(userId, profileType = "") {
    const targetId = Number(userId);
    if (!Number.isInteger(targetId) || targetId <= 0) {
      return false;
    }

    if (followingIdSet.has(targetId) || isOwnProfilePublic) {
      return true;
    }

    if (profileType) {
      return isPublicProfile(profileType);
    }

    return ensureTargetPublicStatus(targetId);
  }

  function canViewDirectMessagesFromTarget(userId) {
    const targetId = Number(userId);
    if (!Number.isInteger(targetId) || targetId <= 0) {
      return false;
    }

    if (isOwnProfilePublic) {
      return true;
    }

    return followingIdSet.has(targetId);
  }

  async function loadChatsAndProfile() {
    setIsLoading(true);
    setError("");

    try {
      const [profile, inboxChats, followingList] = await Promise.all([
        fetchUserData("me"),
        listChats(),
        getFollowing().catch(() => []),
      ]);
      const safeChats = Array.isArray(inboxChats) ? inboxChats : [];
      const directChatUserIds = new Set(
        safeChats
          .filter((chatItem) => chatItem?.type === "direct" && chatItem?.other_user_id)
          .map((chatItem) => Number(chatItem.other_user_id))
      );

      const mergedCandidatesMap = new Map();
      (Array.isArray(followingList) ? followingList : []).forEach((item) => {
        const normalized = normalizeUserItem(item);
        if (!normalized) {
          return;
        }

        if (Number(normalized.id) === Number(profile?.id)) {
          return;
        }

        if (directChatUserIds.has(Number(normalized.id))) {
          return;
        }

        mergedCandidatesMap.set(Number(normalized.id), normalized);
      });

      setCurrentUser(profile || null);
      setChats(safeChats);
      setIsOwnProfilePublic(isPublicProfile(profile?.profile_type));
      setNewChatCandidates(Array.from(mergedCandidatesMap.values()));
      setFollowingIds(
        (Array.isArray(followingList) ? followingList : [])
          .map((item) => Number(item?.id ?? item?.user_id ?? 0))
          .filter((id) => Number.isInteger(id) && id > 0)
      );

      if (safeChats.length > 0) {
        setSelectedChatId((prevId) => prevId || safeChats[0].chat_id);
      } else {
        setSelectedChatId(null);
      }
    } catch (loadError) {
      console.error("Failed to load messages page:", loadError);
      setError(loadError?.message || "Failed to load messages.");
      setCurrentUser(null);
      setChats([]);
      setNewChatCandidates([]);
      setFollowingIds([]);
      setIsOwnProfilePublic(false);
      setProfilePublicByUserId({});
      setSelectedChatId(null);
      setNewChatTarget(null);
    } finally {
      setIsLoading(false);
    }
  }

  async function loadMessages(chatId) {
    if (!chatId) {
      setMessages([]);
      setAccessNotice("");
      return;
    }

    const selected = chats.find((chatItem) => chatItem.chat_id === chatId) || null;
    if (selected?.type === "direct") {
      const canView = canViewDirectMessagesFromTarget(selected.other_user_id);
      if (!canView) {
        setMessages([]);
        setAccessNotice(DIRECT_VIEW_RULE_MESSAGE);
        return;
      }
    }

    setAccessNotice("");

    setIsMessagesLoading(true);
    try {
      const chatMessages = await getChatMessages(chatId, 60);
      setMessages(Array.isArray(chatMessages) ? chatMessages : []);

      const lastMessageId = Array.isArray(chatMessages) && chatMessages.length > 0
        ? chatMessages[chatMessages.length - 1]?.id || 0
        : 0;
      await markChatRead(chatId, lastMessageId).catch(() => null);
      emitUnreadRefresh();
    } catch (loadError) {
      console.error("Failed to load chat messages:", loadError);
      setMessages([]);
      setError(loadError?.message || "Failed to load chat messages.");
    } finally {
      setIsMessagesLoading(false);
    }
  }

  useEffect(() => {
    loadChatsAndProfile();
  }, []);

  useEffect(() => {
    loadMessages(selectedChatId);
  }, [selectedChatId]);

  useEffect(() => {
    selectedChatIdRef.current = selectedChatId;
  }, [selectedChatId]);

  useEffect(() => {
    const interval = setInterval(() => {
      setTimeTick(Date.now());
    }, 60 * 1000);

    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (!currentUser?.id) {
      return undefined;
    }

    let isDisposed = false;
    let reconnectTimeout = null;

    const connect = () => {
      if (isDisposed) {
        return;
      }

      try {
        const socket = new WebSocket(toWebSocketUrl("/ws"));
        socketRef.current = socket;

        socket.onopen = () => {
          if (!isDisposed) {
            setIsSocketConnected(true);
          }
        };

        socket.onmessage = async (event) => {
          if (isDisposed) {
            return;
          }

          let incoming = null;
          try {
            incoming = JSON.parse(event.data);
          } catch {
            return;
          }

          if (incoming?.event === "presence_snapshot") {
            const snapshotIds = Array.isArray(incoming?.online_user_ids)
              ? incoming.online_user_ids
                  .map((value) => Number(value))
                  .filter((value) => Number.isInteger(value) && value > 0)
              : [];
            setOnlineUserIds(snapshotIds);
            return;
          }

          if (incoming?.event === "presence") {
            const targetUserId = Number(incoming?.user_id || 0);
            if (!targetUserId) {
              return;
            }
            const isOnline = incoming?.online === true;
            setOnlineUserIds((prev) => {
              const previous = Array.isArray(prev) ? prev : [];
              if (isOnline) {
                if (previous.includes(targetUserId)) {
                  return previous;
                }
                return [...previous, targetUserId];
              }
              return previous.filter((id) => Number(id) !== targetUserId);
            });
            return;
          }

          const incomingChatId = Number(incoming?.chat_id || 0);
          const incomingMessageId = Number(incoming?.id || 0);
          if (!incomingChatId || !incomingMessageId) {
            return;
          }

          const refreshedChats = await listChats().catch(() => null);
          if (Array.isArray(refreshedChats)) {
            setChats(refreshedChats);
          }

          if (Number(selectedChatIdRef.current) === incomingChatId) {
            setMessages((prev) => {
              if (!Array.isArray(prev)) {
                return [incoming];
              }
              if (prev.some((item) => Number(item?.id) === incomingMessageId)) {
                return prev;
              }
              return [...prev, incoming];
            });
            await markChatRead(incomingChatId, incomingMessageId).catch(() => null);
            emitUnreadRefresh();
          }
        };

        socket.onerror = () => {
          if (!isDisposed) {
            setIsSocketConnected(false);
          }
          if (!isDisposed) {
            socket.close();
          }
        };

        socket.onclose = () => {
          setIsSocketConnected(false);
          if (isDisposed) {
            return;
          }
          reconnectTimeout = setTimeout(connect, 2000);
        };
      } catch {
        reconnectTimeout = setTimeout(connect, 2000);
      }
    };

    connect();

    return () => {
      isDisposed = true;
      setIsSocketConnected(false);
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
      if (socketRef.current) {
        try {
          socketRef.current.onclose = null;
          socketRef.current.close();
        } catch {
          // no-op
        }
        socketRef.current = null;
      }
    };
  }, [currentUser?.id]);

  useEffect(() => {
    if (!activeDirectTargetId) {
      return;
    }

    if (followingIdSet.has(activeDirectTargetId) || profilePublicByUserId[activeDirectTargetId] !== undefined) {
      return;
    }

    ensureTargetPublicStatus(activeDirectTargetId);
  }, [activeDirectTargetId, followingIdSet, profilePublicByUserId]);

  async function handleStartNewChat(userItem) {
    if (!userItem?.id) {
      return;
    }

    const targetId = Number(userItem.id);
    let allowed = followingIdSet.has(targetId);

    if (!allowed) {
      if (userItem.profile_type) {
        allowed = isPublicProfile(userItem.profile_type);
      } else {
        allowed = await ensureTargetPublicStatus(targetId);
      }
    }

    if (!allowed) {
      setError(DIRECT_SEND_RULE_MESSAGE);
      return;
    }

    setError("");
    setAccessNotice("");
    setSelectedChatId(null);
    setMessages([]);
    setNewChatTarget(userItem);
    setDraftMessage("");
    setNewChatQuery("");
  }

  async function handleSendMessage(event) {
    event.preventDefault();

    if ((!selectedChat && !newChatTarget) || isSending) {
      return;
    }

    const trimmedText = draftMessage.trim();
    if (!trimmedText) {
      return;
    }

    if (!canSendInCurrentContext) {
      setError(DIRECT_SEND_RULE_MESSAGE);
      return;
    }

    setIsSending(true);
    setError("");
    setAccessNotice("");

    try {
      if (selectedChat?.type === "direct" && selectedChat.other_user_id) {
        await sendDirectMessage(selectedChat.other_user_id, trimmedText);
      } else if (selectedChat?.type === "group" && selectedChat.group_id) {
        await sendGroupMessage(selectedChat.group_id, trimmedText);
      } else if (newChatTarget?.id) {
        await sendDirectMessage(newChatTarget.id, trimmedText);
      } else {
        throw new Error("Unsupported chat type.");
      }

      setDraftMessage("");
      const refreshedChats = await listChats();
      const safeRefreshedChats = Array.isArray(refreshedChats) ? refreshedChats : [];
      setChats(safeRefreshedChats);

      if (selectedChat?.chat_id) {
        setSelectedChatId(selectedChat.chat_id);
        await loadMessages(selectedChat.chat_id);
      } else if (newChatTarget?.id) {
        const newDirectChat = safeRefreshedChats.find(
          (chatItem) => chatItem?.type === "direct" && Number(chatItem.other_user_id) === Number(newChatTarget.id)
        );

        if (newDirectChat?.chat_id) {
          setSelectedChatId(newDirectChat.chat_id);
          setNewChatTarget(null);
          await loadMessages(newDirectChat.chat_id);
          setNewChatCandidates((prev) => prev.filter((candidate) => Number(candidate.id) !== Number(newDirectChat.other_user_id)));
        }
      }
    } catch (sendError) {
      console.error("Failed to send message:", sendError);
      setError(sendError?.message || "Failed to send message.");
    } finally {
      setIsSending(false);
    }
  }

  return (
    <div className="w-full max-w-5xl pb-8">
      <div className="border border-purple-800 bg-[#140026] rounded-xl overflow-hidden shadow-[0_0_0_1px_rgba(168,85,247,0.25)] min-h-[70vh]">
        <div className="grid grid-cols-[300px_1fr] min-h-[70vh]">
          <aside className="border-r border-purple-900/80 bg-[#17002b]">
            <div className="p-3 border-b border-purple-900/80">
              <div className="flex items-center rounded-full bg-[#26103b] px-3 py-2">
                <span className="text-purple-300 text-sm mr-2">⌕</span>
                <input
                  value={searchTerm}
                  onChange={(event) => setSearchTerm(event.target.value)}
                  placeholder="Search messages..."
                  className="w-full bg-transparent text-sm text-purple-100 placeholder:text-purple-400 outline-none"
                />
              </div>

              <div className="mt-2 rounded-md border border-purple-900/70 bg-[#1b0a30] px-2 py-2">
                <input
                  value={newChatQuery}
                  onChange={(event) => setNewChatQuery(event.target.value)}
                  placeholder="Start chat with..."
                  className="w-full bg-transparent text-xs text-purple-100 placeholder:text-purple-400 outline-none"
                />
                {filteredNewChatCandidates.length > 0 ? (
                  <div className="mt-2 flex flex-col gap-1">
                    {filteredNewChatCandidates.map((candidate) => (
                      <button
                        key={candidate.id}
                        type="button"
                        onClick={() => handleStartNewChat(candidate)}
                        className="w-full flex items-center gap-2 rounded px-2 py-1 text-left hover:bg-[#291144]"
                      >
                        <Image
                          src={parseProfileImage(candidate.profile_picture)}
                          alt="New chat user"
                          width={18}
                          height={18}
                          className="h-4.5 w-4.5 rounded-full"
                        />
                        <Link
                          href={`/profile/${candidate.id}`}
                          className="truncate text-xs text-purple-100 hover:text-purple-200"
                          onClick={(event) => event.stopPropagation()}
                        >
                          {candidate.display_name}
                        </Link>
                      </button>
                    ))}
                  </div>
                ) : null}
              </div>
            </div>

            <div className="max-h-[calc(70vh-64px)] overflow-y-auto">
              {isLoading ? (
                <p className="px-4 py-3 text-sm text-purple-300">Loading chats...</p>
              ) : filteredChats.length === 0 ? (
                <p className="px-4 py-3 text-sm text-purple-300">No chats found.</p>
              ) : (
                filteredChats.map((chatItem) => {
                  const isSelected = chatItem.chat_id === selectedChatId;
                  const rowTitle = formatChatTitle(chatItem);

                  return (
                    <button
                      key={chatItem.chat_id}
                      type="button"
                      onClick={() => {
                        setNewChatTarget(null);
                        setSelectedChatId(chatItem.chat_id);
                      }}
                      className={`w-full text-left px-3 py-3 border-b border-purple-900/50 transition ${
                        isSelected ? "bg-[#24103b]" : "hover:bg-[#1e0b33]"
                      }`}
                    >
                      <div className="flex items-start gap-2">
                        <Image
                          src={parseProfileImage(chatItem.other_user_picture)}
                          alt="Chat avatar"
                          width={28}
                          height={28}
                          className="h-7 w-7 rounded-full shrink-0"
                        />
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center justify-between gap-2">
                            {chatItem.type === "direct" && chatItem.other_user_id ? (
                              <Link
                                href={`/profile/${chatItem.other_user_id}`}
                                className="text-purple-100 text-sm font-semibold truncate hover:text-purple-200"
                                onClick={(event) => event.stopPropagation()}
                              >
                                {rowTitle}
                              </Link>
                            ) : (
                              <span className="text-purple-100 text-sm font-semibold truncate">{rowTitle}</span>
                            )}
                            <span className="text-[11px] text-purple-300 shrink-0">{formatListTime(chatItem.last_message_at, timeTick)}</span>
                          </div>
                          <div className="flex items-center justify-between gap-2 mt-0.5">
                            <span className="text-xs text-purple-300 truncate">
                              {chatItem.last_message_preview || "No messages yet"}
                            </span>
                            {!chatItem.seen ? <span className="h-2 w-2 rounded-full bg-blue-400 shrink-0" /> : null}
                          </div>
                        </div>
                      </div>
                    </button>
                  );
                })
              )}
            </div>
          </aside>

          <section className="flex flex-col bg-[#140026]">
            <header className="h-14 border-b border-purple-900/80 px-4 flex items-center justify-between">
              <div className="flex items-center gap-2 min-w-0">
                <Image
                  src={parseProfileImage(selectedChat?.other_user_picture || newChatTarget?.profile_picture)}
                  alt="Selected chat avatar"
                  width={24}
                  height={24}
                  className="h-6 w-6 rounded-full"
                />
                <div className="min-w-0">
                  {selectedChat?.type === "direct" && selectedChat?.other_user_id ? (
                    <Link
                      href={`/profile/${selectedChat.other_user_id}`}
                      className="text-sm font-semibold text-purple-100 truncate hover:text-purple-200"
                    >
                      {formatChatTitle(selectedChat)}
                    </Link>
                  ) : newChatTarget?.id ? (
                    <Link
                      href={`/profile/${newChatTarget.id}`}
                      className="text-sm font-semibold text-purple-100 truncate hover:text-purple-200"
                    >
                      {newChatTarget.display_name}
                    </Link>
                  ) : (
                    <h1 className="text-sm font-semibold text-purple-100 truncate">Messages</h1>
                  )}
                  {headerActivityText && (
                    <p className="text-[11px] text-purple-300">{headerActivityText}</p>
                  )}
                </div>
              </div>
              {selectedChat || newChatTarget ? (
                <div className="flex items-center gap-3 text-purple-300 text-sm">
                  {activeDirectTargetId ? (
                    <span className="inline-flex items-center gap-1.5 text-[11px] text-purple-200">
                      <span className={`h-2 w-2 rounded-full ${isDirectTargetOnline ? "bg-green-400" : "bg-gray-500"}`} />
                      {isDirectTargetOnline ? "Online" : "Offline"}
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1.5 text-[11px] text-purple-200">
                      <span className={`h-2 w-2 rounded-full ${isSocketConnected ? "bg-green-400" : "bg-gray-500"}`} />
                      {isSocketConnected ? "Live" : "Offline"}
                    </span>
                  )}
                </div>
              ) : null}
            </header>

            <div className="flex-1 overflow-y-auto px-3 py-3">
              {isMessagesLoading ? (
                <p className="text-sm text-purple-300">Loading messages...</p>
              ) : !selectedChat && !newChatTarget ? (
                <p className="text-sm text-purple-300">Select a chat to start messaging.</p>
              ) : !selectedChat && newChatTarget ? (
                <p className="text-sm text-purple-300">Send the first message to start chatting with {newChatTarget.display_name}.</p>
              ) : accessNotice ? (
                <p className="text-sm text-red-300">{accessNotice}</p>
              ) : messages.length === 0 ? (
                <p className="text-sm text-purple-300">No messages yet.</p>
              ) : (
                <div className="flex flex-col gap-3">
                  {messages.map((message) => {
                    const isMine = Number(message.sender_id) === Number(currentUser?.id);
                    return (
                      <div key={message.id} className={`flex ${isMine ? "justify-end" : "justify-start"}`}>
                        <div className="max-w-[70%] rounded-lg border border-purple-500/70 bg-[#3a1c67] px-3 py-2">
                          <p className="text-xs text-purple-100 wrap-break-word">{message.body || ""}</p>
                          <p className="text-[10px] text-purple-300 mt-1">{formatSmallTime(message.created_at)}</p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
              {selectedChat?.type === "direct" && !canSendInCurrentContext ? (
                <p className="text-xs text-red-300 mt-2">{DIRECT_SEND_RULE_MESSAGE}</p>
              ) : null}
            </div>

            <footer className="h-14 border-t border-purple-900/80 px-3 flex items-center gap-2">
              <form onSubmit={handleSendMessage} className="w-full flex items-center gap-2">
                <input
                  value={draftMessage}
                  onChange={(event) => setDraftMessage(event.target.value)}
                  placeholder="Type a message here..."
                  className="flex-1 h-9 rounded-full bg-[#26103b] px-4 text-sm text-purple-100 placeholder:text-purple-400 outline-none"
                  disabled={(!selectedChat && !newChatTarget) || isSending || !canSendInCurrentContext}
                />
                <button
                  type="submit"
                  disabled={(!selectedChat && !newChatTarget) || isSending || !draftMessage.trim() || !canSendInCurrentContext}
                  className="h-9 w-9 flex items-center justify-center rounded-full bg-[#5a37ff] hover:bg-[#4c2df3] disabled:opacity-50"
                >
                  <Image
                    src="/share_icon.svg"
                    alt="Send message"
                    width={16}
                    height={16}
                  />
                </button>
              </form>
            </footer>
          </section>
        </div>
      </div>

      {error ? <p className="text-red-300 text-sm mt-3">{error}</p> : null}
    </div>
  );
};

export default MessagesPage;
