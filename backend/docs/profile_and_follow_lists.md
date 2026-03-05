# Profile & Follow Lists API Documentation

## 1) General
All responses follow the same envelope:

```json
{
  "status": "success | error",
  "message": "descriptive message",
  "data": {}
}
```

**Base URL:** `http://your-api-domain.com/api`

All endpoints below require authentication (`session_id` cookie).

### Quick Reference: Follow Actions

- `POST /users/{id}/follow` — Follow user (creates `pending` or `accepted` based on target privacy)
- `DELETE /users/{id}/unfollow` — Remove your outgoing follow relationship to `{id}`
- `DELETE /users/{id}/remove-follower` — Remove `{id}` from your incoming followers (accepted only)
- `POST /users/{id}/follow/accept` — Accept pending follow request from `{id}`
- `DELETE /users/{id}/follow/reject` — Reject pending follow request from `{id}` (removes row)
- `POST /users/{id}/block` — Block `{id}`
- `DELETE /users/{id}/unblock` — Unblock `{id}`

---

## 2) Get Profile

### Endpoint
`GET /users/{id}`

### Purpose
Returns a profile view for the target user, shaped by:
- whether this is your own profile,
- relationship status (`Following`, `Pending`, `Follow Back`, `Follow`, `Blocked`, `You_Are_Blocked`),
- target privacy (`profile_type`: public/private),
- visibility settings.

### Response fields (may vary by case)
- always possible metadata: `id`, `own_profile`, `current_status`
- basic profile fields: `first_name`, `last_name`, `profile_picture`
- optional profile fields (visibility-controlled):
  - `birthday_date`
  - `relationship_status`
  - `employed_at`
  - `phone_number`
  - `nickname`
  - `about_me`
- follow-related fields:
  - `followers` (int)
  - `following` (int)
  - `follow_vis` (`"visible" | "hidden"`)

### Case Rules

#### A) Own profile (`own_profile: true`)
- Relationship checks are skipped.
- Visibility rules apply to optional fields.
- `follow_vis` is always returned as `"visible"`.

#### B) Target has blocked viewer
- Return blocked status only, no profile info.
- `current_status` is `"You_Are_Blocked"`.

#### C) Target is private and viewer is **not** Following
- Return only basic fields + status metadata.
- `follow_vis` is always forced to `"hidden"` in this basic-only view.

#### D) Target is public OR private+Following
- Return full profile view (except internal fields like `pictures`, `profile_type`, `level`).
- Optional fields are returned only when matching visibility setting is enabled.
- `follow_vis` is returned from DB setting (`visible/hidden`).

### Status mapping
- `Following` => viewer has `accepted` row to target
- `Pending` => viewer has `pending` row to target
- `Follow Back` => target follows viewer (`accepted`) but viewer is not following target
- `Follow` => no active outgoing/pending accepted relationship
- `Blocked` => viewer has blocked target
- `You_Are_Blocked` => target has blocked viewer

---

## 3) Profile Response Examples

### 3.1 Own profile (example)
```json
{
  "status": "success",
  "message": "profile",
  "data": {
    "id": 1,
    "own_profile": true,
    "first_name": "Alice",
    "last_name": "Anderson",
    "profile_picture": "alice.jpg",
    "birthday_date": "1992-03-14",
    "relationship_status": "single",
    "employed_at": "Contoso",
    "phone_number": "123456789",
    "nickname": "ally",
    "about_me": "Hello",
    "followers": 3,
    "following": 2,
    "follow_vis": "visible"
  }
}
```

### 3.2 Private profile + not following (basic-only)
```json
{
  "status": "success",
  "message": "profile",
  "data": {
    "id": 5,
    "own_profile": false,
    "current_status": "Follow Back",
    "first_name": "Eve",
    "last_name": "Edwards",
    "profile_picture": "eve.jpg",
    "followers": 1,
    "following": 1,
    "follow_vis": "hidden"
  }
}
```

### 3.3 Public or private+Following (full view)
```json
{
  "status": "success",
  "message": "profile",
  "data": {
    "id": 2,
    "own_profile": false,
    "current_status": "Following",
    "first_name": "Bob",
    "last_name": "Baker",
    "profile_picture": "bob.jpg",
    "birthday_date": "1988-07-02",
    "relationship_status": "",
    "employed_at": "",
    "nickname": "",
    "about_me": "",
    "followers": 1,
    "following": 1,
    "follow_vis": "visible"
  }
}
```

### 3.4 Blocked by target
```json
{
  "status": "success",
  "message": "profile",
  "data": {
    "id": 4,
    "own_profile": false,
    "current_status": "You_Are_Blocked"
  }
}
```

### Common errors

#### 401 Unauthorized
```json
{
  "status": "error",
  "message": "unauthorized"
}
```

#### 404 Not Found
```json
{
  "status": "error",
  "message": "user not found"
}
```

---

## 4) Follow List Endpoints

These endpoints return lightweight user entries:
- `id`
- `first_name`
- `last_name`
- `profile_picture`

Visibility note:
- Backend returns the list data.
- Frontend can use profile `follow_vis` (from profile/settings responses) to decide whether and how to display these lists.

### 4.1 Get Following
`GET /users/following`

Returns users that **current user follows** with `accepted` status.

Also available for target user:
- `GET /users/{id}/following`

When viewing another user's list (`{id}` is not your own user id), each list item also includes:
- `current_status` (`Following`, `Pending`, `Follow Back`, `Follow`, `Blocked`, `You_Are_Blocked`)

**Own List Example** (`GET /users/following`)
```json
{
  "status": "success",
  "message": "following list",
  "data": [
    {
      "id": 2,
      "first_name": "Bob",
      "last_name": "Baker",
      "profile_picture": "bob.jpg"
    }
  ]
}
```

**Target User Example** (`GET /users/{id}/following`)
```json
{
  "status": "success",
  "message": "following list",
  "data": [
    {
      "id": 1,
      "first_name": "Alice",
      "last_name": "Anderson",
      "profile_picture": "alice.jpg",
      "current_status": "Follow"
    }
  ]
}
```

### 4.2 Get Followers
`GET /users/followers`

Returns users that **follow current user** with `accepted` status.

Also available for target user:
- `GET /users/{id}/followers`

When viewing another user's list (`{id}` is not your own user id), each list item also includes:
- `current_status` (`Following`, `Pending`, `Follow Back`, `Follow`, `Blocked`, `You_Are_Blocked`)

**Own List Example** (`GET /users/followers`)
```json
{
  "status": "success",
  "message": "followers list",
  "data": [
    {
      "id": 5,
      "first_name": "Eve",
      "last_name": "Edwards",
      "profile_picture": "eve.jpg"
    }
  ]
}
```

**Target User Example** (`GET /users/{id}/followers`)
```json
{
  "status": "success",
  "message": "followers list",
  "data": [
    {
      "id": 5,
      "first_name": "Eve",
      "last_name": "Edwards",
      "profile_picture": "eve.jpg",
      "current_status": "Follow Back"
    }
  ]
}
```

### 4.3 Get Pending Incoming Requests
`GET /users/pending`

Returns users whose follow request to current user is `pending`.

**Success Example**
```json
{
  "status": "success",
  "message": "pending requests",
  "data": [
    {
      "id": 6,
      "first_name": "Mallory",
      "last_name": "Mills",
      "profile_picture": "mallory.jpg"
    }
  ]
}
```

### 4.4 Get Blocked Users
`GET /users/blocked`

Returns users that **current user has blocked**.

**Success Example**
```json
{
  "status": "success",
  "message": "blocked users list",
  "data": [
    {
      "id": 4,
      "first_name": "Dave",
      "last_name": "Dawson",
      "profile_picture": "dave.jpg"
    }
  ]
}
```

### Common errors for all list endpoints

#### 401 Unauthorized
```json
{
  "status": "error",
  "message": "unauthorized"
}
```

#### 500 Internal Server Error
```json
{
  "status": "error",
  "message": "failed to get ...: <details>"
}
```

---

## 5) Follow Action Endpoint: Remove Follower

### 5.1 Remove a User from Your Followers
`DELETE /users/{id}/remove-follower`

Purpose:
- Removes `{id}` from the authenticated user's followers list.
- Deletes only an existing `accepted` follow row where:
  - `follower_id = {id}`
  - `followed_id = current_user_id`

Notes:
- This is different from `unfollow`.
  - `unfollow` removes **your outgoing** follow relationship.
  - `remove-follower` removes **incoming follower** relationship.

#### Success (200)
```json
{
  "status": "success",
  "message": "follower removed successfully",
  "data": {
    "user_id": 1,
    "removed_follower_id": 2
  }
}
```

#### Not Found (404)
```json
{
  "status": "error",
  "message": "no accepted follower relationship found"
}
```

#### Unauthorized (401)
```json
{
  "status": "error",
  "message": "unauthorized"
}
```

---

## 7) Follow Action Endpoint: Follow User

### 7.1 Follow a User
`POST /users/{id}/follow`

Purpose:
- Creates a follow relationship from the authenticated user to `{id}`.
- If target user has a public profile, the follow is immediately `accepted`.
- If target user has a private profile, the follow is set to `pending` until the target accepts.

#### Success (201)
```json
{
  "status": "success",
  "message": "follow request created successfully",
  "data": {
    "follower_id": 1,
    "followed_id": 5,
    "status": "accepted"
  }
}
```

Private profile example (pending):
```json
{
  "status": "success",
  "message": "follow request created successfully",
  "data": {
    "follower_id": 1,
    "followed_id": 5,
    "status": "pending"
  }
}
```

#### Errors
- **400 Bad Request:** Invalid target user id.
- **401 Unauthorized:** Missing or invalid session.
- **500 Internal Server Error:** Database or service error.

---

## 8) Follow Action Endpoint: Unfollow User

### 8.1 Unfollow a User
`DELETE /users/{id}/unfollow`

Purpose:
- Removes the follow relationship from the authenticated user to `{id}`.
- Deletes the outgoing follow row regardless of status (`accepted` or `pending`).

#### Success (200)
```json
{
  "status": "success",
  "message": "unfollowed successfully",
  "data": {
    "follower_id": 1,
    "followed_id": 5
  }
}
```

#### Not Found (404)
```json
{
  "status": "error",
  "message": "no follow relationship found"
}
```

#### Unauthorized (401)
```json
{
  "status": "error",
  "message": "unauthorized"
}
```

---

## 9) Follow Action Endpoint: Accept Follow Request

### 9.1 Accept a Pending Incoming Follow Request
`POST /users/{id}/follow/accept`

Purpose:
- Accepts a pending follow request from `{id}` to the authenticated user.
- Updates the follow row status from `pending` to `accepted`.

#### Success (200)
```json
{
  "status": "success",
  "message": "follow request accepted",
  "data": {
    "follower_id": 2,
    "followed_id": 1
  }
}
```

#### Not Found (404)
```json
{
  "status": "error",
  "message": "no pending follow request found"
}
```

#### Unauthorized (401)
```json
{
  "status": "error",
  "message": "unauthorized"
}
```

---

## 10) Follow Action Endpoint: Block User

### 10.1 Block a User
`POST /users/{id}/block`

Purpose:
- Blocks `{id}` from the authenticated user's perspective.
- Removes any existing follow relationships (both directions) between the two users.
- Creates a `blocked` row.

#### Success (200)
```json
{
  "status": "success",
  "message": "user blocked",
  "data": {
    "blocker_id": 1,
    "blocked_id": 5,
    "rows": 1
  }
}
```

#### Errors
- **400 Bad Request:** Invalid target user id.
- **401 Unauthorized:** Missing or invalid session.
- **500 Internal Server Error:** Database or service error.

---

## 11) Follow Action Endpoint: Unblock User

### 11.1 Unblock a User
`DELETE /users/{id}/unblock`

Purpose:
- Removes the block on `{id}` by the authenticated user.
- Deletes the `blocked` row.

#### Success (200)
```json
{
  "status": "success",
  "message": "user unblocked",
  "data": {
    "blocker_id": 1,
    "unblocked_id": 5
  }
}
```

#### Not Found (404)
```json
{
  "status": "error",
  "message": "no blocked relationship found"
}
```

#### Unauthorized (401)
```json
{
  "status": "error",
  "message": "unauthorized"
}
```

---

## 12) Update Profile Media

### 12.1 Update Avatar or Cover Image
`PUT /users/{id}`

Purpose:
- Updates the authenticated user's profile picture and/or cover image.
- The `{id}` path parameter must be `me`.

**Content-Type:** `multipart/form-data`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `avatar` | file | ❌ | New profile picture file. |
| `cover` | file | ❌ | New cover image file. |

At least one file must be provided.

#### Success (200)
```json
{
  "status": "success",
  "message": "profile updated",
  "data": {
    "id": 1,
    "profile_picture": "/uploads/avatars/abc.jpg",
    "cover_image": "/uploads/covers/xyz.jpg"
  }
}
```

#### Errors
- **400 Bad Request:** No avatar or cover file provided, or invalid form.
- **401 Unauthorized:** Missing or invalid session.
- **403 Forbidden:** `{id}` is not `me`.
- **500 Internal Server Error:** Database or file upload error.---

## 6) Follow Action Endpoint: Reject Follow Request

### 6.1 Reject a Pending Incoming Follow Request
`DELETE /users/{id}/follow/reject`

Purpose:
- Rejects pending follow request from `{id}` to authenticated user.
- Deletes only a `pending` row where:
  - `follower_id = {id}`
  - `followed_id = current_user_id`

Notes:
- Request is fully removed so the same user can send a new follow request later.

#### Success (200)
```json
{
  "status": "success",
  "message": "follow request rejected",
  "data": {
    "follower_id": 2,
    "followed_id": 1
  }
}
```

#### Not Found (404)
```json
{
  "status": "error",
  "message": "no pending follow request found"
}
```

#### Unauthorized (401)
```json
{
  "status": "error",
  "message": "unauthorized"
}
```
