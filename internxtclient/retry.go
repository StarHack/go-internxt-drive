package internxtclient

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxRequestRetries = 4

type retryableErrorBody struct {
	Retryable  bool `json:"retryable"`
	RetryAfter int  `json:"retry_after"`
}

func isRetryableHTTPStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return status >= 500 && status <= 599
	}
}

func isRetryableNetworkError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	message := strings.ToLower(err.Error())
	for _, fragment := range []string{
		"no route to host",
		"connection reset",
		"connection refused",
		"broken pipe",
		"i/o timeout",
		"tls handshake timeout",
		"unexpected eof",
		"temporary failure",
		"network is unreachable",
	} {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}

func shouldRetryResponse(status int, body []byte) bool {
	if isRetryableHTTPStatus(status) {
		return true
	}
	if status == http.StatusNotFound && isParentFolderMissingBody(body) {
		return true
	}
	retryable, _ := parseRetryableBody(body)
	return retryable
}

func isParentFolderMissingBody(body []byte) bool {
	return strings.Contains(strings.ToLower(string(body)), "parent folder does not exist")
}

func parseRetryableBody(body []byte) (retryable bool, retryAfter time.Duration) {
	var payload retryableErrorBody
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, 0
	}
	if !payload.Retryable {
		return false, 0
	}
	if payload.RetryAfter > 0 {
		return true, time.Duration(payload.RetryAfter) * time.Second
	}
	return true, 0
}

func retryWait(attempt int, body []byte, headers http.Header) time.Duration {
	if attempt <= 0 {
		return 0
	}
	if retryable, retryAfter := parseRetryableBody(body); retryable && retryAfter > 0 {
		return capRetryWait(retryAfter)
	}
	if headers != nil {
		if value := headers.Get("Retry-After"); value != "" {
			if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
				return capRetryWait(time.Duration(seconds) * time.Second)
			}
		}
	}
	delay := time.Second << (attempt - 1)
	return capRetryWait(delay)
}

func capRetryWait(delay time.Duration) time.Duration {
	const maxDelay = 30 * time.Second
	if delay > maxDelay {
		return maxDelay
	}
	if delay < time.Second {
		return time.Second
	}
	return delay
}

func shouldRetryAttempt(response *Response, err error, attempt int) bool {
	if attempt >= maxRequestRetries {
		return false
	}
	if isRetryableNetworkError(err) {
		return true
	}
	if response == nil {
		return false
	}
	return shouldRetryResponse(response.StatusCode, response.Body)
}
