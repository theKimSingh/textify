package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mock DB and Redis for testing handlers without external dependencies
type fakeDB struct {
	lastInsert struct {
		jobId    string
		imageUrl string
		status   string
	}
	queryRowData []any
}

func (f *fakeDB) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	if len(args) >= 3 {
		if v, ok := args[0].(string); ok {
			f.lastInsert.jobId = v
		}
		if v, ok := args[1].(string); ok {
			f.lastInsert.imageUrl = v
		}
		if v, ok := args[2].(string); ok {
			f.lastInsert.status = v
		}
	}
	return 1, nil
}

func (f *fakeDB) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return &fakeRow{data: f.queryRowData}
}

type fakeRow struct{ data []any }

func (r *fakeRow) Scan(dest ...any) error {
	for i := range dest {
		switch d := dest[i].(type) {
		case *string:
			*d = r.data[i].(string)
		}
	}
	return nil
}

type fakeRedis struct {
	pushed [][]byte
	blpop  chan []string
}

func (f *fakeRedis) LPush(ctx context.Context, key string, values ...interface{}) error {
	for _, v := range values {
		if b, ok := v.([]byte); ok {
			f.pushed = append(f.pushed, b)
		}
	}
	return nil
}

func (f *fakeRedis) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	select {
	case v := <-f.blpop:
		return v, nil
	case <-time.After(timeout):
		return nil, nil
	}
}
func (f *fakeRedis) Close() error             { return nil }
func (f *fakeRedis) Context() context.Context { return context.Background() }

func TestHandleTextify(t *testing.T) {
	db := &fakeDB{}
	r := &fakeRedis{pushed: make([][]byte, 0)}
	srv := &Server{db: db, rclient: r}

	body := `{"imageUrl":"http://example.com/image.jpg"}`
	req := httptest.NewRequest("POST", "/textify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleTextify(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 got %d", res.StatusCode)
	}

	if len(r.pushed) != 1 {
		t.Fatalf("expected pushed 1 got %d", len(r.pushed))
	}
	var payload map[string]string
	_ = json.Unmarshal(r.pushed[0], &payload)
	if payload["imageUrl"] != "http://example.com/image.jpg" {
		t.Fatalf("unexpected payload")
	}
	if db.lastInsert.jobId == "" {
		t.Fatalf("expected db insert")
	}
}

func TestHandleResults(t *testing.T) {
	db := &fakeDB{}
	db.queryRowData = []any{"job-1", "http://x", "result text", "completed", "now"}
	srv := &Server{db: db, rclient: &fakeRedis{}}

	req := httptest.NewRequest("GET", "/results/job-1", nil)
	w := httptest.NewRecorder()
	srv.handleResults(w, req)
	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.StatusCode)
	}
}
