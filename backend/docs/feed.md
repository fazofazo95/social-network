## 📌 Table of Contents
* [Feed & Discovery](#5-feed--discovery)
    * [Get User Feed](#get-user-feed)
    * [Discover Users](#discover-users)

---

## 5. Feed & Discovery

# Get User Feed
`GET /api/feed`

Retrieves a paginated list of posts from the user's network (including public posts and posts from followed users). 

**Note:** On the first page (`page=1`), the response also includes a list of suggested users to follow.

#### **Security**
* **Cookie:** `session_id` required.

#### **Query Parameters**
| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `page` | integer | `1` | The page number for pagination (10 posts per page). |

#### **Success Response (200 OK)**
```json
{
    "status": "success",
    "message": "Feed loaded",
    "data": {
        "posts": [
            {
                "id": 101,
                "user_id": 2,
                "content": "Hello world",
                "extra_content": "",
                "privacy": "public",
                "created_at": "2024-05-23 10:00:00",
                "author_first_name": "Alice",
                "author_last_name": "Smith",
                "author_profile_picture": "url"
            }
        ],
        "suggestions": [
            {
                "id": 15,
                "first_name": "Bob",
                "last_name": "Jones",
                "profile_picture": "url",
                "status": "Follow"
            }
        ],
        "page": 1
    }
}

```

---

# Discover Users

`GET /api/discover`

Returns a list of 5 random users that the current user is not yet following and has no existing relationship with.

#### **Security**

* **Cookie:** `session_id` required.

#### **Success Response (200 OK)**

```json
{
    "status": "success",
    "message": "Discovered users",
    "data": [
        {
            "id": 20,
            "first_name": "Charlie",
            "last_name": "Brown",
            "profile_picture": "url",
            "status": "Follow"
        }
    ]
}

```

#### **Discovered User Object**

| Field | Type | Description |
| --- | --- | --- |
| `id` | int | Unique user identifier. |
| `first_name` | string | User's first name. |
| `last_name` | string | User's last name. |
| `profile_picture` | string | URL to profile picture. |
| `status` | string | Relationship status: `Follow`, `Follow Back`, `Following`, `Pending`. |

---