package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient is a shared HTTP helper for all providers.
type APIClient struct {
	BaseURL      string
	APIKey       string
	KeyHeader    string            // e.g. "X-API-Key", "Authorization"
	KeyPrefix    string            // e.g. "Bearer ", "" (Anthropic uses raw key)
	ExtraHeaders map[string]string // e.g. {"anthropic-version": "2023-06-01"}
	MaxRetries   int               // default 3
	ProviderName string            // for error messages / retry logs
	Verbose      bool
}

// retryableStatus returns true if the HTTP status code should trigger a retry.
func retryableStatus(code int) bool {
	switch code {
	case 429, 500, 502, 503, 504:
		return true
	}
	return false
}

// Post marshals payload as JSON, POSTs to BaseURL+path, and decodes the response into result.
// Handles auth header, retry with exponential backoff on 429/5xx.
func (c *APIClient) Post(ctx context.Context, path string, payload interface{}, result interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	maxRetries := c.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	providerName := c.ProviderName
	if providerName == "" {
		providerName = "api"
	}

	url := c.BaseURL + path
	var lastErr error

	for attempt := 1; attempt <= maxRetries+1; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		if c.APIKey != "" {
			req.Header.Set(c.KeyHeader, c.KeyPrefix+c.APIKey)
		}
		for k, v := range c.ExtraHeaders {
			req.Header.Set(k, v)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			// Network errors are retryable
			if attempt <= maxRetries {
				backoff := time.Duration(1<<uint(attempt-1)) * time.Second
				if c.Verbose {
					fmt.Printf("[retry] %s: network error, retrying in %s (attempt %d/%d)\n",
						providerName, backoff, attempt, maxRetries)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
				}
				continue
			}
			return lastErr
		}

		if resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return fmt.Errorf("failed to decode %s response: %w", providerName, err)
			}
			return nil
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		statusText := resp.Status

		// Non-retryable client errors
		if !retryableStatus(resp.StatusCode) {
			switch resp.StatusCode {
			case 401, 403:
				return fmt.Errorf("%s API rejected the request (%s). Check your API key is valid.\nDetails: %s",
					providerName, statusText, string(body))
			default:
				return fmt.Errorf("%s API error (%s): %s", providerName, statusText, string(body))
			}
		}

		// Retryable — check if we have retries left
		if attempt <= maxRetries {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if c.Verbose {
				if resp.StatusCode == 429 {
					fmt.Printf("[retry] %s: rate limited, retrying in %s (attempt %d/%d)\n",
						providerName, backoff, attempt, maxRetries)
				} else {
					fmt.Printf("[retry] %s: server error %d, retrying in %s (attempt %d/%d)\n",
						providerName, resp.StatusCode, backoff, attempt, maxRetries)
				}
			}
			lastErr = fmt.Errorf("%s API error (%s): %s", providerName, statusText, string(body))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			continue
		}

		return fmt.Errorf("%s API error (%s): %s", providerName, statusText, string(body))
	}

	return lastErr
}
