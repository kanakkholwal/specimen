package engine

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kanak/design-supply/internal/adapter"
)

// memStore is an in-memory engine.Store for exercising the pipeline without SQLite.
type memStore struct {
	mu        sync.Mutex
	pending   []adapter.Target
	seen      map[string]bool
	completed map[string]FetchMeta
	written   []adapter.Result
	etags     map[string]string
}

func newMemStore() *memStore {
	return &memStore{seen: map[string]bool{}, completed: map[string]FetchMeta{}, etags: map[string]string{}}
}

func (m *memStore) Push(_ context.Context, ts []adapter.Target) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	added := 0
	for _, t := range ts {
		if m.seen[t.URL] {
			continue
		}
		m.seen[t.URL] = true
		m.pending = append(m.pending, t)
		added++
	}
	return added, nil
}

func (m *memStore) Lease(_ context.Context, n int) ([]adapter.Target, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if n > len(m.pending) {
		n = len(m.pending)
	}
	out := m.pending[:n]
	m.pending = m.pending[n:]
	return out, nil
}

func (m *memStore) Complete(_ context.Context, u string, meta FetchMeta) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.completed[u] = meta
	if meta.ETag != "" {
		m.etags[u] = meta.ETag
	}
	return nil
}

func (m *memStore) Release(_ context.Context, u string) error { return nil }

func (m *memStore) Pending(_ context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.pending), nil
}

func (m *memStore) Validators(_ context.Context, u string) (string, string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.etags[u]
	return e, "", ok
}

func (m *memStore) Write(_ context.Context, _ adapter.Target, r adapter.Result) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.written = append(m.written, r)
	return nil
}

// testAdapter crawls whatever URLs it is seeded with and records nothing but the body.
type testAdapter struct {
	seeds  []string
	parsed atomic.Int64
}

func (a *testAdapter) Name() string { return "test" }

func (a *testAdapter) Seeds(context.Context, adapter.Fetcher) ([]adapter.Target, error) {
	out := make([]adapter.Target, 0, len(a.seeds))
	for _, s := range a.seeds {
		out = append(out, adapter.Target{URL: s, Kind: "page"})
	}
	return out, nil
}

func (a *testAdapter) Classify(*url.URL) (adapter.Kind, bool) { return "page", true }

func (a *testAdapter) Request(ctx context.Context, t adapter.Target) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
}

func (a *testAdapter) Parse(_ context.Context, t adapter.Target, body []byte, _ http.Header) (adapter.Result, error) {
	a.parsed.Add(1)
	return adapter.Result{Records: []adapter.Record{{Kind: "page", ID: t.URL, Data: string(body)}}}, nil
}

func quietConfig(c Config) Config {
	c.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	c.RPS = 1000
	c.Jitter = 0
	return c
}

func runEngine(t *testing.T, ad adapter.Adapter, st Store, cfg Config) *Engine {
	t.Helper()
	e := New(ad, st, quietConfig(cfg))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := e.Seed(ctx); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := e.Run(ctx); err != nil {
		t.Fatalf("run: %v", err)
	}
	return e
}

func TestRetriesTransientFailures(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if hits.Add(1) <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.WriteString(w, "ok")
	}))
	defer srv.Close()

	st := newMemStore()
	ad := &testAdapter{seeds: []string{srv.URL + "/a"}}
	e := runEngine(t, ad, st, Config{Workers: 1, Retries: 4})

	if got := e.Stats().Fetched.Load(); got != 1 {
		t.Fatalf("fetched = %d, want 1", got)
	}
	if got := e.Stats().Retries.Load(); got != 2 {
		t.Errorf("retries = %d, want 2", got)
	}
	if got := hits.Load(); got != 3 {
		t.Errorf("server saw %d requests, want 3", got)
	}
}

func TestConditionalGetSkipsUnchanged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		if r.Header.Get("If-None-Match") == `"v1"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		io.WriteString(w, "body")
	}))
	defer srv.Close()

	st := newMemStore()
	first := runEngine(t, &testAdapter{seeds: []string{srv.URL + "/a"}}, st, Config{Workers: 1})
	if got := first.Stats().Parsed.Load(); got != 1 {
		t.Fatalf("first run parsed = %d, want 1", got)
	}

	st.pending = nil
	st.seen = map[string]bool{}
	second := runEngine(t, &testAdapter{seeds: []string{srv.URL + "/a"}}, st, Config{Workers: 1})
	if got := second.Stats().NotModified.Load(); got != 1 {
		t.Fatalf("second run 304s = %d, want 1", got)
	}
	if got := second.Stats().Parsed.Load(); got != 0 {
		t.Errorf("second run parsed = %d, want 0", got)
	}
}

func TestRobotsBlocksDisallowedPath(t *testing.T) {
	var fetched atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			io.WriteString(w, "User-agent: *\nDisallow: /api/\n")
			return
		}
		fetched.Add(1)
		io.WriteString(w, "ok")
	}))
	defer srv.Close()

	st := newMemStore()
	ad := &testAdapter{seeds: []string{srv.URL + "/api/secret", srv.URL + "/allowed"}}
	e := runEngine(t, ad, st, Config{Workers: 1, RespectRobots: true})

	if got := fetched.Load(); got != 1 {
		t.Fatalf("server served %d pages, want 1", got)
	}
	if got := e.Stats().Failed.Load(); got != 1 {
		t.Errorf("failed = %d, want 1 (the disallowed target)", got)
	}
}

func TestLimitStopsExactly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		io.WriteString(w, "ok")
	}))
	defer srv.Close()

	var seeds []string
	for i := 0; i < 20; i++ {
		seeds = append(seeds, srv.URL+"/p"+string(rune('a'+i)))
	}
	st := newMemStore()
	e := runEngine(t, &testAdapter{seeds: seeds}, st, Config{Workers: 4, Limit: 6})

	if got := e.Stats().Fetched.Load(); got != 6 {
		t.Fatalf("fetched = %d, want exactly 6", got)
	}
}

func TestThrottleHalvesRateOnPushBack(t *testing.T) {
	l := newHostLimiter(8, 1, 0)
	if got := l.Rate("h"); got != 8 {
		t.Fatalf("initial rate = %v, want 8", got)
	}
	l.Throttle("h", 0)
	if got := l.Rate("h"); got != 4 {
		t.Fatalf("throttled rate = %v, want 4", got)
	}
	for i := 0; i < recoverAfterSuccesses; i++ {
		l.Succeed("h")
	}
	if got := l.Rate("h"); got <= 4 {
		t.Fatalf("recovered rate = %v, want above 4", got)
	}
}

func TestBreakerTripsAfterConsecutiveFailures(t *testing.T) {
	l := newHostLimiter(8, 1, 0)
	for i := 0; i < breakerFailures-1; i++ {
		if l.Fail("h") {
			t.Fatalf("breaker tripped early at failure %d", i+1)
		}
	}
	if !l.Fail("h") {
		t.Fatal("breaker did not trip at the threshold")
	}
}
