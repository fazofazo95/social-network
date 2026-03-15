# Pulse — Social Network

A full-stack social networking application built with **Go** and **Next.js**. Pulse features real-time chat, group collaboration, follow systems, privacy controls, and more.

## Tech Stack

| Layer      | Technology                            |
| ---------- | ------------------------------------- |
| Backend    | Go 1.24, `net/http`                   |
| Frontend   | Next.js 16, React 19, Tailwind CSS 4  |
| Database   | SQLite 3 (WAL mode)                   |
| Real-time  | WebSocket (chat), SSE (notifications) |
| Auth       | Cookie-based sessions                 |
| Containers | Docker, Docker Compose                |

## Features

- **Authentication** — Register, login, logout with secure cookie-based sessions
- **Posts** — Create, edit, and delete posts with privacy controls (public, private, followers-only)
- **Comments** — Threaded comments on posts with edit and soft-delete support
- **Reactions** — Like posts and comments with real-time count updates
- **Profiles** — User profiles with avatar, cover image, bio, and location
- **Follow System** — Follow/unfollow users, pending request approval, follower/following lists
- **Groups** — Create and join groups, group posts, group events with RSVP
- **Real-time Chat** — Direct messaging via WebSocket with read receipts
- **Notifications** — Real-time notifications via Server-Sent Events (SSE) for follows, group invites, and events
- **Search** — Search for users, groups, and posts
- **Settings** — Profile visibility, account settings, content preferences
- **Image Uploads** — Avatar, cover, and group images (JPEG, PNG, GIF up to 20 MiB)

## Project Structure

```
social-network/
├── docker-compose.yml          # Container orchestration
├── backend/
│   ├── Dockerfile              # Multi-stage Go build
│   ├── main.go                 # Entry point
│   ├── server.go               # Server setup & route registration
│   ├── pkg/
│   │   ├── db/
│   │   │   ├── migrations/     # SQL migration files
│   │   │   └── sqlite/         # DB init & migration runner
│   │   ├── handlers/           # HTTP request handlers
│   │   ├── middleware/         # Auth, CORS middleware
│   │   ├── models/             # Data models
│   │   ├── repository/         # Database access layer
│   │   ├── services/           # Business logic layer
│   │   ├── sse/                # Server-Sent Events hub
│   │   ├── ws/                 # WebSocket hub & handler
│   │   └── utils/              # File upload utilities
│   └── uploads/                # User-uploaded images
├── frontend/
│   ├── Dockerfile              # Multi-stage Next.js build
│   ├── src/
│   │   ├── app/
│   │   │   ├── (auth)/         # Login & register pages
│   │   │   └── (dashboard)/    # Main app pages (feed, profile, groups, messages, settings)
│   │   ├── components/         # Reusable UI components
│   │   └── lib/                # API client, services, utilities
│   └── public/                 # Static assets
└── README.md
```

**Backend architecture:** Handler → Service → Repository pattern with clear separation of concerns.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)

For local development without Docker:

- [Go 1.24+](https://golang.org/dl/) with GCC (required for SQLite CGO bindings)
- [Node.js 22+](https://nodejs.org/) with npm

## Running with Docker

Build and start both containers:

```bash
docker compose up --build
```

The application will be available at:

- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080

Stop the containers:

```bash
docker compose down
```

Stop and remove all data (database + uploads):

```bash
docker compose down -v
```

Rebuild a single service:

```bash
docker compose build backend
docker compose build frontend
```

## Running Locally (without Docker)

### Backend

```bash
cd backend
go run .
```

The server starts on http://localhost:8080. The SQLite database is created automatically at `pkg/db/social_network.db` and migrations run on startup.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The development server starts on http://localhost:3000.

## Environment Variables

### Backend

| Variable      | Default                 | Description                     |
| ------------- | ----------------------- | ------------------------------- |
| `CORS_ORIGIN` | `http://localhost:3000` | Allowed origin for CORS headers |

### Frontend

| Variable                   | Default                 | Description                           |
| -------------------------- | ----------------------- | ------------------------------------- |
| `NEXT_PUBLIC_API_BASE_URL` | `http://localhost:8080` | Backend API URL (baked at build time) |

## Docker Architecture

The project uses two Docker images orchestrated with Docker Compose:

- **Backend container** (`social-network-backend`) — Runs the Go server, handles API requests, manages the SQLite database, and serves uploaded images. Data is persisted via Docker volumes.
- **Frontend container** (`social-network-frontend`) — Serves the Next.js application as a standalone Node.js server. Communicates with the backend via HTTP/WebSocket requests.

### Volumes

| Volume            | Mount Point    | Purpose                          |
| ----------------- | -------------- | -------------------------------- |
| `backend-data`    | `/app/pkg/db`  | SQLite database persistence      |
| `backend-uploads` | `/app/uploads` | User-uploaded images persistence |

## API Overview

All API endpoints are prefixed with `/api/` and use JSON request/response bodies. Authentication is cookie-based with `credentials: "include"`.

| Area          | Endpoints                                                |
| ------------- | -------------------------------------------------------- |
| Auth          | `/api/signup`, `/api/login`, `/api/logout`, `/api/me`    |
| Posts         | `/api/posts`, `/api/posts/{id}`                          |
| Comments      | `/api/comments`, `/api/comments/{id}`                    |
| Reactions     | `/api/reactions`                                         |
| Profiles      | `/api/profile/{id}`                                      |
| Follow        | `/api/follow`, `/api/followers`, `/api/following`        |
| Groups        | `/api/groups`, `/api/groups/{id}`, `/api/groups/active`  |
| Chat          | `/api/chat`, WebSocket at `/ws`                          |
| Notifications | `/api/notifications`, SSE at `/api/notifications/stream` |
| Search        | `/api/search`                                            |
| Settings      | `/api/settings`                                          |
| Feed          | `/api/feed`                                              |
| Uploads       | `/uploads/{filename}` (static file serving)              |
