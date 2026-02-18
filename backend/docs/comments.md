## 📌 Table of Contents
* [Comments](#4-comments)
    * [Get Post Comments](#get-post-comments)
    * [Create Comment](#create-comment)
    * [Update Comment](#update-comment)
    * [Delete Comment](#delete-comment)
    * [Restore Comment](#restore-comment)

---

## 4. Comments

### Get Post Comments
`GET /api/posts/{id}/comments`

Retrieves all comments for a specific post. Access depends on the post's privacy settings and the viewer's relationship with the post author.

#### **Security**
* **Cookie:** `session_id` required.

#### **Path Parameters**
| Parameter | Type | Required | Description |
| :--- | :--- | :---: | :--- |
| `id` | integer | ✅ | The ID of the parent post. |

#### **Success Response (200 OK)**
```json
{
    "status": "success",
    "message": "comments retrieved successfully",
    "data": [
        {
            "id": 1,
            "user_id": 5,
            "parent_type": "post",
            "parent_id": 10,
            "content": "Great post!",
            "extra_content": "[https://url-to-comment-image.jpg](https://url-to-comment-image.jpg)",
            "created_at": "2024-05-23 15:00:00",
            "author_first_name": "Jane",
            "author_last_name": "Doe",
            "author_profile_picture": "[https://url-to-avatar.jpg](https://url-to-avatar.jpg)"
        }
    ]
}

```

---

### Create Comment

`POST /api/comments`

Creates a new comment on a post.

**Content-Type:** `multipart/form-data`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `content` | string | ✅ | The text of the comment. |
| `parent_type` | string | ✅ | Usually "post". |
| `parent_id` | integer | ✅ | The ID of the post being commented on. |
| `image` | file | ❌ | Optional image attachment (mapped to `extra_content`). |

---

### Update Comment

`PUT /api/comments/{id}`

Updates the text content of a comment. Only the owner can perform this action.

**Content-Type:** `application/json`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `content` | string | ✅ | The new text content. |

**Responses:**

* **200 OK:** Success.
* **403 Forbidden:** Not the owner of the comment.
* **404 Not Found:** Comment not found.

---

### Delete Comment

`PUT /api/comments/{id}/delete`

Soft deletes a comment. Only the owner can perform this action.

---

### Restore Comment

`PUT /api/comments/{id}/restore`

Restores a soft-deleted comment. Only the owner can perform this action.

---