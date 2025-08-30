# Simple Calendar HTTP Server

A lightweight HTTP server for managing calendar events.  
Events are stored in memory. The server provides CRUD operations and daily/weekly/monthly views.

---

## Server Setup

Run the server on the port specified in the environment variable:

```bash
export PORT=8080
go run main.go
````

Default port is `8080`. All responses are in JSON format.

---

## Middleware

* **Logger:** Logs every request method, URL, and duration to stdout.

---

## API Endpoints

| Method | Endpoint                 | Description                            | Request Body / Query Parameters                                                          |
| ------ | ------------------------ | -------------------------------------- | ---------------------------------------------------------------------------------------- |
| POST   | /api/user/register       | Register a new user                    | JSON: `{"email":"user@example.com","name":"Alice","password":"secret"}`                  |
| POST   | /api/user/login          | Login and receive JWT token            | JSON: `{"email":"user@example.com","password":"secret"}`                                 |
| POST   | /api/events/             | Create a new event                     | JSON: `{"title":"Meeting","description":"Project update","event_date":"2025-09-01"}`     |
| PUT    | /api/events/{id}         | Update an event by ID                  | JSON: `{"title":"Updated title","description":"Updated desc","event_date":"2025-09-02"}` |
| DELETE | /api/events/{id}         | Delete an event by ID                  | -                                                                                        |
| GET    | /api/events/day/{date}   | Get events for a specific day          | Path param: `date` (YYYY-MM-DD)                                                          |
| GET    | /api/events/week/{date}  | Get events for a week containing date  | Path param: `date` (YYYY-MM-DD)                                                          |
| GET    | /api/events/month/{date} | Get events for a month containing date | Path param: `date` (YYYY-MM-DD)                                                          |

**Protected routes** (events endpoints) require an `Authorization: Bearer <token>` header.

---

## Response Format

* Success:

```json
{
  "result": "some message or list of events"
}
```

* Error:

```json
{
  "error": "error description"
}
```

* HTTP status codes:

    * `200 OK` — successful requests
    * `400 Bad Request` — invalid input
    * `401 Unauthorized` — missing/invalid token
    * `503 Service Unavailable` — business logic error
    * `500 Internal Server Error` — unexpected server error

---

## Example Requests

**Create Event**

```bash
curl -X POST http://localhost:8080/api/events/ \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <token>" \
-d '{"title":"Meeting","description":"Project update","event_date":"2025-09-01"}'
```

**Get Events for a Day**

```bash
curl -X GET http://localhost:8080/api/events/day/2025-09-01 \
-H "Authorization: Bearer <token>"
```
