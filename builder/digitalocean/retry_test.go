package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/require"
)

func TestNextBackoff_capsAtMax(t *testing.T) {
	require.Equal(t, 2*time.Second, nextBackoff(time.Second, 15*time.Second))
	require.Equal(t, 4*time.Second, nextBackoff(2*time.Second, 15*time.Second))
	require.Equal(t, 8*time.Second, nextBackoff(4*time.Second, 15*time.Second))
	require.Equal(t, 15*time.Second, nextBackoff(8*time.Second, 15*time.Second))
	require.Equal(t, 15*time.Second, nextBackoff(15*time.Second, 15*time.Second))
}

func TestRetryWithBackoff_succeedsAfterRetries(t *testing.T) {
	var attempts int32
	err := retryWithBackoff(
		context.Background(),
		time.Second,
		time.Millisecond,
		4*time.Millisecond,
		func(error) bool { return true },
		func() error {
			if atomic.AddInt32(&attempts, 1) < 3 {
				return errors.New("not yet")
			}
			return nil
		},
	)
	require.NoError(t, err)
	require.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRetryWithBackoff_returnsNonRetryableError(t *testing.T) {
	permanent := errors.New("permanent")
	err := retryWithBackoff(
		context.Background(),
		time.Second,
		time.Millisecond,
		time.Millisecond,
		func(err error) bool { return false },
		func() error { return permanent },
	)
	require.ErrorIs(t, err, permanent)
}

func TestRetryWithBackoff_timesOut(t *testing.T) {
	err := retryWithBackoff(
		context.Background(),
		20*time.Millisecond,
		time.Millisecond,
		time.Millisecond,
		func(error) bool { return true },
		func() error { return errors.New("still lagging") },
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "timeout")
}

func TestIsInvalidSSHKeyError(t *testing.T) {
	t.Parallel()

	makeAPIErr := func(status int, msg string) error {
		return &godo.ErrorResponse{
			Response: &http.Response{StatusCode: status},
			Message:  msg,
		}
	}

	require.True(t, isInvalidSSHKeyError(makeAPIErr(http.StatusUnprocessableEntity, "59524275 are invalid key identifiers for Droplet creation.")))
	require.False(t, isInvalidSSHKeyError(makeAPIErr(http.StatusUnprocessableEntity, "invalid image")))
	require.False(t, isInvalidSSHKeyError(makeAPIErr(http.StatusNotFound, "invalid key identifiers")))
	require.True(t, isInvalidSSHKeyError(fmt.Errorf("POST https://api.digitalocean.com/v2/droplets: 422 1 are invalid key identifiers for Droplet creation.")))
	require.False(t, isInvalidSSHKeyError(errors.New("boom")))
}

func TestIsNotFoundError(t *testing.T) {
	require.True(t, isNotFoundError(&godo.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusNotFound},
		Message:  "The resource you were accessing could not be found.",
	}))
	require.False(t, isNotFoundError(&godo.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusUnprocessableEntity},
		Message:  "nope",
	}))
	require.False(t, isNotFoundError(errors.New("not found")))
}

func TestDropletIPv4(t *testing.T) {
	t.Parallel()

	_, err := dropletIPv4(nil, false)
	require.ErrorIs(t, err, errIPv4NotFound)

	_, err = dropletIPv4(&godo.Droplet{}, false)
	require.ErrorIs(t, err, errIPv4NotFound)

	droplet := &godo.Droplet{
		Networks: &godo.Networks{
			V4: []godo.NetworkV4{
				{IPAddress: "10.0.0.8", Type: "private"},
				{IPAddress: "203.0.113.10", Type: "public"},
			},
		},
	}

	ip, err := dropletIPv4(droplet, false)
	require.NoError(t, err)
	require.Equal(t, "203.0.113.10", ip)

	ip, err = dropletIPv4(droplet, true)
	require.NoError(t, err)
	require.Equal(t, "10.0.0.8", ip)

	_, err = dropletIPv4(&godo.Droplet{
		Networks: &godo.Networks{V4: []godo.NetworkV4{{IPAddress: "10.0.0.8", Type: "private"}}},
	}, false)
	require.ErrorIs(t, err, errIPv4NotFound)
	require.Contains(t, err.Error(), "public")
}
