package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Handlers holds shared dependencies for all HTTP handlers.
//
// This pattern (sometimes called a "handler receiver" or "handler struct")
// lets us inject the store without using global variables. Global state makes
// code hard to test — with this pattern, tests can create a fresh store for
// each test case.
type Handlers struct {
	store   URLStore // interface — works with MemoryStore, RedisStore, or anything else
	baseURL string
}

// ShortenRequest is the JSON body we expect for POST /shorten.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse is the JSON body we return after shortening.
type ShortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// StatsResponse is the JSON body we return for GET /stats/{code}.
type StatsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	Clicks      int    `json:"clicks"`
	CreatedAt   string `json:"created_at"`
}

// Shorten handles POST /shorten
// It reads a long URL from the request body and returns a short code.
func (h *Handlers) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate that the URL is actually a URL.
	if err := validateURL(req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code, err := GenerateUniqueCode(h.store)
	if err != nil {
		http.Error(w, "failed to generate code", http.StatusInternalServerError)
		return
	}

	h.store.Save(code, req.URL)

	resp := ShortenResponse{
		ShortURL:    fmt.Sprintf("%s/%s", h.baseURL, code),
		OriginalURL: req.URL,
	}
	writeJSON(w, http.StatusCreated, resp)
}

// Redirect handles GET /{code}
// It looks up the code and sends the user to the original URL.
func (h *Handlers) Redirect(w http.ResponseWriter, r *http.Request) {
	// Strip the leading "/" from the path to get the code.
	// e.g. "/aB3xZ9" → "aB3xZ9"
	code := strings.TrimPrefix(r.URL.Path, "/")

	if code == "" {
		http.Error(w, "missing short code", http.StatusBadRequest)
		return
	}

	entry, ok := h.store.Get(code)
	if !ok {
		http.Error(w, "short URL not found", http.StatusNotFound)
		return
	}

	// Record the click BEFORE redirecting.
	h.store.IncrementClicks(code)

	// 301 = Permanent Redirect (browsers cache it, good for prod)
	// 302 = Temporary Redirect (browsers don't cache, easier to change)
	// We use 302 here so click counts aren't skewed by browser caching.
	http.Redirect(w, r, entry.LongURL, http.StatusFound)
}

// Stats handles GET /stats/{code}
// It returns metadata about a short URL without redirecting.
func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/stats/")

	if code == "" {
		http.Error(w, "missing short code", http.StatusBadRequest)
		return
	}

	entry, ok := h.store.Get(code)
	if !ok {
		http.Error(w, "short URL not found", http.StatusNotFound)
		return
	}

	resp := StatsResponse{
		ShortURL:    fmt.Sprintf("%s/%s", h.baseURL, code),
		OriginalURL: entry.LongURL,
		Clicks:      entry.Clicks,
		CreatedAt:   entry.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	writeJSON(w, http.StatusOK, resp)
}

// validateURL checks that a string is a valid http/https URL.
func validateURL(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %s", raw)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	return nil
}

// writeJSON is a small helper that sets the Content-Type header and encodes
// any Go value as JSON.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
