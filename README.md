# Textify (Go)

A small OCR processing service written in Go. It accepts image URLs, enqueues OCR jobs into Redis, runs a background worker to call an OCR service, and persists results to PostgreSQL.

---

## Requirements

- Go 1.21+
- Redis
- PostgreSQL
- An OCR HTTP endpoint (the project calls `http://localhost:8866/predict/ocr_system` by default)

---

## Quick Start (Local)

1. Set environment variables (example):

```bash
export DATABASE_URL="postgres://user:password@localhost:5432/my_database"
export REDIS_URL="redis://localhost:6379"
export PORT=3000
```

2. Build and run the server:

```bash
# Build
go build -o textify ./src

# Run server
./textify

# Or run directly
go run ./src
```

3. Run the worker (separate terminal):

```bash
WORKER=1 go run ./src
```

---

## Docker (Compose)

The repository includes `docker_compose.yml` that starts `server`, `worker`, `db`, `redis`, and `minio` services. To start everything:

```bash
docker compose up --build
```

The API will be available on port 3000.

---

## API

- Health: `GET /`
- Queue an image: `POST /textify` with JSON body `{ "imageUrl": "https://..." }`
- Get result: `GET /results/:jobId`

Example `curl` to enqueue:

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"imageUrl":"https://example.com/doc.png"}' \
  http://localhost:3000/textify
```

---

## Tests

Run unit tests:

```bash
go test ./...
```

---

## Schema

The app automatically ensures a simple `ocr_results` table exists with the following columns:
- `id` (uuid), `job_id`, `image_url`, `extracted_text`, `status`, `created_at`.

---

## Project layout

```
.
├── Dockerfile
├── docker_compose.yml
├── go.mod
├── src/
│   ├── main.go       # HTTP server + worker mode switch
│   ├── worker.go     # worker loop (BLPOP)
│   ├── db.go         # DB init & schema
│   ├── redis_adapter.go
│   ├── db_adapter.go
│   └── main_test.go  # unit tests
```

