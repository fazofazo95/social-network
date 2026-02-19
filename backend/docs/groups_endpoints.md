# Groups API Endpoints

This document covers **all group-related endpoints** currently implemented in the backend.

All endpoints below:
- require authenticated session (`session_id` cookie)
- return the standard envelope:

```json
{
  "status": "success|error",
  "message": "human readable message",
  "data": {}
}
```

---

## Enums used

### `visibility`
- `public`
- `private`

### `join_mode`
- `auto`
- `request`
- `invite`
- `request_and_invite`

### Group member role (`group_members.role`)
- `owner`
- `moderator`
- `member`

### Group membership status (`group_members.status`)
- `active`
- `requested`

### Pending item type in list/group page responses
- `requested`
- `invited`
- `none` (group page only)

---

## Endpoint Index

- `POST /api/groups`
- `GET /api/groups/{id}`
- `GET /api/groups/{id}/members`
- `GET /api/groups/{id}/requests/pending`
- `GET /api/groups/{id}/invites/pending`
- `POST /api/groups/{id}/join`
- `DELETE /api/groups/{id}/requests/me`
- `POST /api/groups/{id}/invite/{user_id}`
- `POST /api/groups/{id}/invite/accept`
- `POST /api/groups/{id}/invite/reject`
- `DELETE /api/groups/{id}/invites/{user_id}`
- `POST /api/groups/{id}/requests/{user_id}/accept`
- `POST /api/groups/{id}/requests/{user_id}/reject`
- `POST /api/groups/{id}/members/{user_id}/kick`
- `POST /api/groups/{id}/members/{user_id}/promote`
- `POST /api/groups/{id}/members/{user_id}/demote`
- `GET /api/groups/{id}/settings`
- `PATCH /api/groups/{id}/settings`
- `PUT /api/groups/{id}/settings`
- `POST /api/groups/{id}/leave`
- `DELETE /api/groups/{id}`

---

## 1) Create Group

### `POST /api/groups`
**Who can use:** any authenticated user.

### Request
```json
{
  "name": "Hiking Club",
  "description": "Group for mountain lovers",
  "visibility": "public",
  "join_mode": "request_and_invite"
}
```

### Success (201)
```json
{
  "status": "success",
  "message": "group created successfully",
  "data": {
    "id": 1,
    "name": "Hiking Club",
    "description": "Group for mountain lovers",
    "owner_id": 1,
    "visibility": "public",
    "join_mode": "request_and_invite",
    "group_members": 1,
    "created_at": "2026-02-19 08:36:04"
  }
}
```

---

## 2) Show Group Page

### `GET /api/groups/{id}`
**Who can use:** any authenticated user.

Behavior:
- if viewer is active member: return full group info + viewer role
- if viewer is non-member and group is public: return full group info + pending type + can_request
- if viewer is non-member and group is private: return limited public metadata only

### A) Public group, no pending request/invite (can request)
```json
{
  "status": "success",
  "message": "group page",
  "data": {
    "id": 44,
    "name": "Group Page 1771495736063469300",
    "description": "group page integration",
    "visibility": "public",
    "join_mode": "request_and_invite",
    "group_picture": "",
    "group_members": 1,
    "created_at": "2026-02-19 10:08:56",
    "pending_type": "none",
    "can_request": true
  }
}
```

### B) Public group, pending invite
```json
{
  "status": "success",
  "message": "group page",
  "data": {
    "id": 44,
    "name": "Group Page 1771495736063469300",
    "description": "group page integration",
    "visibility": "public",
    "join_mode": "request_and_invite",
    "group_picture": "",
    "group_members": 1,
    "created_at": "2026-02-19 10:08:56",
    "pending_type": "invited",
    "can_request": false
  }
}
```

### C) Public group, pending request
```json
{
  "status": "success",
  "message": "group page",
  "data": {
    "id": 44,
    "name": "Group Page 1771495736063469300",
    "description": "group page integration",
    "visibility": "public",
    "join_mode": "request_and_invite",
    "group_picture": "",
    "group_members": 1,
    "created_at": "2026-02-19 10:08:56",
    "pending_type": "requested",
    "can_request": false
  }
}
```

### Member view example
```json
{
  "status": "success",
  "message": "group page",
  "data": {
    "id": 10,
    "name": "Hiking Club",
    "description": "Group for mountain lovers",
    "visibility": "public",
    "join_mode": "request",
    "group_picture": "",
    "group_members": 12,
    "created_at": "2026-02-19 10:00:00",
    "role": "moderator"
  }
}
```

### Private non-member view example
```json
{
  "status": "success",
  "message": "group page",
  "data": {
    "id": 11,
    "name": "Private Team",
    "description": "Internal coordination",
    "visibility": "private",
    "join_mode": "invite",
    "group_picture": ""
  }
}
```

---

## 3) Members + Pending Lists

### `GET /api/groups/{id}/members`
**Who can use:** any authenticated user.

Returns active members only.

### Success (200)
```json
{
  "status": "success",
  "message": "group active members",
  "data": [
    {
      "id": 1,
      "first_name": "Alice",
      "last_name": "Anderson",
      "profile_picture": "alice.jpg",
      "group_status": "owner"
    },
    {
      "id": 2,
      "first_name": "Bob",
      "last_name": "Baker",
      "profile_picture": "bob.jpg",
      "group_status": "moderator"
    }
  ]
}
```

### `GET /api/groups/{id}/requests/pending`
**Who can use:** owner or moderator only.

Returns pending join requests (`request_type=request`, `status=request`).

### Success (200)
```json
{
  "status": "success",
  "message": "group pending requests",
  "data": [
    {
      "id": 3,
      "first_name": "Carol",
      "last_name": "Clark",
      "profile_picture": "carol.jpg",
      "type": "requested"
    }
  ]
}
```

### `GET /api/groups/{id}/invites/pending`
**Who can use:** owner or moderator only.

Returns pending invites (`request_type=invite`, `status=invite`).

### Success (200)
```json
{
  "status": "success",
  "message": "group pending invites",
  "data": [
    {
      "id": 4,
      "first_name": "Dave",
      "last_name": "Dawson",
      "profile_picture": "dave.jpg",
      "type": "invited"
    }
  ]
}
```

### Example forbidden (403)
```json
{
  "status": "error",
  "message": "only group owner or moderators can view pending requests"
}
```

```json
{
  "status": "error",
  "message": "only group owner or moderators can view sent invites"
}
```

---

## 4) Join / Request Flow

### `POST /api/groups/{id}/join`
**Who can use:** any authenticated user.

- blocked if private
- blocked if invite-only
- behavior depends on `join_mode` for public groups:
  - `auto` => active membership immediately
  - `request` / `request_and_invite` => pending request row

### Success examples
```json
{
  "status": "success",
  "message": "joined group successfully",
  "data": {
    "group_id": 6,
    "user_id": 2,
    "membership_status": "active"
  }
}
```

```json
{
  "status": "success",
  "message": "join request submitted",
  "data": {
    "group_id": 7,
    "user_id": 2,
    "membership_status": "requested"
  }
}
```

### Remove own pending request
#### `DELETE /api/groups/{id}/requests/me`
**Who can use:** only the user who owns that pending request.

Removes pending request from both `group_members` and `group_join_requests`.

```json
{
  "status": "success",
  "message": "pending request removed",
  "data": {
    "group_id": 43,
    "user_id": 3
  }
}
```

---

## 5) Invite Flow

### `POST /api/groups/{id}/invite/{user_id}`
**Who can use:** owner or moderator.

Current behavior:
- creates a pending invite (`type=invited`) even if group mode is `auto`.
- no immediate active acceptance on invite send.

```json
{
  "status": "success",
  "message": "group invitation sent",
  "data": {
    "group_id": 20,
    "invited_by": 1,
    "user_id": 3,
    "membership_status": "requested"
  }
}
```

### `POST /api/groups/{id}/invite/accept`
**Who can use:** invitee only (authenticated user accepting own invite).

```json
{
  "status": "success",
  "message": "group invite accepted",
  "data": {
    "group_id": 26,
    "user_id": 2
  }
}
```

### `POST /api/groups/{id}/invite/reject`
**Who can use:** invitee only.

```json
{
  "status": "success",
  "message": "group invite rejected",
  "data": {
    "group_id": 27,
    "user_id": 3
  }
}
```

### Remove pending invite (owner/mod)
#### `DELETE /api/groups/{id}/invites/{user_id}`
**Who can use:** owner or moderator.

Removes pending invite from both `group_members` and `group_join_requests`.
Only valid for non-active pending invite state.

```json
{
  "status": "success",
  "message": "pending invite removed",
  "data": {
    "group_id": 43,
    "removed_by": 2,
    "user_id": 4
  }
}
```

---

## 6) Approve / Reject Join Requests (moderation)

### `POST /api/groups/{id}/requests/{user_id}/accept`
**Who can use:** owner or moderator.

```json
{
  "status": "success",
  "message": "group join request approved",
  "data": {
    "group_id": 11,
    "approved_by": 1,
    "user_id": 2
  }
}
```

### `POST /api/groups/{id}/requests/{user_id}/reject`
**Who can use:** owner or moderator.

```json
{
  "status": "success",
  "message": "group join request rejected",
  "data": {
    "group_id": 15,
    "rejected_by": 1,
    "user_id": 2
  }
}
```

---

## 7) Kick / Leave / Role Management

### `POST /api/groups/{id}/members/{user_id}/kick`
**Who can use:** owner or moderator.

Rules:
- target must be active member with role `member`
- cannot kick owner/moderator
- decrements `groups.group_members` by 1

```json
{
  "status": "success",
  "message": "group member kicked",
  "data": {
    "group_id": 28,
    "kicked_by": 2,
    "user_id": 3
  }
}
```

### `POST /api/groups/{id}/leave`
**Who can use:** any active member.

Rules:
- non-owner leaves => member removed, count decremented
- owner leaves:
  - promote random active moderator to owner
  - else promote random active member to owner
  - else delete group if nobody else exists

Success examples:
```json
{
  "status": "success",
  "message": "left group successfully",
  "data": {
    "group_id": 29,
    "user_id": 2
  }
}
```

```json
{
  "status": "success",
  "message": "left group and ownership transferred",
  "data": {
    "group_id": 30,
    "user_id": 1,
    "owner_transferred": true,
    "new_owner_id": 2
  }
}
```

```json
{
  "status": "success",
  "message": "left group and group was deleted",
  "data": {
    "group_id": 32,
    "user_id": 1,
    "group_deleted": true
  }
}
```

### `POST /api/groups/{id}/members/{user_id}/promote`
**Who can use:** owner only.

Promotes active `member` => `moderator`.

```json
{
  "status": "success",
  "message": "member promoted to moderator",
  "data": {
    "group_id": 37,
    "promoted_by": 1,
    "user_id": 2
  }
}
```

### `POST /api/groups/{id}/members/{user_id}/demote`
**Who can use:** owner only.

Demotes active `moderator` => `member`.

```json
{
  "status": "success",
  "message": "moderator demoted to member",
  "data": {
    "group_id": 37,
    "demoted_by": 1,
    "user_id": 2
  }
}
```

---

## 8) Group Settings (owner only)

### `GET /api/groups/{id}/settings`
**Who can use:** owner only.

```json
{
  "status": "success",
  "message": "group settings retrieved successfully",
  "data": {
    "group_id": 40,
    "visibility": "private",
    "join_mode": "request_and_invite"
  }
}
```

### `PATCH /api/groups/{id}/settings`
### `PUT /api/groups/{id}/settings`
**Who can use:** owner only.

Partial update is supported.

Request:
```json
{
  "visibility": "private",
  "join_mode": "request_and_invite"
}
```

Response:
```json
{
  "status": "success",
  "message": "group settings updated successfully",
  "data": {
    "group_id": 40,
    "visibility": "private",
    "join_mode": "request_and_invite"
  }
}
```

Invalid value example (400):
```json
{
  "status": "error",
  "message": "join_mode must be auto, request, invite, or request_and_invite"
}
```

---

## 9) Delete Group

### `DELETE /api/groups/{id}`
**Who can use:** owner only.

Deletes:
- `groups` row
- all `group_members` rows for that group
- all `group_join_requests` rows for that group

```json
{
  "status": "success",
  "message": "group deleted successfully",
  "data": {
    "group_id": 2
  }
}
```

---

## Typical Error Codes by Endpoint Family

- `400 Bad Request`: invalid path id or invalid enum payload
- `401 Unauthorized`: missing/invalid session
- `403 Forbidden`: role/ownership permission denied
- `404 Not Found`: group or pending item not found
- `409 Conflict`: invalid state transition (already member, active when pending required, wrong role state)
- `500 Internal Server Error`: unexpected backend/database error

---

## Frontend Notes

- For group page CTA logic on public groups:
  - use `pending_type` (`none`/`requested`/`invited`)
  - use `can_request` for showing request/join button
- For role badges in members list, use `group_status`.
- For pending queues, use `type` field (`requested` or `invited`) to render tabs or tags.
