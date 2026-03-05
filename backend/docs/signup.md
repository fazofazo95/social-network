
# API Documentation

## 1. General Response Format
All API responses return a standardized JSON object based on the internal Go `responses` package structure.

**Base URL:** `http://your-api-domain.com/api`

### Success Response Structure
```json
{
    "status": "success",
    "message": "Descriptive success message",
    "data": { ... } // Can be an object, array, or null
}

```

### Error Response Structure

```json
{
    "status": "error",
    "message": "Specific error message explaining the failure"
}

```


## 📌 Table of Contents
* [Register User](#register-user)
* [Login](#login)
* [Logout](#logout)
* [Verify Session](#verify-session)

---

## 2. Authentication Endpoints

### Register User

`POST /api/users`

Creates a new user account. This endpoint processes data as `multipart/form-data` to support profile picture uploads.

#### **Request Headers**

| Key | Value |
| --- | --- |
| `Content-Type` | `multipart/form-data` |

#### **Request Body (Form Data)**

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `email` | string | ✅ | User's email address (must be unique). |
| `password` | string | ✅ | User's password. |
| `firstname` | string | ✅ | User's first name. |
| `lastname` | string | ✅ | User's last name. |
| `username` | string | ✅ | Unique username for the profile. |
| `date_of_birth` | string | ✅ | Format: `DD/MM/YYYY`. |
| `nickname` | string | ❌ | Optional display name. |
| `about_me` | string | ❌ | Optional short biography. |
| `avatar` | file | ❌ | Optional profile image file (e.g., .jpg, .png). |

---

#### **Responses**

**✅ 201 Created**
User was successfully registered.

```json
{
    "status": "success",
    "message": "user created successfully",
    "data": null
}

```

**❌ 400 Bad Request**
The form data is invalid or the multipart form could not be parsed.

```json
{
    "status": "error",
    "message": "Invalid Form"
}

```

**❌ 409 Conflict**
The email or username is already registered in the system.

```json
{
    "status": "error",
    "message": "email already in use"
}
// OR
{
    "status": "error",
    "message": "username already in use"
}

```

**❌ 500 Internal Server Error**
An unexpected server error occurred (e.g., database failure or file upload issue).

```json
{
    "status": "error",
    "message": "detailed error message from server"
}

```

---

### Login

`POST /api/login`

Authenticates a user and creates a session. On success, a `session_id` cookie is set.

#### **Request Headers**

| Key | Value |
| --- | --- |
| `Content-Type` | `application/json` |

#### **Request Body**
```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `email` | string | ✅ | User's email address. |
| `password` | string | ✅ | User's password. |

---

#### **Responses**

**✅ 200 OK**
Login successful. A `session_id` cookie is set on the response.

```json
{
    "status": "success",
    "message": "login successful",
    "data": null
}
```

**❌ 400 Bad Request**
The request body could not be parsed.

```json
{
    "status": "error",
    "message": "invalid request body"
}
```

**❌ 401 Unauthorized**
Invalid email or password.

```json
{
    "status": "error",
    "message": "invalid username, email, or password"
}
```

**❌ 500 Internal Server Error**
Session creation failed.

```json
{
    "status": "error",
    "message": "failed to create session"
}
```

---

### Logout

`DELETE /api/logout`

Logs out the current user by invalidating the session and clearing the `session_id` cookie.

#### **Security**
* **Cookie:** `session_id` required.

---

#### **Responses**

**✅ 200 OK**
Logout successful. The `session_id` cookie is cleared.

```json
{
    "status": "success",
    "message": "logout successful",
    "data": null
}
```

**❌ 401 Unauthorized**
Missing or invalid session cookie.

```json
{
    "status": "error",
    "message": "unauthorized"
}
```

**❌ 500 Internal Server Error**

```json
{
    "status": "error",
    "message": "failed to logout"
}
```

---

### Verify Session

`GET /api/verify-session`

Checks whether the current `session_id` cookie represents a valid, active session. Useful for frontend to verify auth state on page load.

---

#### **Responses**

**✅ 200 OK**
Session is valid. Returns the authenticated user's ID.

```json
{
    "status": "success",
    "message": "session exists",
    "data": 1
}
```

**❌ 401 Unauthorized**
Missing or invalid session cookie.

```json
{
    "status": "error",
    "message": "unauthorized"
}
```