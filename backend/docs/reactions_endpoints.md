# Reactions (Likes) API Endpoints

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
* [Like a Post](#1-like-a-post)
* [Unlike a Post](#2-unlike-a-post)

---

## 1) Like a Post

### `POST /api/posts/{id}/like`
**Who can use:** any authenticated user.

Adds a like reaction from the current user to the specified post.

#### **Path Parameters**
| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| `id` | integer | ✅ | The ID of the post to like. |

#### **Success Response (200 OK)**
```json
{
  "status": "success",
  "message": "reaction added successfully",
  "data": {
    "like_count": 5
  }
}
```

| Field | Type | Description |
| --- | --- | --- |
| `like_count` | int | The updated total like count for the post after adding the reaction. |

#### **Errors**
- **401 Unauthorized:** Missing or invalid session.
- **500 Internal Server Error:** Failed to add reaction (e.g., already liked, database error).

---

## 2) Unlike a Post

### `DELETE /api/posts/{id}/like`
**Who can use:** any authenticated user.

Removes the current user's like reaction from the specified post.

#### **Path Parameters**
| Parameter | Type | Required | Description |
| --- | --- | --- | --- |
| `id` | integer | ✅ | The ID of the post to unlike. |

#### **Success Response (200 OK)**
```json
{
  "status": "success",
  "message": "reaction removed successfully",
  "data": {
    "like_count": 4
  }
}
```

| Field | Type | Description |
| --- | --- | --- |
| `like_count` | int | The updated total like count for the post after removing the reaction. |

#### **Errors**
- **401 Unauthorized:** Missing or invalid session.
- **500 Internal Server Error:** Failed to remove reaction (e.g., not previously liked, database error).
