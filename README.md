# Textify

A **production-ready OCR processing pipeline** built with **Elysia**, **PaddleOCR**, and a reliable background job system. This setup is designed for scalability, fault tolerance, and clean persistence.

---

## Tech Stack

* **Runtime:** [Bun](https://bun.sh)
* **Framework:** [ElysiaJS](https://elysiajs.com/)
* **Queue:** [BullMQ](https://docs.bullmq.io/) + Redis
* **ORM:** [Drizzle ORM](https://orm.drizzle.team/)
* **Database:** PostgreSQL
* **OCR Engine:** [PaddleOCR](https://github.com/PaddlePaddle/PaddleOCR)

---

## Quick Start

### 1. Clone & Install Dependencies

```bash
# Install dependencies
bun install
```

---

### 2. Environment Configuration

Create a `.env` file in the project root:

```env
DATABASE_URL="postgres://postgres:password@localhost:5432/ocr_db"
REDIS_URL="redis://localhost:6379"
PADDLE_OCR_URL="http://localhost:8866/predict/ocr_system"
```

---

### 3. Spin Up Infrastructure (Docker)

If you don’t already have PostgreSQL, Redis, or PaddleOCR running locally:

```bash
# Redis
docker run -d --name redis-ocr -p 6379:6379 redis

# PostgreSQL
docker run -d --name postgres-ocr \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=ocr_db \
  -p 5432:5432 postgres

# PaddleOCR (CPU version)
docker run -d --name paddle-ocr \
  -p 8866:8866 paddlepaddle/paddleocr:latest-cpu-2.0-serv
```

---

### 4. Database Migration

Sync your Drizzle schema with PostgreSQL:

```bash
bunx drizzle-kit push
```

---

### 5. Start the Application

```bash
bun run index.ts
```

---

## API Documentation

### Queue an Image

Submit an image URL to be processed asynchronously.

**Health Check Endpoint**

```
GET /
```

**Main Endpoint**

```
POST /textify
```

**Request Body**

```json
{
  "imageUrl": "https://example.com/document.png"
}
```

---

### Check Status / Retrieve Results

Fetch OCR results using the returned `jobId`.

**Endpoint**

```
GET /results/:jobId
```

**Response Example**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "jobId": "12",
  "imageUrl": "https://example.com/document.png",
  "extractedText": "The quick brown fox...",
  "status": "completed",
  "createdAt": "2026-01-26T10:00:00Z"
}
```

---

## 📁 Project Structure

```
.
├── db.ts              # Database connection singleton
├── schema.ts          # Drizzle PostgreSQL schema
├── index.ts           # Elysia API + BullMQ worker
├── drizzle.config.ts  # Drizzle migration configuration
├── .env               # Environment variables
└── package.json       # Dependencies & scripts
```

---

## Notes

* OCR jobs are processed **asynchronously** via BullMQ workers.
* Failed jobs can be retried automatically using BullMQ retry strategies.
* PaddleOCR runs as a separate service and can be scaled independently.

---
