import { apiRequest } from "src/lib/apiClient";

export async function listChats(limit = 30, offset = 0) {
  const payload = await apiRequest(`/api/chats?limit=${limit}&offset=${offset}`, {
    method: "GET",
  });

  return Array.isArray(payload?.data?.items) ? payload.data.items : [];
}

export async function getChatMessages(chatId, limit = 50, beforeId = 0) {
  const beforePart = beforeId > 0 ? `&before_id=${beforeId}` : "";
  const payload = await apiRequest(`/api/chats/${chatId}/messages?limit=${limit}${beforePart}`, {
    method: "GET",
  });

  return Array.isArray(payload?.data?.items) ? payload.data.items : [];
}

export async function sendDirectMessage(targetUserId, body) {
  return apiRequest(`/api/chats/direct/${targetUserId}/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      message_type: "text",
      body,
      media_url: "",
    }),
  });
}

export async function sendGroupMessage(groupId, body) {
  return apiRequest(`/api/groups/${groupId}/chat/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      message_type: "text",
      body,
      media_url: "",
    }),
  });
}

export async function markChatRead(chatId, lastMessageId = 0) {
  return apiRequest(`/api/chats/${chatId}/read`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      last_message_id: lastMessageId,
    }),
  });
}
