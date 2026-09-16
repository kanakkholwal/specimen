// Package engine crawls a site through an adapter, handling scheduling, politeness,
// retries, resumption and storage. It knows nothing about any particular site.
package engine

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/kanak/design-supply/internal/adapter"
)

// Store is the persistence and scheduling surface the engine needs.
type Store interface {
	// Push adds targets to the frontier, ignoring ones already seen. Returns how many
	// were genuinely new.
	Push(ctx context.Context, ts []adapter.Target) (int, error)
	// Lease claims up to n pending targets and marks them in flight.
	Lease(ctx context.Context, n int) ([]adapter.Target, error)
	// Complete records the outcome of a target.
	Complete(ctx context.Context, rawURL string, meta FetchMeta) error
	// Release returns an unfinished target to the pending pool on shutdown.
	Release(ctx context.Context, rawURL string) error
	// Pending reports how many targets remain claimable.
	Pending(ctx context.Context) (int, error)
	// Validators returns cached conditional-GET headers for a URL.
	Validators(ctx context.Context, rawURL string) (etag, lastModified string, ok bool)
	// Write persists one parsed page.
	Write(ctx context.Context, t adapter.Target, r adapter.Result) error
}

type Config struct {
	Workers       int
	RPS           float64
	Burst         int
	Jitter        time.Duration
	Retries       int
	UserAgent     string
	RespectRobots bool
	MaxBodyBytes  int64
	Limit         int
	Refresh       bool
	DryRun        bool
	Log           *slog.Logger
}

func (c *Config) defaults() {
	if c.Workers <= 0 {
		c.Workers = 4
	}
	if c.RPS <= 0 {
		c.RPS = 2
	}
	if c.Burst <= 0 {
		c.Burst = 2
	}
	if c.Retries <= 0 {
		c.Retries = 4
	}
	if c.MaxBodyBytes <= 0 {
		c.MaxBodyBytes = 32 << 20
	}
	if c.UserAgent == "" {
		c.UserAgent = DefaultUserAgent
	}
	if c.Log == nil {
		c.Log = slog.Default()
	}
}

// DefaultUserAgent identifies the crawler and does not impersonate a browser.
const DefaultUserAgent = "design-supply/0.1 (personal design archive; +https://github.com/kanak/design-supply)"

// Stats is the per-run tally reported at the end.
type Stats struct {
	Fetched     atomic.Int64
	NotModified atomic.Int64
	Parsed      atomic.Int64
	Discovered  atomic.Int64
	Failed      atomic.Int64
	Skipped     atomic.Int64
	Bytes       atomic.Int64
	Retries     atomic.Int64
}

type Engine struct {
	cfg      Config
	ad       adapter.Adapter
	store    Store
	fetch    *fetcher
	stats    *Stats
	log      *slog.Logger
	inFlight atomic.Int64
}

func New(ad adapter.Adapter, st Store, cfg Config) *Engine {
	cfg.defaults()
	client := newHTTPClient(cfg.Workers)
	return &Engine{
		cfg:   cfg,
		ad:    ad,
		store: st,
		stats: &Stats{},
		log:   cfg.Log,
		fetch: &fetcher{
			client:  client,
			limiter: newHostLimiter(cfg.RPS, cfg.Burst, cfg.Jitter),
			robots:  newRobotsCache(client, cfg.UserAgent, cfg.RespectRobots),
			agent:   cfg.UserAgent,
			retries: cfg.Retries,
			maxBody: cfg.MaxBodyBytes,
		},
	}
}

func (e *Engine) Stats() *Stats { return e.stats }

// Seed asks the adapter for its starting targets and queues them.
func (e *Engine) Seed(ctx context.Context) (int, error) {
	targets, err := e.ad.Seeds(ctx, e.fetch)
	if err != nil {
		return 0, err
	}
	n, err := e.store.Push(ctx, targets)
	if err != nil {
		return 0, err
	}
	e.log.Info("seeded", "adapter", e.ad.Name(), "targets", len(targets), "new", n)
	return n, nil
}

// Run drains the frontier. It returns when no work remains or the context ends.
func (e *Engine) Run(ctx context.Context) error {
	work := make(chan adapter.Target)
	g, ctx := errgroup.WithContext(ctx)

	var processed atomic.Int64
	var mu sync.Mutex // serialises store writes, which SQLite requires anyway

	for i := 0; i < e.cfg.Workers; i++ {
		g.Go(func() error {
			for t := range work {
				if err := e.process(ctx, t, &mu); err != nil {
					if ctx.Err() != nil {
						// Hand the target back so a resumed run retries it.
						e.store.Release(context.WithoutCancel(ctx), t.URL)
						e.inFlight.Add(-1)
						return err
					}
					e.log.Warn("target failed", "url", t.URL, "err", err)
				}
				e.inFlight.Add(-1)
				processed.Add(1)
			}
			return nil
		})
	}

	g.Go(func() error {
		defer close(work)
		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if e.cfg.Limit > 0 && processed.Load() >= int64(e.cfg.Limit) {
				return nil
			}
			batch, err := e.lease(ctx, processed.Load())
			if err != nil {
				return err
			}
			if len(batch) == 0 {
				if e.inFlight.Load() == 0 {
					return nil
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(100 * time.Millisecond):
				}
				continue
			}
			for _, t := range batch {
				e.inFlight.Add(1)
				select {
				case work <- t:
				case <-ctx.Done():
					e.store.Release(context.WithoutCancel(ctx), t.URL)
					e.inFlight.Add(-1)
					return ctx.Err()
				}
			}
		}
	})

	err := g.Wait()
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (e *Engine) lease(ctx context.Context, processed int64) ([]adapter.Target, error) {
	n := e.cfg.Workers * 2
	if e.cfg.Limit > 0 {
		n = min(n, e.cfg.Limit-int(processed)-int(e.inFlight.Load()))
	}
	if n <= 0 {
		return nil, nil
	}
	return e.store.Lease(ctx, n)
}

func (e *Engine) process(ctx context.Context, t adapter.Target, mu *sync.Mutex) error {
	u, err := url.Parse(t.URL)
	if err != nil {
		return e.complete(ctx, mu, t.URL, FetchMeta{Err: err.Error()})
	}
	if _, ok := e.ad.Classify(u); !ok {
		e.stats.Skipped.Add(1)
		return e.complete(ctx, mu, t.URL, FetchMeta{Err: "not handled by adapter"})
	}

	req, err := e.ad.Request(ctx, t)
	if err != nil {
		return err
	}
	if !e.cfg.Refresh {
		mu.Lock()
		etag, lastMod, ok := e.store.Validators(ctx, t.URL)
		mu.Unlock()
		if ok {
			if etag != "" {
				req.Header.Set("If-None-Match", etag)
			}
			if lastMod != "" {
				req.Header.Set("If-Modified-Since", lastMod)
			}
		}
	}

	if e.cfg.DryRun {
		e.log.Info("dry-run", "url", t.URL, "kind", string(t.Kind))
		return e.complete(ctx, mu, t.URL, FetchMeta{Status: 0, Err: "dry-run"})
	}

	body, resp, meta, err := e.fetch.do(ctx, req)
	if meta.Attempts > 1 {
		e.stats.Retries.Add(int64(meta.Attempts - 1))
	}
	if err != nil {
		e.stats.Failed.Add(1)
		e.complete(ctx, mu, t.URL, meta)
		return err
	}
	if meta.NotModified {
		e.stats.NotModified.Add(1)
		return e.complete(ctx, mu, t.URL, meta)
	}
	e.stats.Fetched.Add(1)
	e.stats.Bytes.Add(meta.Bytes)

	var hdr http.Header
	if resp != nil {
		hdr = resp.Header
	}
	res, err := e.ad.Parse(ctx, t, body, hdr)
	if err != nil {
		e.stats.Failed.Add(1)
		meta.Err = err.Error()
		e.complete(ctx, mu, t.URL, meta)
		return err
	}
	e.stats.Parsed.Add(1)

	mu.Lock()
	defer mu.Unlock()
	if err := e.store.Write(ctx, t, res); err != nil {
		return err
	}
	if len(res.Discover) > 0 {
		n, err := e.store.Push(ctx, res.Discover)
		if err != nil {
			return err
		}
		e.stats.Discovered.Add(int64(n))
	}
	return e.store.Complete(ctx, t.URL, meta)
}

func (e *Engine) complete(ctx context.Context, mu *sync.Mutex, rawURL string, meta FetchMeta) error {
	mu.Lock()
	defer mu.Unlock()
	return e.store.Complete(context.WithoutCancel(ctx), rawURL, meta)
}
