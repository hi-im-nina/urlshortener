package main

import (
	"sync"
	"time"
)

// URLEntry holds everything we know about a shortened URL.
type URLEntry struct {
	LongURL   string
	CreatedAt time.Time
	Clicks    int
}

// Store is our in-memory database.
//
// Go maps are NOT thread-safe. If two requests come in at the same time and
// both try to write to the map, the program will crash with a "concurrent map
// writes" panic. We protect the map with a sync.RWMutex:
//
//   - Multiple goroutines can READ at the same time (RLock / RUnlock).
//   - Only one goroutine can WRITE at a time (Lock / Unlock), and no reads
//     are allowed while a write is happening.
type Store struct {
	mu   sync.RWMutex
	urls map[string]*URLEntry
}

// NewStore creates an empty store.
func NewStore() *Store {
	return &Store{
		urls: make(map[string]*URLEntry),
	}
}

// Save stores a new short code → long URL mapping.
func (s *Store) Save(code, longURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urls[code] = &URLEntry{
		LongURL:   longURL,
		CreatedAt: time.Now(),
	}
}

// Get looks up a short code. Returns the entry and whether it was found.
func (s *Store) Get(code string) (*URLEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.urls[code]
	return entry, ok
}

// IncrementClicks records that a short URL was visited.
// We need a write lock here because we're modifying the entry.
func (s *Store) IncrementClicks(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.urls[code]; ok {
		entry.Clicks++
	}
}

// Exists checks whether a code is already taken.
func (s *Store) Exists(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.urls[code]
	return ok
}
