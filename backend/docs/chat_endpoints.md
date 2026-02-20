# Chat Endpoints (MVP)

Base: authenticated endpoints (`session_id` cookie required).

## Message types
- `text`: requires `body`
- `image`: requires `media_url`
- `text_image`: requires both `body` and `media_url`

## Endpoints

### 1) Send direct message
`POST /api/chats/direct/{user_id}/messages`

Body:
```json
{
  "message_type": "text",
  "body": "hello",
  "media_url": ""
}
```

Behavior:
- Auto-creates direct chat lazily on first message if allowed.
- Allowed only when users have accepted follow relationship and no block between them.

---

### 2) Send group message
`POST /api/groups/{id}/chat/messages`

Body:
```json
{
  "message_type": "image",
  "media_url": "/uploads/chat/abc.png"
}
```

Behavior:
- Sender must be active member in that group.
- Uses the auto-created group chat for that group.

---

### 3) List inbox chats
`GET /api/chats?limit=30&offset=0`

Returns list with:
- `chat_id`, `type` (`direct|group`)
- direct peer info (`other_user_*`) for direct chats
- `last_message_*` metadata
- `seen` (per current user)

`seen` is true when:
- chat has no messages, or
- current user sent last message, or
- `chat_participants.last_read_message_id >= chats.last_message_id`

---

### 4) Get chat messages (history + pagination)
`GET /api/chats/{chat_id}/messages?before_id=123&limit=30`

Behavior:
- Requires membership in chat.
- Returns ascending order (oldest -> newest) in the requested page.
- Use `before_id` for older pages on scroll-up.

---

### 5) Mark chat as read
`POST /api/chats/{chat_id}/read`

Body (optional):
```json
{
  "last_message_id": 456
}
```

Behavior:
- If `last_message_id` omitted or `0`, server marks up to current chat last message.
- Stores per-user read pointer in `chat_participants.last_read_message_id`.

## Notes
- Full history is stored in `chat_messages`.
- Inbox speed is optimized via `chats.last_message_id` and `chats.last_message_at`.
- Group chat participants are synced with active `group_members` lifecycle (join/accept/kick/leave).
