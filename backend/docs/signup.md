
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


## 2. Authentication Endpoints

### Register User

`POST /signup`

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