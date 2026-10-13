package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/digitalocean/godo"
)

const (
	defaultRetryInitial    = 1 * time.Second
	defaultRetryMaxBackoff = 15 * time.Second
	defaultRetryTimeout    = 1 * time.Minute
)

var errIPv4NotFound = errors.New("IPv4 address not found for droplet")

func nextBackoff(current, max time.Duration) time.Duration {
	next := current * 2
	if next > max || next < current {
		return max
	}
	return next
}

// retryWithBackoff calls fn until it succeeds, ctx is cancelled, or timeout is
// reached. Errors for which shouldRetry returns false are returned immediately.
// Delay doubles after each retry and is capped at maxBackoff.
func retryWithBackoff(
	ctx context.Context,
	timeout, initial, maxBackoff time.Duration,
	shouldRetry func(error) bool,
	fn func() error,
) error {
	if timeout <= 0 {
		timeout = defaultRetryTimeout
	}
	if initial <= 0 {
		initial = defaultRetryInitial
	}
	if maxBackoff < initial {
		maxBackoff = initial
	}

	deadline := time.Now().Add(timeout)
	delay := initial
	attempt := 0
	var lastErr error

	for {
		attempt++
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return fmt.Errorf("%w: %v", err, lastErr)
			}
			return err
		}

		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if shouldRetry != nil && !shouldRetry(err) {
			return err
		}

		remaining := time.Until(deadline)
		if remaining <= 0 {
			return fmt.Errorf("timeout after %d attempts: %w", attempt, lastErr)
		}

		wait := delay
		if wait > remaining {
			wait = remaining
		}

		log.Printf("[DEBUG] retryable error (attempt %d), waiting %s: %s", attempt, wait, err)

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			if lastErr != nil {
				return fmt.Errorf("%w: %v", ctx.Err(), lastErr)
			}
			return ctx.Err()
		case <-timer.C:
		}

		delay = nextBackoff(delay, maxBackoff)
	}
}

func isInvalidSSHKeyError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *godo.ErrorResponse
	if errors.As(err, &apiErr) {
		if apiErr.Response != nil && apiErr.Response.StatusCode != http.StatusUnprocessableEntity {
			return false
		}
		return strings.Contains(apiErr.Message, "invalid key identifiers")
	}

	return strings.Contains(err.Error(), "invalid key identifiers")
}

func isNotFoundError(err error) bool {
	var apiErr *godo.ErrorResponse
	if errors.As(err, &apiErr) && apiErr.Response != nil {
		return apiErr.Response.StatusCode == http.StatusNotFound
	}
	return false
}

func dropletIPv4(droplet *godo.Droplet, private bool) (string, error) {
	if droplet == nil || droplet.Networks == nil || len(droplet.Networks.V4) == 0 {
		return "", errIPv4NotFound
	}

	want := "public"
	if private {
		want = "private"
	}

	for _, network := range droplet.Networks.V4 {
		if network.Type == want {
			return network.IPAddress, nil
		}
	}

	if private {
		return "", fmt.Errorf("%w: could not find a private IPv4 address for this droplet", errIPv4NotFound)
	}
	return "", fmt.Errorf("%w: could not find a public IPv4 address for this droplet", errIPv4NotFound)
}
