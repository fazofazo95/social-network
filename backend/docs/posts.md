# API Documentation

## 📌 Table of Contents
* [Create Post](#create-post)
* [Get Single Post](#get-single-post)
* [Get User's Posts](#get-users-posts)
* [Update Post](#update-post)
* [Delete Post](#delete-post)
* [Restore Post](#restore-post)
---
# Create Post
`POST /api/posts`

Creates a new post. Supports file uploads for images and custom privacy settings.

**Content-Type:** `multipart/form-data`

| Field | Type | Required | Description |
| :--- | :--- | :---: | :--- |
| `content` | string | ✅ | The main text of the post. |
| `privacy` | string | ✅ | `public`, `followers`, or `custom`. |
| `image` | file | ❌ | Optional image for the post. |
| `whitelisted_users` | int[] | ❌ | Array of user IDs (Required only if privacy is `custom`). |

**Example Response (201 Created):**
```json
{
    "status": "success",
    "message": "user created successfully",
    "data": null
}

```

---

# Update Post

`PUT /api/posts/{id}`

Updates the content of an existing post. Only the owner can perform this action.

**Content-Type:** `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `content` | string | ✅ | The new content of the post. |

**Responses:**

* **200 OK:** Post updated successfully.
* **403 Forbidden:** You are not the owner of this post.
* **404 Not Found:** Post not found.

---

# Delete Post

`PUT /api/posts/{id}/delete`

Soft deletes a post. Only the owner can perform this action.

**Response (200 OK):**

```json
{
    "status": "success",
    "message": "post deleted successfully",
    "data": null
}

```

---

# Restore Post

`PUT /api/posts/{id}/restore`

Restores a previously deleted post. Only the owner can perform this action.

**Response (200 OK):**

```json
{
    "status": "success",
    "message": "post restored successfully",
    "data": null
}

```
---

# Get Single Post
`GET /posts/{id}`

Retrieves detailed information about a specific post by its ID. 

#### **Security**
This endpoint is protected by session-based authentication.

| Type | Name | Description |
| :--- | :--- | :--- |
| **Cookie** | `session_id` | Valid session identifier required to view the post. |

#### **Path Parameters**
| Parameter | Type    | Required | Description                |
| :-------- | :------ | :------: | :------------------------- |
| `id`      | integer | ✅       | The unique ID of the post. |

---

#### **Responses**

**✅ 200 OK**
Post retrieved successfully. The `data` field contains the full Post object.

```json
{
    "status": "success",
    "message": "post retrieved successfully",
    "data": {
        "id": 10,
        "user_id": 5,
        "content": "This is a single post content",
        "extra_content": "",
        "privacy": "public",
        "created_at": "2024-05-23 12:00:00",
        "author_first_name": "John",
        "author_last_name": "Doe",
        "author_profile_picture": "/uploads/test.jpg"
    }
}

```

#### **Post Object Fields**

| Field | Type | Description |
| --- | --- | --- |
| `id` | int | Unique post identifier. |
| `user_id` | int | ID of the author. |
| `content` | string | Main text content. |
| `extra_content` | string | Additional text (returns empty string if null). |
| `image` | string | URL to post image (omitted if empty). |
| `privacy` | string | `public`, `followers`, or `custom`. |
| `created_at` | string | Formatted creation timestamp. |
| `author_first_name` | string | Author's first name. |
| `author_last_name` | string | Author's last name. |
| `author_profile_picture` | string | URL to author's profile picture. |

---

**❌ 400 Bad Request**
The provided post ID is not a valid integer.

```json
{
    "status": "error",
    "message": "Invalid post ID"
}

```

**❌ 401 Unauthorized**
The `session_id` cookie is missing or invalid.

```json
{
    "status": "error",
    "message": "unauthorized"
}

```

**❌ 500 Internal Server Error**
Database error or the post does not exist (if the service returns an error).

```json
{
    "status": "error",
    "message": "internal server error message"
}

```

---

# Get User's Posts
`GET /users/{id}/posts`

Retrieves a list of posts for a specific user. The results are filtered based on the privacy settings of each post and the relationship between the viewer and the author.

#### **Security**
This endpoint is protected by session-based authentication.

| Type | Name | Description |
| :--- | :--- | :--- |
| **Cookie** | `session_id` | Required to identify the viewer for privacy filtering. |

#### **Path Parameters**
| Parameter | Type    | Required | Description                      |
| :-------- | :------ | :------: | :------------------------------- |
| `id`      | integer | ✅       | The unique ID of the post author. |

---

#### **Privacy Logic**
The API automatically filters posts based on the following rules:
* **Public:** Visible to everyone.
* **Followers:** Visible only if the viewer follows the author (accepted status).
* **Custom:** Visible only if the viewer is explicitly whitelisted for that post.
* **Owner:** The author can always see all their own posts.

---

#### **Responses**

**✅ 200 OK**
Returns an array of Post objects.

```json
{
    "status": "success",
    "message": "user posts retrieved successfully",
    "data": [
        {
            "id": 42,
            "user_id": 5,
            "content": "Exploring the mountains!",
            "extra_content": "Check out the photos below",
            "privacy": "public",
            "created_at": "2024-05-23 14:20:01",
            "author_first_name": "John",
            "author_last_name": "Doe",
            "author_profile_picture": "/uploads/test.jpg"
        }
    ]
}

```

#### **Post Object Fields**

| Field | Type | Description |
| --- | --- | --- |
| `id` | int | Unique post identifier. |
| `user_id` | int | ID of the author. |
| `content` | string | Main text content. |
| `extra_content` | string | Additional text content (returns empty string if null). |
| `image` | string | URL to the post image (optional). |
| `privacy` | string | `public`, `followers`, or `custom`. |
| `created_at` | string | Formatted creation timestamp. |
| `author_first_name` | string | Author's first name. |
| `author_last_name` | string | Author's last name. |
| `author_profile_picture` | string | URL to author's profile picture. |

---

**❌ 401 Unauthorized**
Missing or invalid `session_id` cookie.

```json
{
    "status": "error",
    "message": "unauthorized"
}

```

**❌ 400 Bad Request**
Invalid user ID format in path.

```json
{
    "status": "error",
    "message": "Invalid post ID"
}

```