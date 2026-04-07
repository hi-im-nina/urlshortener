package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestHandlers creates a fresh Handlers instance for each test.
// This ensures tests don't share state with each other — a key principle
// of good testing called "test isolation".
func newTestHandlers() *Handlers {
	return &Handlers{
		store:   NewStore(),
		baseURL: "http://localhost:8080",
	}
}

func TestShorten_Success(t *testing.T) {
	h := newTestHandlers()

	body := `{"url": "https://www.google.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Shorten(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var resp ShortenResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.ShortURL == "" {
		t.Error("expected a short URL in response, got empty string")
	}
	if resp.OriginalURL != "https://www.google.com" {
		t.Errorf("expected original URL to be echoed back, got %q", resp.OriginalURL)
	}
}

func TestShorten_InvalidURL(t *testing.T) {
	h := newTestHandlers()

	body := `{"url": "not-a-url"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Shorten(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid URL, got %d", rec.Code)
	}
}

func TestRedirect_Success(t *testing.T) {
	h := newTestHandlers()
	h.store.Save("abc123", "https://www.google.com")

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()

	h.Redirect(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://www.google.com" {
		t.Errorf("expected redirect to google.com, got %q", loc)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	h := newTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	rec := httptest.NewRecorder()

	h.Redirect(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestStats_ClickCount(t *testing.T) {
	h := newTestHandlers()
	h.store.Save("abc123", "https://www.google.com")

	// Simulate 3 redirects.
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		rec := httptest.NewRecorder()
		h.Redirect(rec, req)
	}

	// Now check stats.
	req := httptest.NewRequest(http.MethodGet, "/stats/abc123", nil)
	rec := httptest.NewRecorder()
	h.Stats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var stats StatsResponse
	json.NewDecoder(rec.Body).Decode(&stats)

	if stats.Clicks != 3 {
		t.Errorf("expected 3 clicks, got %d", stats.Clicks)
	}
}
