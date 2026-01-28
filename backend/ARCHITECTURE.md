# Backend Architecture - Separation of Concerns

## Overview

The backend is now organized into three clear layers:

1. **Handlers** - HTTP request/response handling
2. **Services** - Business logic
3. **Queries (Database)** - Data access layer

## Layer Structure

### 1. Handlers (`pkg/handlers/`)

**Responsibility**: HTTP request/response handling and validation

- Parse incoming HTTP requests
- Validate HTTP-level constraints (method, headers, content-type)
- Convert HTTP input to service request types
- Call service layer methods
- Map service errors to HTTP status codes
- Format and send HTTP responses

**Files**:

- `login.go` - Handles POST /login requests
- `signup.go` - Handles POST /signup requests
- `logout.go` - Handles POST /logout requests

**Example Flow**:

```
HTTP Request → Parse JSON → Create Service Request → Call Service → Handle Response → HTTP Response
```

### 2. Services (`pkg/services/`)

**Responsibility**: Business logic and orchestration

- Validate business rules
- Coordinate database operations
- Manage application state logic
- Handle service-level errors
- Define service contracts and DTOs

**Files**:

- `auth.go` - Authentication business logic (login, signup, logout, sessions)
- `user.go` - User profile management logic

**Key Features**:

- Clean service interfaces
- Business logic validation (passwords, emails, etc.)
- Error mapping and custom error types
- Request/Response DTOs for service layer

**Example Flow**:

```
Service Request → Validate Business Rules → Call Queries → Transform Result → Service Response
```

### 3. Queries (`pkg/db/queries/`)

**Responsibility**: Pure database operations

- Execute SQL queries
- Manage database transactions
- Map SQL errors to application errors
- Handle database-specific constraints
- No business logic - only data access

**Files**:

- `dbLogin.go` - Login and session queries
- `dbSignUp.go` - User registration queries
- `dbUsers.go` - User profile queries
- `dbSession.go` - Session management queries

**Key Principles**:

- Single responsibility per function
- Consistent transaction handling
- Clear input/output contracts
- No HTTP knowledge

## Data Flow

### Login Example

```
┌─────────────────────────────────────────────────────────┐
│ HTTP Handler (login.go)                                 │
│ - Parse JSON request                                    │
│ - Create LoginRequest DTO                               │
│ - Call authService.Login()                              │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│ AuthService (services/auth.go)                          │
│ - Validate input (email/username required, etc)         │
│ - Create LoginInput for queries                         │
│ - Call queries.LogIn()                                  │
│ - Call queries.GenerateSession()                        │
│ - Return LoginResponse with SessionID                   │
└──────────────────────┬──────────────────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        ▼                             ▼
   ┌─────────────┐            ┌──────────────────┐
   │ queries.    │            │ queries.         │
   │ LogIn()     │            │ GenerateSession()│
   │ - SQL exec  │            │ - SQL exec       │
   └─────────────┘            └──────────────────┘
        │                             │
        └──────────────┬──────────────┘
                       ▼
┌─────────────────────────────────────────────────────────┐
│ Database Results                                        │
│ - userID + SessionID returned to Service               │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│ HTTP Handler                                            │
│ - Set session cookie                                    │
│ - Return HTTP 200 with success message                  │
└─────────────────────────────────────────────────────────┘
```

## Error Handling Strategy

### Query Layer Errors

- Database-specific errors (ErrNoRows, constraint violations)
- SQL execution errors
- Transaction errors

### Service Layer Errors

- Business logic errors (ErrInvalidCredentials, ErrEmailTaken)
- Maps query errors to service errors
- Provides meaningful error types

### Handler Layer Errors

- Maps service errors to HTTP status codes
- Returns appropriate HTTP status (401, 409, 500, etc)
- Sends meaningful error messages to client

**Example**:

```go
// Query Layer
queries.ErrNoRows → errors.New("row not found")

// Service Layer
if err == queries.ErrNoRows {
    return ErrInvalidCredentials  // Meaningful service error
}

// Handler Layer
if err == services.ErrInvalidCredentials {
    responses.SendError(w, http.StatusUnauthorized, "invalid username, email, or password")
}
```

## Benefits of This Architecture

✅ **Separation of Concerns** - Each layer has a single responsibility
✅ **Testability** - Easy to test each layer independently (mock services, queries)
✅ **Reusability** - Services can be called from multiple handlers
✅ **Maintainability** - Changes to one layer don't affect others
✅ **Scalability** - Easy to add new services/queries as features grow
✅ **Error Handling** - Clear error propagation and mapping
✅ **Type Safety** - Strong contracts between layers with DTOs

## Adding New Features

### To add a new authentication feature:

1. **Query Layer** - Add database operation in `queries/`
2. **Service Layer** - Add business logic method to `AuthService`
3. **Handler Layer** - Add HTTP handler that calls service method

Example: Add password reset feature

```
queries.ResetPassword() → AuthService.ResetPassword() → ResetPasswordHandler()
```

### To add user profile functionality:

1. **Query Layer** - Add queries in `queries/dbUsers.go`
2. **Service Layer** - Add methods to `UserService` (already done!)
3. **Handler Layer** - Create `profile.go` handler

## Service DTOs (Request/Response Types)

Located in `services/` files for clarity:

**Request DTOs**:

- `SignUpRequest` - Input for signup
- `LoginRequest` - Input for login
- `UserProfileRequest` - Input for profile operations

**Response DTOs**:

- `LoginResponse` - Output with SessionID
- `UserProfileResponse` - User profile output

These DTOs are the contracts between handlers and services.

## Next Steps

1. Create handlers for user profile endpoints using `UserService`
2. Add post/comment service layer for post functionality
3. Add relationship service layer for following/unfollowing
4. Add notification service layer for notifications
5. Implement dependency injection for cleaner service initialization

## Summary

```
User Input (HTTP)
        ↓
   HANDLER (HTTP concerns)
        ↓
   SERVICE (Business logic)
        ↓
   QUERIES (Database operations)
        ↓
   DATABASE
```

Each layer is independent and testable!
