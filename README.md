# Textify (Go)

An API gateway to parse PDFs and images via LLM-based OCR models, utilizing a Redis queue and
asynchronous service workers for high-concurrency.

**Prereqs:** Go, PostgreSQL, Redis, Python3 with `paddleocr` and `requests` (for Paddle runner).

**Install Python deps (runner):**

```bash
python3 -m pip install paddleocr requests
```

**Set env and run server:**

```bash
export DATABASE_URL="postgres://user:password@localhost:5432/my_database"
export REDIS_URL="redis://localhost:6379"
export PORT=3000
go run ./src
```

**Run worker using PaddleOCR runner:**

```bash
WORKER=1 OCR_PROVIDER=paddle go run ./src
```

**Fallback (legacy HTTP OCR service):**

```bash
WORKER=1 go run ./src
```

That's it — enqueue with `POST /textify` and fetch results with `GET /results/{jobId}`.

