package scraper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	// MaxFileSize defines the maximum permitted content size (500KB).
	MaxFileSize int64 = 500 * 1024
	// DefaultFetchTimeout is the maximum duration for a single fetch request.
	DefaultFetchTimeout = 30 * time.Second
	// MaxFetchRetries is the maximum number of retry attempts for transient errors.
	MaxFetchRetries = 3
)

var (
	// ErrFileTooLarge is returned when content exceeds MaxFileSize.
	ErrFileTooLarge = errors.New("file size exceeds maximum allowed limit (500KB)")
)

// CacheEntry stores cached response data for conditional requests.
type CacheEntry struct {
	ETag         string
	LastModified string
	Body         []byte
	FetchedAt    time.Time
}

// Fetcher handles HTTP retrieval with in-memory caching, retries, and size guards.
type Fetcher struct {
	client  *http.Client
	cache   map[string]CacheEntry
	cacheMu sync.RWMutex
}

// NewFetcher creates a new Fetcher instance. If client is nil, a default client is configured.
func NewFetcher(client *http.Client) *Fetcher {
	if client == nil {
		client = &http.Client{
			Timeout: DefaultFetchTimeout,
		}
	}
	return &Fetcher{
		client: client,
		cache:  make(map[string]CacheEntry),
	}
}

// Fetch retrieves the body of the specified URL, honoring ETags and size limits.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	backoff := 100 * time.Millisecond

	for attempt := 1; attempt <= MaxFetchRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create fetch request: %w", err)
		}

		req.Header.Set("User-Agent", "AgentSight-Scraper/1.0 (+https://github.com/agentsight)")

		// Check cache for conditional headers
		f.cacheMu.RLock()
		cached, exists := f.cache[rawURL]
		f.cacheMu.RUnlock()

		if exists {
			if cached.ETag != "" {
				req.Header.Set("If-None-Match", cached.ETag)
			}
			if cached.LastModified != "" {
				req.Header.Set("If-Modified-Since", cached.LastModified)
			}
		}

		resp, err := f.client.Do(req)
		if err != nil {
			lastErr = err
			slog.Warn("fetch attempt failed", "url", rawURL, "attempt", attempt, "err", err)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
				continue
			}
		}

		// Handle 304 Not Modified
		if resp.StatusCode == http.StatusNotModified && exists {
			resp.Body.Close()
			return cached.Body, nil
		}

		// Check rate limiting or server errors
		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			lastErr = fmt.Errorf("server returned status code: %d", resp.StatusCode)
			slog.Warn("fetch transient status code", "url", rawURL, "status", resp.StatusCode, "attempt", attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
				continue
			}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			return nil, fmt.Errorf("fetch failed with status %d for %s", resp.StatusCode, rawURL)
		}

		// Check Content-Length header if present
		if clHeader := resp.Header.Get("Content-Length"); clHeader != "" {
			if cl, err := strconv.ParseInt(clHeader, 10, 64); err == nil && cl > MaxFileSize {
				resp.Body.Close()
				return nil, ErrFileTooLarge
			}
		}

		// Read up to MaxFileSize + 1 to detect overflows
		limitedReader := io.LimitReader(resp.Body, MaxFileSize+1)
		bodyBytes, err := io.ReadAll(limitedReader)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
				continue
			}
		}

		if int64(len(bodyBytes)) > MaxFileSize {
			return nil, ErrFileTooLarge
		}

		// Save into cache
		etag := resp.Header.Get("ETag")
		lastModified := resp.Header.Get("Last-Modified")
		f.cacheMu.Lock()
		f.cache[rawURL] = CacheEntry{
			ETag:         etag,
			LastModified: lastModified,
			Body:         bodyBytes,
			FetchedAt:    time.Now(),
		}
		f.cacheMu.Unlock()

		return bodyBytes, nil
	}

	// If all attempts failed, fallback to cache if available
	f.cacheMu.RLock()
	cached, exists := f.cache[rawURL]
	f.cacheMu.RUnlock()
	if exists {
		slog.Warn("falling back to cached data after failed retries", "url", rawURL)
		return cached.Body, nil
	}

	return nil, fmt.Errorf("all %d fetch attempts failed for %s: %w", MaxFetchRetries, rawURL, lastErr)
}

// FetchString retrieves content and converts it to a string.
func (f *Fetcher) FetchString(ctx context.Context, rawURL string) (string, error) {
	bytes, err := f.Fetch(ctx, rawURL)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
