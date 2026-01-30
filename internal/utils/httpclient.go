package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

// HTTPClient wraps http.Client with additional utilities
type HTTPClient struct {
	client      *http.Client
	maxRetries  int
	userAgent   string
	enableDebug bool
}

// NewHTTPClient creates a new HTTP client with the given timeout
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetries:  3,
		userAgent:   "vibe",
		enableDebug: os.Getenv("VIBE_DEBUG") == "true",
	}
}

// WithMaxRetries sets the maximum number of retries
func (c *HTTPClient) WithMaxRetries(maxRetries int) *HTTPClient {
	c.maxRetries = maxRetries
	return c
}

// WithUserAgent sets the user agent string
func (c *HTTPClient) WithUserAgent(ua string) *HTTPClient {
	c.userAgent = ua
	return c
}

// DoRequest executes an HTTP request with the given parameters
func (c *HTTPClient) DoRequest(
	ctx context.Context,
	method, url string,
	body io.Reader,
	headers map[string]string,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Debug logging
	if c.enableDebug {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s\n", method, url)
		for key, values := range req.Header {
			for _, value := range values {
				// Mask authorization headers
				if key == "Authorization" {
					value = "***REDACTED***"
				}
				fmt.Fprintf(os.Stderr, "[DEBUG] %s: %s\n", key, value)
			}
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Debug response
	if c.enableDebug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Response: %d %s\n", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

// DoJSONRequest executes a JSON request and decodes the response
func (c *HTTPClient) DoJSONRequest(
	ctx context.Context,
	method, url string,
	reqBody interface{},
	respBody interface{},
	headers map[string]string,
) error {
	var bodyReader io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)

		if c.enableDebug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Request body: %s\n", string(jsonData))
		}
	}

	// Set Content-Type for JSON
	if headers == nil {
		headers = make(map[string]string)
	}
	if reqBody != nil {
		headers["Content-Type"] = "application/json"
	}

	resp, err := c.DoRequest(ctx, method, url, bodyReader, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // Body.Close error in defer is acceptable

	// Read response body
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if c.enableDebug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Response body: %s\n", string(respData))
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(respData),
		}
	}

	// Decode response if respBody is provided
	if respBody != nil && len(respData) > 0 {
		if err := json.Unmarshal(respData, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// RetryWithBackoff retries an operation with exponential backoff
func (c *HTTPClient) RetryWithBackoff(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff duration
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}

			if c.enableDebug {
				fmt.Fprintf(os.Stderr, "[DEBUG] Retrying in %v (attempt %d/%d)\n", backoff, attempt, c.maxRetries)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on client errors (4xx), only on 5xx or network errors
		if httpErr, ok := err.(*HTTPError); ok {
			if httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
				return err
			}
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", c.maxRetries, lastErr)
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s - %s", e.StatusCode, e.Status, e.Body)
}

// IsHTTPError checks if an error is an HTTP error
func IsHTTPError(err error) bool {
	_, ok := err.(*HTTPError)
	return ok
}

// GetHTTPError extracts HTTPError from an error
func GetHTTPError(err error) *HTTPError {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr
	}
	return nil
}
