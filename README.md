## Calendar Service

A small HTTP server for managing calendar events with JWT-based authentication and bcrypt-secured user registration.
The service provides CRUD operations for events, daily/weekly/monthly queries, email reminders, and background archiving.

---

## Features

* User authentication and registration (`JWT + bcrypt`)
* CRUD operations for calendar events
* Query events by day, week, or month
* **Email reminders** via background worker
* **Automatic archiving** of old events every configurable interval
* Middleware logging of all requests (**asynchronous logger**)
* PostgreSQL persistence with migrations (via `goose`)
* Configurable via `.env`
* Dockerized for easy setup

---

## Project Structure

```
.
├── cmd                     
│   └── server              
│       └── main.go          # Application entrypoint
├── config                   # Application config (YAML)
├── internal                
│   ├── api                 
│   │   ├── handlers         # HTTP handlers (auth, event)
│   │   ├── response         # Unified JSON response helpers
│   │   ├── router           # HTTP routes
│   │   └── server           # HTTP server
│   ├── config               # Config loader
│   ├── logger               # Logger setup (zap)
│   ├── middlewares          # Middleware (auth, logging)
│   ├── model                # Domain models (User, Event, Reminder, etc.)
│   ├── repository           # Data access layer
│   ├── service              # Business logic layer
│   └── worker               # Background workers
│       ├── archiver         # Archiving old events periodically
│       └── reminder         # Sending event reminders via email
├── migrations               # SQL migrations
├── go.mod                   
├── go.sum                   
├── Dockerfile               
├── docker-compose.yml       
├── Makefile                 
└── README.md                
```

---

## API Endpoints

### Public routes

#### `POST /api/user/register`

Register a new user.

#### `POST /api/user/login`

Authenticate and receive a JWT token.

---

### Protected routes (require `Authorization: Bearer <token>`)

#### `POST /api/events/`

Create an event (optionally with `reminder_at` to schedule an email reminder).

#### `PUT /api/events/{id}`

Update an existing event.

#### `DELETE /api/events/{id}`

Delete an event by ID.

#### Event Queries

* `GET /api/events/day?date=YYYY-MM-DD`
* `GET /api/events/week?date=YYYY-MM-DD`
* `GET /api/events/month?date=YYYY-MM-DD`

---

## Background Workers

### Reminder Worker

* Listens to a channel of `Reminder` tasks.
* Sends an email notification at the scheduled time.

### Archiver Worker

* Runs periodically (configurable interval).
* Moves old events to an archive table to keep the main events table clean.

### Async Logger

* HTTP handlers no longer write to stdout directly.
* Logs are pushed into a channel and written by a separate goroutine.

---

## Installation & Setup

### 1. Clone repository

```bash
git clone https://github.com/aliskhannn/calendar-service.git
cd calendar-service
```

### 2. Configure environment

Copy `.env.example` to `.env` and set values:

```bash
cp .env.example .env
```

**Notes:**

* SMTP credentials: create an account on Mailtrap (or any SMTP provider) and copy SMTP host, port, username, password, and sender email into `.env`.
* JWT secret: set a long random string.
* Database credentials: set host, port, username, password, database name.

### 3. Run with Docker

```bash
make docker-up
```

Stop and remove containers:

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

## Tech Stack

* **Go** — backend implementation
* **Chi** — HTTP router
* **Zap** — structured logging
* **JWT** — authentication
* **bcrypt** — password hashing
* **PostgreSQL** — database
* **Goose** — migrations
* **Docker & docker-compose** — environment setup

---

## API Responses

* **200 OK** — success
* **201 Created** — resource created
* **400 Bad Request** — invalid input
* **401 Unauthorized** — missing/invalid token
* **403 Forbidden** — not allowed
* **404 Not Found** — resource not found
* **409 Conflict** — already exists
* **500 Internal Server Error** — unexpected error
* **503 Service Unavailable** — business logic error (e.g. user not found)