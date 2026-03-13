package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func workerMain() {
	ctx := context.Background()
	db := InitDB(ctx)
	defer db.Close()

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "redis://localhost:6379"
	}
	opt, err := redis.ParseURL(redisAddr)
	if err != nil {
		log.Fatalf("invalid REDIS_URL: %v", err)
	}
	rclient := redis.NewClient(opt)
	defer rclient.Close()

	// worker loop: BLPop with timeout and process messages
	for {
		res, err := rclient.BLPop(ctx, 5*time.Second, "ocr-tasks").Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			log.Printf("redis blpop error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if len(res) < 2 {
			continue
		}
		payloadBytes := []byte(res[1])
		var payload struct {
			ImageUrl string `json:"imageUrl"`
			JobId    string `json:"jobId"`
		}
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			log.Printf("invalid payload: %v", err)
			continue
		}

		// call OCR service
		reqBody, _ := json.Marshal(map[string][]string{"images": {payload.ImageUrl}})
		req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8866/predict/ocr_system", bytes.NewReader(reqBody))
		if err != nil {
			updateStatus(ctx, db, payload.JobId, "failed")
			log.Printf("request create error: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			updateStatus(ctx, db, payload.JobId, "failed")
			log.Printf("request error: %v", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var parsed struct {
			Results [][]struct {
				Text string `json:"text"`
			} `json:"results"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			updateStatus(ctx, db, payload.JobId, "failed")
			log.Printf("parse error: %v", err)
			continue
		}

		var textContent string
		if len(parsed.Results) > 0 {
			parts := make([]string, 0, len(parsed.Results[0]))
			for _, r := range parsed.Results[0] {
				parts = append(parts, r.Text)
			}
			textContent = strings.Join(parts, " ")
		}

		if _, err := db.Exec(ctx, "UPDATE ocr_results SET extracted_text=$1, status='completed' WHERE job_id=$2", textContent, payload.JobId); err != nil {
			log.Printf("db update error: %v", err)
			continue
		}
	}
}

func updateStatus(ctx context.Context, db *pgxpool.Pool, jobId, status string) {
	if jobId == "" {
		return
	}
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, _ = db.Exec(ctx2, "UPDATE ocr_results SET status=$1 WHERE job_id=$2", status, jobId)
}
