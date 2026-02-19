# Settings API Documentation (Visibility + Content)

## 1) General
All responses use the standard envelope:

```json
{
  "status": "success | error",
  "message": "descriptive message",
  "data": {}
}
```

**Base URL:** `http://your-api-domain.com/api`

All endpoints below require authentication (`session_id` cookie).

---

## 2) Visibility Settings

Visibility settings control how profile data is exposed in profile views.

### 2.1 Fetch Visibility Settings
`GET /users/settings`

Returns current user's visibility settings plus `profile_type`.

#### Success Example
```json
{
  "status": "success",
  "message": "visibility settings",
  "data": {
    "email_vis": "hidden",
    "birthday_date_vis": "visible",
    "relationship_status_vis": "visible",
    "employed_at_vis": "visible",
    "phone_number_vis": "hidden",
    "about_me_vis": "visible",
    "nickname_vis": "visible",
    "follow_vis": "hidden",
    "profile_type": "public"
  }
}
```

### 2.2 Update Visibility Settings
`PATCH /users/settings`
`PUT /users/settings`

Accepts any subset of fields:
- `email_vis`
- `birthday_date_vis`
- `relationship_status_vis`
- `employed_at_vis`
- `phone_number_vis`
- `about_me_vis`
- `nickname_vis`
- `follow_vis`
- `profile_type`

#### Accepted values
- Visibility fields: `"visible" | "hidden"` or boolean `true | false`
- `profile_type`: `"public" | "private"`

#### Request Example
```json
{
  "email_vis": "visible",
  "phone_number_vis": false,
  "about_me_vis": "hidden",
  "follow_vis": "visible",
  "profile_type": "private"
}
```

#### Success Example
```json
{
  "status": "success",
  "message": "settings updated",
  "data": {
    "email_vis": "visible",
    "birthday_date_vis": "visible",
    "relationship_status_vis": "visible",
    "employed_at_vis": "visible",
    "phone_number_vis": "hidden",
    "about_me_vis": "hidden",
    "nickname_vis": "visible",
    "follow_vis": "visible",
    "profile_type": "private"
  }
}
```

#### Validation Errors
- Invalid JSON:
```json
{
  "status": "error",
  "message": "invalid JSON"
}
```

- Invalid visibility/profile_type value:
```json
{
  "status": "error",
  "message": "invalid follow_vis value"
}
```
(or corresponding field message, e.g. `invalid profile_type value`)

---

## 3) Content Settings

Content settings are editable profile fields shown in the settings content manager.

Fields:
- `first_name`
- `last_name`
- `birthday_date`
- `relationship_status`
- `employed_at`
- `phone_number`
- `nickname`
- `about_me`

### 3.1 Fetch Content Settings
`GET /users/settings/content`

#### Success Example
```json
{
  "status": "success",
  "message": "content settings",
  "data": {
    "first_name": "Alice",
    "last_name": "Anderson",
    "birthday_date": "1992-03-14",
    "relationship_status": "single",
    "employed_at": "Contoso",
    "phone_number": "123456789",
    "nickname": "ally",
    "about_me": "Loves hiking and coffee"
  }
}
```

### 3.2 Update Content Settings
`PATCH /users/settings/content`
`PUT /users/settings/content`

Accepts any subset of content fields.

#### Request Example
```json
{
  "first_name": "AliceUpdated",
  "last_name": "AndersonUpdated",
  "birthday_date": "1992-03-15",
  "relationship_status": "married",
  "employed_at": "Contoso",
  "phone_number": "123456789",
  "nickname": "ally",
  "about_me": "Updated from settings"
}
```

#### Success Example
```json
{
  "status": "success",
  "message": "content settings updated",
  "data": {
    "first_name": "AliceUpdated",
    "last_name": "AndersonUpdated",
    "birthday_date": "1992-03-15",
    "relationship_status": "married",
    "employed_at": "Contoso",
    "phone_number": "123456789",
    "nickname": "ally",
    "about_me": "Updated from settings"
  }
}
```

#### Content Validation Notes
- `birthday_date` must be in `YYYY-MM-DD` format when provided.
- For nullable text fields, sending an empty string clears the value in DB.

#### Validation Error Example
```json
{
  "status": "error",
  "message": "birthday_date must be YYYY-MM-DD"
}
```

---

## 4) Common Errors

### 401 Unauthorized
```json
{
  "status": "error",
  "message": "unauthorized"
}
```

### 500 Internal Server Error
```json
{
  "status": "error",
  "message": "failed to ...: <details>"
}
```
