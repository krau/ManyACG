package telegoapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/mymmrac/telego/telegoapi"
)

var _ telegoapi.Caller = (*RetryRateLimitCaller)(nil)

// RetryRateLimitCaller decorator over [Caller] that provides retries with exponential backoff on rate limit errors
// Depending on [RetryRateLimit] will wait for rate limit timeout to reset or abort, defaults to do nothing
// Delay = min((ExponentBase ^ AttemptNumber) * StartDelay, MaxDelay)
type RetryRateLimitCaller struct {
	// Underling caller
	Caller telegoapi.Caller
	// Max number of attempts to make a call
	MaxAttempts int
	// Exponent base for delay
	ExponentBase float64
	// Starting delay duration
	StartDelay time.Duration
	// Maximum delay duration
	MaxDelay time.Duration
	// Rate limit behavior
	RateLimit telegoapi.RetryRateLimit
	// Buffer request data, if set to true requests that usually stream body using io.Reader will be buffered
	// to support retrying such requests
	//
	// Warning: Enabling this may lead to excessive memory consumption and OOMKill
	BufferRequestData bool
}

// ErrMaxRetryAttempts returned when max retry attempts reached
var ErrMaxRetryAttempts = errors.New("max retry attempts reached")

// Call makes calls using provided caller with retries
func (r *RetryRateLimitCaller) Call(ctx context.Context, url string, data *telegoapi.RequestData) (response *telegoapi.Response, err error) {
	if data.BodyStream != nil && r.BufferRequestData {
		data.BodyRaw, err = io.ReadAll(data.BodyStream)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		data.BodyStream = nil
	}

	for i := 0; i < r.MaxAttempts; i++ {
		response, err = r.Caller.Call(ctx, url, data)
		if err == nil && (response.Error == nil || response.ErrorCode == 0) {
			return response, nil
		}
		if err == nil {
			err = response.Error
		}

		if i == r.MaxAttempts-1 {
			break
		}

		delay, ok := r.handleError(err)
		if !ok {
			return nil, err
		}
		if delay == 0 {
			delay = min(time.Duration(math.Pow(r.ExponentBase, float64(i)))*r.StartDelay, r.MaxDelay)
		}

		select {
		case <-ctx.Done():
			return nil, errors.Join(err, ctx.Err())
		case <-time.After(delay):
			// Continue
		}
	}

	return nil, errors.Join(err, ErrMaxRetryAttempts)
}

func (r *RetryRateLimitCaller) handleError(err error) (time.Duration, bool) {
	var apiErr *telegoapi.Error
	if errors.As(err, &apiErr) && apiErr.ErrorCode == 429 && apiErr.Parameters != nil { // Rate limit
		switch r.RateLimit {
		case telegoapi.RetryRateLimitSkip:
			return 0, true
		case telegoapi.RetryRateLimitAbort:
			return 0, false
		case telegoapi.RetryRateLimitWait:
			return time.Duration(apiErr.Parameters.RetryAfter) * time.Second, true
		case telegoapi.RetryRateLimitWaitOrAbort:
			delay := time.Duration(apiErr.Parameters.RetryAfter) * time.Second
			if delay > r.MaxDelay {
				return 0, false
			}
			return delay, true
		default:
			// Skip unknown rate limit behavior
		}
	}
	// only retry on rate limit errors
	return 0, false
}
