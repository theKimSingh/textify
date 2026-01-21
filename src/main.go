package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Server struct {
	db      DB
	rclient RedisClient
}

func main() {

	// TODO: in production, use separate worker binary or env var to distinguish
	// for simplicity, we run worker in same binary here
	if os.Getenv("WORKER") == "1" {
		workerMain()
		return
	}
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
	rclientReal := redis.NewClient(opt)
	defer rclientReal.Close()

	rclient := NewGoRedisClient(rclientReal)
	srv := &Server{db: NewPGXDB(db), rclient: rclient}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"online","service":"ocr-api"}`)
	})

	http.HandleFunc("/textify", srv.handleTextify)
	http.HandleFunc("/results/", srv.handleResults)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := ":" + port
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func (s *Server) handleTextify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		ImageUrl string `json:"imageUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if body.ImageUrl == "" {
		http.Error(w, "imageUrl is required", http.StatusBadRequest)
		return
	}

	jobId := uuid.New().String()
	payload := map[string]string{"imageUrl": body.ImageUrl, "jobId": jobId}
	p, _ := json.Marshal(payload)

	// push to Redis list
	if err := s.rclient.LPush(s.rclient.Context(), "ocr-tasks", p); err != nil {
		http.Error(w, "failed to enqueue task", http.StatusInternalServerError)
		return
	}

	// insert record
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.db.Exec(ctx, "INSERT INTO ocr_results (job_id, image_url, status) VALUES ($1,$2,$3)", jobId, body.ImageUrl, "pending"); err != nil {
		http.Error(w, "failed to insert record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"success": true, "jobId": jobId, "message": "Image queued for OCR processing"})
}

func (s *Server) handleResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// path: /results/{jobId}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Error(w, "job id required", http.StatusBadRequest)
		return
	}
	jobId := parts[2]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := s.db.QueryRow(ctx, "SELECT job_id, image_url, extracted_text, status, created_at FROM ocr_results WHERE job_id=$1 LIMIT 1", jobId)
	var jobID, imageUrl, extractedText, status string
	var createdAt string
	if err := row.Scan(&jobID, &imageUrl, &extractedText, &status, &createdAt); err != nil {
		http.Error(w, "job ID not found", http.StatusNotFound)
		return
	}

	result := map[string]any{
		"jobId":         jobID,
		"imageUrl":      imageUrl,
		"extractedText": extractedText,
		"status":        status,
		"createdAt":     createdAt,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
