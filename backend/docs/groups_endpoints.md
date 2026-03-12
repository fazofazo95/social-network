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

### Group event participation state (`my_reaction` in events timeline)
- `pending`
- `going`
- `not_going`
- `null` (no row yet, member can still respond)

---

## Endpoint Index

- `POST /api/groups`
- `GET /api/groups/discover`
- `GET /api/groups/active`
- `GET /api/groups/requests/pending`
- `GET /api/groups/invites/pending`
- `GET /api/groups/{id}`
- `POST /api/groups/{id}/events`
- `GET /api/groups/{id}/events`
- `POST /api/groups/{id}/events/{event_id}/respond`
- `PATCH /api/groups/{id}/events/{event_id}/respond`
- `DELETE /api/groups/{id}/events/{event_id}`
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
- `POST /api/groups/{id}/posts`
- `GET /api/groups/{id}/posts`
- `DELETE /api/groups/{id}/posts/{post_id}`

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

## 2) Discover Groups (paginated)

### `GET /api/groups/discover?page=1`
**Who can use:** any authenticated user.

Returns a paginated batch of groups for discovery.

Rules applied:
- batch size is fixed at `10`
- `page` is 1-based (`page=1` by default)
- only groups with `visibility = public`
- excludes groups with `join_mode = invite`
- excludes groups where current user is already involved:
  - `group_members.status IN ('active', 'requested', 'invited')`, or
  - open request/invite in `group_join_requests.status IN ('request', 'invite')`

### Success (200)
```json
{
  "status": "success",
  "message": "discover groups",
  "data": {
    "page": 1,
    "limit": 10,
    "items": [
      {
        "id": 7,
        "name": "Hiking Club",
        "description": "Group for mountain lovers",
        "group_picture": "",
        "group_members": 12,
        "owner_first_name": "Alice",
        "owner_last_name": "Anderson",
        "type": "request_and_invite"
      },
      {
        "id": 5,
        "name": "Cycling Team",
        "description": "Road cycling enthusiasts",
        "group_picture": "/uploads/groups/cycle.png",
        "group_members": 8,
        "owner_first_name": "Dave",
        "owner_last_name": "Dawson",
        "type": "auto"
      }
    ]
  }
}
```

## 2b) Active Groups (current user)

### `GET /api/groups/active`
**Who can use:** any authenticated user.

Returns groups where the current user has an active membership.

### Success (200)
```json
{
  "status": "success",
  "message": "active groups",
  "data": [
    {
      "id": 12,
      "name": "Board Games Club",
      "description": "Weekly meetups",
      "group_picture": "",
      "group_members": 23,
      "owner_id": 3,
      "owner_first_name": "Alex",
      "owner_last_name": "Lee",
      "role": "member",
      "created_at": "2026-02-19 10:08:56"
    }
  ]
}
```

## 2c) Pending Group Requests (current user)

### `GET /api/groups/requests/pending`
**Who can use:** any authenticated user.

Returns groups where the current user has a pending join request.

### Success (200)
```json
{
  "status": "success",
  "message": "pending group requests",
  "data": [
    {
      "id": 12,
      "name": "Board Games Club",
      "description": "Weekly meetups",
      "group_picture": "",
      "group_members": 23,
      "owner_id": 3,
      "owner_first_name": "Alex",
      "owner_last_name": "Lee",
      "join_mode": "request",
      "requested_at": "2026-02-19 10:08:56",
      "type": "requested"
    }
  ]
}
```

## 2d) Pending Group Invites (current user)

### `GET /api/groups/invites/pending`
**Who can use:** any authenticated user.

Returns groups where the current user has a pending invite.

### Success (200)
```json
{
  "status": "success",
  "message": "pending group invites",
  "data": [
    {
      "id": 14,
      "name": "Hiking Crew",
      "description": "Weekend trails",
      "group_picture": "",
      "group_members": 8,
      "owner_id": 5,
      "owner_first_name": "Maya",
      "owner_last_name": "Ortiz",
      "join_mode": "invite",
      "requested_at": "2026-02-20 14:33:12",
      "type": "invited"
    }
  ]
}
```

## 2e) Create Group Event

### `POST /api/groups/{id}/events`
**Who can use:** any active member of the group.

Creates a group event.

Behavior:
- all currently active members are initialized as `pending`
- creator is also `pending` at creation time (not auto-`going`)
- all active members receive a `group_event_created` notification
- active members who join the group later are not backfilled as `pending`, but can still respond (see `POST /respond`)

### Request body
```json
{
  "title": "Game Night",
  "description": "Bring your favorite board game",
  "event_day": "2026-03-01",
  "event_time": "19:30"
}
```

### Success (201)
```json
{
  "status": "success",
  "message": "group event created",
  "data": {
    "id": 5,
    "group_id": 12,
    "creator_id": 3,
    "title": "Game Night",
    "description": "Bring your favorite board game",
    "event_day": "2026-03-01",
    "event_time": "19:30",
    "created_at": "2026-02-25 14:03:11",
    "going": 0,
    "not_going": 0,
    "pending": 8
  }
}
```

## 2f) List Group Events Timeline

### `GET /api/groups/{id}/events`
**Who can use:** any active member of the group.

Returns two lists for this group only:
- `upcoming_events`: event date/time has not passed yet
- `older_events`: event date/time has already passed
- each event includes `my_reaction`:
  - `pending`
  - `going`
  - `not_going`
  - `null` (member has no event reaction row yet, e.g. joined group after event creation)

### Success (200)
```json
{
  "status": "success",
  "message": "group events fetched",
  "data": {
    "upcoming_events": [
      {
        "id": 5,
        "group_id": 12,
        "creator_id": 3,
        "title": "Game Night",
        "description": "Bring your favorite board game",
        "event_day": "2026-03-01",
        "event_time": "19:30:00",
        "created_at": "2026-02-25 14:03:11",
        "going": 4,
        "not_going": 1,
        "pending": 7,
        "my_reaction": "pending"
      },
      {
        "id": 9,
        "group_id": 12,
        "creator_id": 7,
        "title": "Weekend Run",
        "description": "5k around the park",
        "event_day": "2026-03-03",
        "event_time": "08:00:00",
        "created_at": "2026-02-27 09:10:00",
        "going": 2,
        "not_going": 0,
        "pending": 5,
        "my_reaction": null
      }
    ],
    "older_events": []
  }
}
```

## 2g) Event Invite Endpoints (Removed)

Event invite endpoints were removed from the API. Event participation now happens directly through `respond`/`change response` and timeline listing.

## 2h) Respond to Group Event Invite

### `POST /api/groups/{id}/events/{event_id}/respond`
**Who can use:** any active member.

Records a first response to an event.

Rules:
- valid `reaction_type`: `going`, `not_going`
- if current state is `pending`, converts pending to selected response
- if current state is `null`, creates response directly (member can still join event)
- if already `going` or `not_going`, returns conflict (use `PATCH` endpoint)

### Request body
```json
{
  "reaction_type": "going"
}
```

### Success (200)
```json
{
  "status": "success",
  "message": "group event response recorded",
  "data": {
    "group_id": 12,
    "event_id": 5,
    "reaction_type": "going"
  }
}
```

### Example errors
```json
{
  "status": "error",
  "message": "reaction_type must be going or not_going"
}
```

```json
{
  "status": "error",
  "message": "event response already recorded"
}
```

## 2i) Change Group Event Response

### `PATCH /api/groups/{id}/events/{event_id}/respond`
**Who can use:** members who already responded.

Changes a response between `going` and `not_going`.

### Request body
```json
{
  "reaction_type": "not_going"
}
```

### Success (200)
```json
{
  "status": "success",
  "message": "group event response updated",
  "data": {
    "group_id": 12,
    "event_id": 5,
    "reaction_type": "not_going"
  }
}
```

### Example errors
```json
{
  "status": "error",
  "message": "no existing response to change"
}
```

## 2j) Cancel Group Event

### `DELETE /api/groups/{id}/events/{event_id}`
**Who can use:** owner, moderator, or event creator.

Deletes the event and all related rows.

### Success (200)
```json
{
  "status": "success",
  "message": "group event cancelled",
  "data": {
    "group_id": 12,
    "event_id": 5
  }
}
```

### Example forbidden (403)
```json
{
  "status": "error",
  "message": "only group owner, moderators, or event creator can cancel events"
}
```

---

## 3) Show Group Page

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

## 4) Members + Pending Lists

Note:
- These pending lists are for **group membership workflow** (`request` / `invite`).
- They are not related to group event pending participation.

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

### Empty list (200)
```json
{
  "status": "success",
  "message": "group pending requests",
  "data": []
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

### Empty list (200)
```json
{
  "status": "success",
  "message": "group pending invites",
  "data": []
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

## 5) Join / Request Flow

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

## 6) Invite Flow

### `POST /api/groups/{id}/invite/{user_id}`
**Who can use:** owner or moderator.

Current behavior:
- works for both `public` and `private` groups
- inviter must be an **active** group member with role `owner` or `moderator`
- creates a pending invite (`type=invited`) even if group mode is `auto`.
- no immediate active acceptance on invite send.

```json
{
  "status": "success",
  "message": "user invited to group",
  "data": {
    "group_id": 20,
    "invited_user_id": 3,
    "membership_status": "requested"
  }
}
```

### `POST /api/groups/{id}/invite/accept`
**Who can use:** invitee only (authenticated user accepting own invite).

```json
{
  "status": "success",
  "message": "invite accepted",
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
  "message": "invite rejected",
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

## 7) Approve / Reject Join Requests (moderation)

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

## 8) Kick / Leave / Role Management

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

## 10) Group Posts

### 10.1) Create Group Post

### `POST /api/groups/{id}/posts`
**Who can use:** active group members only.

### Request
Multipart form:
- `content` (string) — post text content
- `avatar` (file, optional) — image attachment

At least one of `content` or image must be provided.

### Success (201)
```json
{
  "status": "success",
  "message": "group post created",
  "data": {
    "id": 42,
    "user_id": 3,
    "group_id": 1,
    "content": "Hello everyone!",
    "extra_content": "",
    "created_at_time": "2026-03-06T10:00:00Z",
    "author_first_name": "Alice",
    "author_last_name": "Johnson",
    "author_profile_picture": "/uploads/abc.jpg"
  }
}
```

### Errors
- `400` — missing content and image (`"post content or image is required"`)
- `403` — not an active group member (`"only active group members can post"`)

---

### 10.2) Get Group Posts (paginated)

### `GET /api/groups/{id}/posts?page=1`
**Who can use:** active group members only.

### Query Parameters
- `page` (int, optional, default `1`) — page number

Returns 10 posts per page, ordered by newest first.

### Success (200)
```json
{
  "status": "success",
  "message": "group posts",
  "data": {
    "page": 1,
    "posts": [
      {
        "id": 42,
        "user_id": 3,
        "group_id": 1,
        "content": "Hello everyone!",
        "extra_content": "",
        "created_at_time": "2026-03-06T10:00:00Z",
        "likes_count": 5,
        "has_current_user_liked": true,
        "author_first_name": "Alice",
        "author_last_name": "Johnson",
        "author_profile_picture": "/uploads/abc.jpg"
      }
    ]
  }
}
```

### Errors
- `403` — not an active group member (`"only active group members can view posts"`)

---

### 10.3) Delete Group Post

### `DELETE /api/groups/{id}/posts/{post_id}`
**Who can use:** post author, group owner, or group moderator.

Soft-deletes the post (sets `deleted_at`).

### Success (200)
```json
{
  "status": "success",
  "message": "group post deleted",
  "data": {
    "group_id": 1,
    "post_id": 42
  }
}
```

### Errors
- `403` — not post author, owner, or moderator (`"only post author, group owner, or moderators can delete"`)
- `404` — post not found in this group (`"group post not found"`)

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
