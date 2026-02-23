// Package client provides an HTTP client for communicating with the inventory API.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"inventario/shared/dto"
)

// maxResponseSize limits the response body size to 1MB to prevent OOM.
const maxResponseSize = 1 << 20 // 1 MB

// AuthError is returned when the API responds with 401 or 403.
type AuthError struct {
	StatusCode int
	Message    string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error (status %d): %s", e.StatusCode, e.Message)
}

// Client communicates with the central inventory API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	logger     *slog.Logger
}

// New creates a new API client pointing at the given base URL.
func New(baseURL string, insecureSkipVerify bool, logger *slog.Logger) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify, //nolint:gosec // configurable for development
		},
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		logger: logger,
	}
}

// SetToken sets the bearer token used for authenticated requests.
func (c *Client) SetToken(token string) {
	c.token = token
}

// Enroll registers the device with the API using the enrollment key.
func (c *Client) Enroll(ctx context.Context, enrollmentKey, hostname, serialNumber string) (*dto.EnrollResponse, error) {
	body := dto.EnrollRequest{
		Hostname:     hostname,
		SerialNumber: serialNumber,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal enroll request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/enroll", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Enrollment-Key", enrollmentKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("enroll request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return nil, &AuthError{StatusCode: resp.StatusCode, Message: string(respBody)}
		}
		return nil, fmt.Errorf("enroll failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result dto.EnrollResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode enroll response: %w", err)
	}

	return &result, nil
}

// SubmitInventory sends the full inventory payload to the API.
func (c *Client) SubmitInventory(ctx context.Context, inventory *dto.InventoryRequest) error {
	data, err := json.Marshal(inventory)
	if err != nil {
		return fmt.Errorf("marshal inventory: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/inventory", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("inventory request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return &AuthError{StatusCode: resp.StatusCode, Message: string(respBody)}
		}
		return fmt.Errorf("inventory submit failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// SubmitWithRetry sends the inventory with exponential backoff retry on failure.
func (c *Client) SubmitWithRetry(ctx context.Context, inventory *dto.InventoryRequest, maxRetries int) error {
	for attempt := range maxRetries + 1 {
		err := c.SubmitInventory(ctx, inventory)
		if err == nil {
			return nil
		}

		if attempt == maxRetries {
			return fmt.Errorf("submit failed after %d attempts: %w", maxRetries+1, err)
		}

		backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		jitter := time.Duration(rand.IntN(1000)) * time.Millisecond
		wait := backoff + jitter

		c.logger.Warn("inventory submit failed, retrying",
			"attempt", attempt+1,
			"max_retries", maxRetries,
			"next_retry_in", wait,
			"error", err,
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	return nil
}

// IsAuthError checks if an error indicates an authentication problem (401/403).
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	var authErr *AuthError
	return errors.As(err, &authErr)
}
