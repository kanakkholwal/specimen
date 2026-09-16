package engine

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// FetchMeta is what a completed fetch tells the store.
type FetchMeta struct {
	Status       int
	Bytes        int64
	Duration     time.Duration
	Attempts     int
	ETag         string
	LastModified string
	NotModified  bool
	Err          string
}

type fetcher struct {
	client  *http.Client
	limiter *hostLimiter
	robots  *robotsCache
	agent   string
	retries int
	maxBody int64
}

func newHTTPClient(workers int) *http.Client {
	perHost := max(workers*2, 8)
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          perHost * 2,
		MaxIdleConnsPerHost:   perHost,
		MaxConnsPerHost:       perHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
	}
	return &http.Client{Transport: tr, Timeout: 60 * time.Second}
}

var errRobotsDenied = errors.New("blocked by robots.txt")

// do issues a request with retries, honouring the per-host limiter and robots rules.
// It returns the body, the response, and metadata describing what happened.
func (f *fetcher) do(ctx context.Context, req *http.Request) ([]byte, *http.Response, FetchMeta, error) {
	meta := FetchMeta{}
	if !f.robots.Allowed(ctx, req.URL) {
		meta.Err = errRobotsDenied.Error()
		return nil, nil, meta, errRobotsDenied
	}
	host := req.URL.Host
	start := time.Now()

	var lastErr error
	for attempt := 1; attempt <= f.retries; attempt++ {
		meta.Attempts = attempt
		if err := f.limiter.Wait(ctx, host); err != nil {
			meta.Err = err.Error()
			return nil, nil, meta, err
		}

		r := req.Clone(ctx)
		if r.Header.Get("User-Agent") == "" {
			r.Header.Set("User-Agent", f.agent)
		}
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := f.client.Do(r)
		if err != nil {
			lastErr = err
			f.limiter.Fail(host)
			if !f.sleepBackoff(ctx, attempt, 0) {
				break
			}
			continue
		}

		meta.Status = resp.StatusCode
		meta.ETag = resp.Header.Get("ETag")
		meta.LastModified = resp.Header.Get("Last-Modified")

		switch {
		case resp.StatusCode == http.StatusNotModified:
			resp.Body.Close()
			meta.NotModified = true
			meta.Duration = time.Since(start)
			f.limiter.Succeed(host)
			return nil, resp, meta, nil

		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
			resp.Body.Close()
			f.limiter.Throttle(host, retryAfter)
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			if !f.sleepBackoff(ctx, attempt, retryAfter) {
				break
			}
			continue

		case resp.StatusCode >= 400:
			io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
			resp.Body.Close()
			meta.Duration = time.Since(start)
			meta.Err = fmt.Sprintf("http %d", resp.StatusCode)
			return nil, resp, meta, fmt.Errorf("%s: http %d", req.URL, resp.StatusCode)
		}

		body, err := readBody(resp, f.maxBody)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			f.limiter.Fail(host)
			if !f.sleepBackoff(ctx, attempt, 0) {
				break
			}
			continue
		}
		meta.Bytes = int64(len(body))
		meta.Duration = time.Since(start)
		f.limiter.Succeed(host)
		return body, resp, meta, nil
	}

	meta.Duration = time.Since(start)
	if lastErr != nil {
		meta.Err = lastErr.Error()
	}
	return nil, nil, meta, fmt.Errorf("%s: %w", req.URL, lastErr)
}

// sleepBackoff waits out an exponential, fully jittered delay. It reports false when the
// context ended or the attempt budget is spent.
func (f *fetcher) sleepBackoff(ctx context.Context, attempt int, retryAfter time.Duration) bool {
	if attempt >= f.retries {
		return false
	}
	d := retryAfter
	if d <= 0 {
		base := time.Second << (attempt - 1)
		d = time.Duration(rand.Int64N(int64(base)) + int64(base)/2)
	}
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func readBody(resp *http.Response, limit int64) ([]byte, error) {
	var r io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		zr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer zr.Close()
		r = zr
	}
	if limit > 0 {
		r = io.LimitReader(r, limit)
	}
	return io.ReadAll(r)
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

// Get implements adapter.Fetcher for use during Seeds.
func (f *fetcher) Get(ctx context.Context, rawURL string) ([]byte, *http.Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	body, resp, _, err := f.do(ctx, req)
	return body, resp, err
}
