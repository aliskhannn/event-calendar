# Calendar Service

A small HTTP server for managing calendar events with JWT-based authentication and bcrypt-secured user registration.
The service provides CRUD operations for events, as well as endpoints for querying daily, weekly, and monthly schedules.

---

## Features

* User authentication and registration (`JWT + bcrypt`)
* CRUD operations for calendar events
* Query events by day, week, or month
* Middleware logging of all requests
* PostgreSQL persistence with migrations (via `goose`)
* Configurable via `.env`
* Dockerized for easy setup

---

## Project Structure

```
cmd
|-- server          # Application entrypoint (main.go)
config              # Application config (YAML)
internal
|-- api
│   └── handlers    # HTTP handlers (auth, event)
│   └── response    # Unified JSON response helpers
│   └── router      # HTTP routes
│   └── server      # HTTP server
|-- config          # Config loader
|-- logger          # Logger setup (zap)
|-- middlewares     # Middleware (auth, logging)
|-- model           # Domain models (User, Event, etc.)
|-- repository      # Data access layer
|-- service         # Business logic layer
|-- migrations      # SQL migrations
|-- docker-compose.yml
|-- Dockerfile
|-- Makefile
|-- README.md
```

---

## API Endpoints

### Public routes

#### `POST /api/user/register`

Register a new user.
**Request (JSON):**

```json
{
  "name": "Alice",
  "email": "alice@example.com",
  "password": "strongpassword"
}
```

#### `POST /api/user/login`

Authenticate and receive a JWT token.
**Request (JSON):**

```json
{
  "email": "alice@example.com",
  "password": "strongpassword"
}
```

**Response:**

```json
{
  "token": "Bearer <jwt_token>"
}
```

---

### Protected routes (require `Authorization: Bearer <token>` header)

#### `POST /api/events/`

Create an event.
**Request (JSON):**

```json
{
  "title": "New Year Party",
  "description": "Celebration with friends",
  "event_date": "2023-12-30T00:00:00Z"
}
```

#### `PUT /api/events/{id}`

Update an existing event.
**Path parameter:**

* `id` — Event UUID

**Request (JSON):**

```json
{
  "title": "Updated Party Title",
  "description": "Changed description",
  "event_date": "2023-12-30T00:00:00Z"
}
```

#### `DELETE /api/events/{id}`

Delete an event by ID.
**Path parameter:**

* `id` — Event UUID

---

### Event Queries

#### `GET /api/events/day?date=2023-12-31`

Get events for a specific day.

#### `GET /api/events/week?date=2023-12-31`

Get events for the week containing the given date.

#### `GET /api/events/month?date=2023-12-31`

Get events for the month containing the given date.

**Query parameters:**

* `date` — required, `YYYY-MM-DD`

---

## 📦 Installation & Setup

### 1. Clone repository

```bash
git clone https://github.com/aliskhannn/calendar-service.git
cd calendar-service
```

### 2. Configure environment

Copy `.env.example` to `.env` and adjust values:

```bash
cp .env.example .env
```

### 3. Run with Docker

```bash
make docker-up
```

To stop and remove containers:

```bash
make docker-down
```

### 4. Run tests

```bash
make test
```

### 5. Lint & format

```bash
make lint
make format
```

---

## ⚙️ Tech Stack

* **Go** — backend implementation
* **Chi** — HTTP router
* **Zap** — structured logging
* **JWT** — authentication
* **bcrypt** — password hashing
* **PostgreSQL** — database
* **Goose** — migrations
* **Docker & docker-compose** — environment setup

---

## 📑 API Responses

* **200 OK** — success
* **201 Created** — resource created
* **400 Bad Request** — invalid input
* **401 Unauthorized** — missing/invalid token
* **403 Forbidden** — not allowed
* **404 Not Found** — resource not found
* **409 Conflict** — already exists
* **500 Internal Server Error** — unexpected error
* **503 Service Unavailable** — business logic error (e.g. user not found)