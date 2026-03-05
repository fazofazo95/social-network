# Notifications API Endpoints

All endpoints below require authentication (`session_id` cookie) and return the standard envelope:

```json
{
  "status": "success|error",
  "message": "human readable message",
  "data": {}
}
```

---

## 📌 Table of Contents
* [List Notifications](#1-list-notifications)
* [Mark Notification as Seen](#2-mark-notification-as-seen)
* [Mark All Notifications as Seen](#3-mark-all-notifications-as-seen)
* [Respond to Notification](#4-respond-to-notification-acceptreject)
* [Notification SSE Stream](#5-notification-sse-stream)

---

## Notification Types

| Type | Description |
| --- | --- |
| `follow_request` | Someone sent a follow request |
| `group_invite` | Invited to join a group |
| `group_join_request` | Someone requested to join your group |
| `group_event_created` | A new event was created in a group |

## Notification Statuses

| Status | Description |
| --- | --- |
| `pending` | Actionable, not yet responded to |
| `accepted` | Action was accepted |
| `rejected` | Action was rejected |
| `read` | Seen but not yet acted upon (for actionable notifications) |

---

## 1) List Notifications

### `GET /api/notifications?limit=20&offset=0`
**Who can use:** any authenticated user.

Returns the current user's notifications, newest first, with actor info attached.

#### **Query Parameters**
| Parameter | Type | Default | Description |
| --- | --- | --- | --- |
| `limit` | integer | `20` | Number of notifications per page. |
| `offset` | integer | `0` | Offset for pagination. |

#### **Success Response (200 OK)**
```json
{
  "status": "success",
  "message": "notifications fetched",
  "data": {
    "items": [
      {
        "id": 1,
        "recipient_id": 1,
        "actor_id": 2,
        "type": "follow_request",
        "status": "pending",
        "content": "sent you a follow request",
        "metadata": "",
        "seen": false,
        "created_at": "2026-03-01 10:00:00",
        "updated_at": "2026-03-01 10:00:00",
        "actor_first_name": "Bob",
        "actor_last_name": "Baker",
        "actor_picture": "bob.jpg"
      },
      {
        "id": 2,
        "recipient_id": 1,
        "actor_id": 3,
        "type": "group_invite",
        "status": "pending",
        "group_id": 5,
        "content": "invited you to join Hiking Club",
        "metadata": "",
        "seen": false,
        "created_at": "2026-03-01 09:30:00",
        "updated_at": "2026-03-01 09:30:00",
        "actor_first_name": "Carol",
        "actor_last_name": "Clark",
        "actor_picture": "carol.jpg"
      }
    ],
    "limit": 20,
    "offset": 0,
    "has_more": false
  }
}
```

#### **Notification Object Fields**

| Field | Type | Description |
| --- | --- | --- |
| `id` | int | Unique notification identifier. |
| `recipient_id` | int | ID of the user receiving the notification. |
| `actor_id` | int | ID of the user who triggered the notification (optional). |
| `type` | string | Notification type (see Notification Types above). |
| `status` | string | Current status (see Notification Statuses above). |
| `group_id` | int | Associated group ID (optional, for group-related notifications). |
| `event_id` | int | Associated event ID (optional, for event notifications). |
| `content` | string | Human-readable notification content. |
| `metadata` | string | Additional metadata (JSON string, may be empty). |
| `seen` | bool | Whether the user has seen this notification. |
| `created_at` | string | Creation timestamp. |
| `updated_at` | string | Last update timestamp. |
| `actor_first_name` | string | Actor's first name. |
| `actor_last_name` | string | Actor's last name. |
| `actor_picture` | string | Actor's profile picture URL. |

---

## 2) Mark Notification as Seen

### `POST /api/notifications/{id}/read`
**Who can use:** recipient of the notification only.

Marks a single notification as seen.

#### **Path Parameters**
| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| `id` | integer | ✅ | The notification ID. |

#### **Success Response (200 OK)**
```json
{
  "status": "success",
  "message": "notification marked as seen",
  "data": {
    "notification_id": 1
  }
}
```

#### **Errors**
- **400 Bad Request:** Invalid notification ID.
- **401 Unauthorized:** Missing or invalid session.
- **404 Not Found:** Notification not found or does not belong to user.

---

## 3) Mark All Notifications as Seen

### `POST /api/notifications/read-all`
**Who can use:** any authenticated user.

Marks all of the current user's unseen notifications as seen.

#### **Success Response (200 OK)**
```json
{
  "status": "success",
  "message": "notifications marked as seen",
  "data": {
    "updated": 5
  }
}
```

#### **Errors**
- **401 Unauthorized:** Missing or invalid session.
- **500 Internal Server Error:** Database error.

---

## 4) Respond to Notification (Accept/Reject)

### `POST /api/notifications/{id}/action`
**Who can use:** recipient of the notification only.

Performs an accept or reject action on an actionable notification. This triggers the corresponding backend action (e.g., accepting a follow request, accepting a group invite, or approving a group join request).

#### **Path Parameters**
| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| `id` | integer | ✅ | The notification ID. |

#### **Request Body**
```json
{
  "action": "accept"
}
```

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `action` | string | ✅ | Must be `accept` or `reject`. |

#### **Actionable notification types**

| Notification Type | Accept action | Reject action |
| --- | --- | --- |
| `follow_request` | Accepts the follow request | Rejects and removes the follow request |
| `group_invite` | Accepts the group invitation | Rejects the group invitation |
| `group_join_request` | Approves the join request | Rejects the join request |

#### **Success Response (200 OK)**

Accepted:
```json
{
  "status": "success",
  "message": "notification accepted",
  "data": {
    "notification_id": 1,
    "status": "accepted"
  }
}
```

Rejected:
```json
{
  "status": "success",
  "message": "notification rejected",
  "data": {
    "notification_id": 1,
    "status": "rejected"
  }
}
```

#### **Errors**
- **400 Bad Request:** Invalid notification ID, invalid JSON, action must be `accept` or `reject`, or notification type does not support action.
- **401 Unauthorized:** Missing or invalid session.
- **403 Forbidden:** Notification does not belong to user.
- **404 Not Found:** Notification not found.
- **409 Conflict:** Notification is already resolved (already accepted or rejected).

---

## 5) Notification SSE Stream

### `GET /api/notifications/stream`
**Who can use:** any authenticated user.

Opens a Server-Sent Events (SSE) connection for real-time notification delivery. The connection stays open and pushes new notifications as they occur.

#### **Response Headers**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

#### **Events**

**`ready` event** (sent immediately on connection):
```
event: ready
data: {"status":"ok"}
```

**`ping` event** (sent every 25 seconds as heartbeat):
```
event: ping
data: {}
```

**Notification event** (sent when a new notification arrives):
```
event: notification
data: {"event":"notification","notification":{"id":1,"recipient_id":1,"actor_id":2,"type":"follow_request","status":"pending","content":"sent you a follow request","metadata":"","seen":false,"created_at":"2026-03-01 10:00:00","updated_at":"2026-03-01 10:00:00","actor_first_name":"Bob","actor_last_name":"Baker","actor_picture":"bob.jpg"}}
```

#### **Notes**
- The SSE stream automatically subscribes the user to notifications and unsubscribes on disconnect.
- Heartbeat pings keep the connection alive through proxies.
- Frontend should reconnect on connection loss.

#### **Errors**
- **401 Unauthorized:** Missing or invalid session.
- **500 Internal Server Error:** Streaming not supported by server.
