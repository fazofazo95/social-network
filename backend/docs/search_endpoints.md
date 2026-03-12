# Search API Endpoints

This document covers the combined search endpoint for users and groups.

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

## Endpoint Index

- `GET /api/search`

---

## 1) Combined Search (Users + Groups)

### `GET /api/search?q=<term>&limit=<n>`
**Who can use:** any authenticated user.

Searches both users and groups and returns two arrays in one response:
- `users`
- `groups`

### Query Parameters
- `q` (required): search text
- `query` (optional fallback): used only when `q` is empty
- `limit` (optional): positive integer, clamped to max `25`, default `10`

### Visibility and Filtering Rules

User results:
- excludes the logged-in user
- excludes users who have blocked the logged-in user
- can include users that the logged-in user has blocked, with `current_status = "Blocked"`

Group results:
- includes all `public` groups that match
- includes private groups where user is `active`
- includes private groups where user is `invited`

### Relevance Ordering
Both users and groups are ranked by weighted matching in SQL:
- exact match gets highest weight
- prefix match gets medium-high weight
- contains match gets lower weight

Then ties are sorted deterministically.

### Response Fields

`users[]`:
- `id`
- `first_name`
- `last_name`
- `username`
- `profile_picture`
- `current_status`: one of
  - `Follow`
  - `Pending`
  - `Following`
  - `Follow Back`
  - `Blocked`

`groups[]`:
- `id`
- `name`
- `description`
- `group_picture`
- `group_members`
- `visibility`
- `join_mode`
- `current_status`: one of
  - `none`
  - `requested`
  - `invited`
  - `active`

### Success (200)
```json
{
  "status": "success",
  "message": "search results",
  "data": {
    "query": "alex",
    "limit": 5,
    "users": [
      {
        "id": 12,
        "first_name": "Alex",
        "last_name": "King",
        "username": "alexk",
        "profile_picture": "/uploads/profiles/12.jpg",
        "current_status": "Following"
      },
      {
        "id": 44,
        "first_name": "Alexa",
        "last_name": "Moore",
        "username": "alexam",
        "profile_picture": "",
        "current_status": "Follow Back"
      },
      {
        "id": 81,
        "first_name": "Alexis",
        "last_name": "Reed",
        "username": "lexi_r",
        "profile_picture": "",
        "current_status": "Pending"
      }
    ],
    "groups": [
      {
        "id": 3,
        "name": "Alex Devs",
        "description": "Backend and frontend collab",
        "group_picture": "/uploads/groups/3.png",
        "group_members": 27,
        "visibility": "public",
        "join_mode": "request",
        "current_status": "active"
      },
      {
        "id": 9,
        "name": "Private Alex Circle",
        "description": "Invite-only group",
        "group_picture": "",
        "group_members": 8,
        "visibility": "private",
        "join_mode": "invite",
        "current_status": "invited"
      },
      {
        "id": 14,
        "name": "Alex Fans",
        "description": "",
        "group_picture": "",
        "group_members": 102,
        "visibility": "public",
        "join_mode": "auto",
        "current_status": "none"
      }
    ]
  }
}
```

### Success (200, No Matches)
```json
{
  "status": "success",
  "message": "search results",
  "data": {
    "query": "zzzzzz",
    "limit": 10,
    "users": [],
    "groups": []
  }
}
```

### Error (400, missing query)
```json
{
  "status": "error",
  "message": "q query parameter is required"
}
```

### Error (400, bad limit)
```json
{
  "status": "error",
  "message": "limit must be a positive integer"
}
```

### Error (401)
```json
{
  "status": "error",
  "message": "unauthorized"
}
```
