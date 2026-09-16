package engine

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/temoto/robotstxt"
)

// robotsCache fetches and caches one robots.txt per host for the life of a run.
type robotsCache struct {
	mu      sync.Mutex
	hosts   map[string]*robotstxt.RobotsData
	client  *http.Client
	agent   string
	enabled bool
}

func newRobotsCache(client *http.Client, agent string, enabled bool) *robotsCache {
	return &robotsCache{hosts: map[string]*robotstxt.RobotsData{}, client: client, agent: agent, enabled: enabled}
}

// Allowed reports whether the URL may be fetched. A missing or unreadable robots.txt is
// treated as permissive, which is what the standard prescribes.
func (r *robotsCache) Allowed(ctx context.Context, u *url.URL) bool {
	if !r.enabled {
		return true
	}
	data := r.load(ctx, u)
	if data == nil {
		return true
	}
	return data.TestAgent(u.Path, r.agent)
}

func (r *robotsCache) load(ctx context.Context, u *url.URL) *robotstxt.RobotsData {
	key := u.Scheme + "://" + u.Host
	r.mu.Lock()
	data, ok := r.hosts[key]
	r.mu.Unlock()
	if ok {
		return data
	}
	data = r.fetch(ctx, key)
	r.mu.Lock()
	r.hosts[key] = data
	r.mu.Unlock()
	return data
}

func (r *robotsCache) fetch(ctx context.Context, origin string) *robotstxt.RobotsData {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin+"/robots.txt", nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", r.agent)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512<<10))
	if err != nil {
		return nil
	}
	data, err := robotstxt.FromStatusAndBytes(resp.StatusCode, body)
	if err != nil {
		return nil
	}
	return data
}
