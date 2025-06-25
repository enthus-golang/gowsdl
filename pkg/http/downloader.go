// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hooklift/gowsdl/pkg/types"
)

// HTTPConfig interface defines methods required by the downloader
type HTTPConfig interface {
	Build() *http.Client
	GetUserAgent() string
	GetMaxRetries() int
	GetRetryDelay() time.Duration
	GetMaxResponseSize() int64
}

// DownloadFile downloads a file from the given URL using the provided HTTP configuration
func DownloadFile(ctx context.Context, url string, httpConfig HTTPConfig) ([]byte, error) {
	client := httpConfig.Build()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	if userAgent := httpConfig.GetUserAgent(); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	var resp *http.Response
	var lastErr error
	maxRetries := httpConfig.GetMaxRetries()
	retryDelay := httpConfig.GetRetryDelay()

	// Retry logic
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
		}

		resp, err = client.Do(req)
		// Check for context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err == nil && resp.StatusCode < 500 {
			break
		}
		lastErr = err
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed after %d retries: %w", maxRetries+1, lastErr)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, &types.WSDLError{
			Op:   "download",
			Path: url,
			Err:  fmt.Errorf("received HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
	}

	// Limit response size
	limitedReader := io.LimitReader(resp.Body, httpConfig.GetMaxResponseSize())
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}