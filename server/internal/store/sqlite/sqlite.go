// Package sqlite persists the crawl frontier and the normalised design corpus.
package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/kanakkholwal/design-supply/internal/adapter"
	"github.com/kanakkholwal/design-supply/internal/engine"
)

//go:embed schema.sql
var schema string

type DB struct {
	db *sql.DB

	mu     sync.Mutex
	colors map[string]int64
	fonts  map[string]int64
	values map[string]int64 // table|value for shadows, radii, spacings, gradients
	runID  int64
}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // one writer; the engine serialises callers anyway
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &DB{
		db:     db,
		colors: map[string]int64{},
		fonts:  map[string]int64{},
		values: map[string]int64{},
	}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func (d *DB) SQL() *sql.DB { return d.db }

// StartRun opens a run row and returns its id.
func (d *DB) StartRun(ctx context.Context, adapterName string) error {
	res, err := d.db.ExecContext(ctx,
		`INSERT INTO runs(adapter, started_at) VALUES(?, ?)`, adapterName, now())
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	d.mu.Lock()
	d.runID = id
	d.mu.Unlock()
	return nil
}

func (d *DB) FinishRun(ctx context.Context, s *engine.Stats) error {
	d.mu.Lock()
	id := d.runID
	d.mu.Unlock()
	if id == 0 {
		return nil
	}
	_, err := d.db.ExecContext(ctx,
		`UPDATE runs SET finished_at=?, fetched=?, parsed=?, not_modified=?, failed=?, bytes=? WHERE id=?`,
		now(), s.Fetched.Load(), s.Parsed.Load(), s.NotModified.Load(), s.Failed.Load(), s.Bytes.Load(), id)
	return err
}

func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }

// ---- frontier ---------------------------------------------------------------

func (d *DB) Push(ctx context.Context, ts []adapter.Target) (int, error) {
	if len(ts) == 0 {
		return 0, nil
	}
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO frontier(url, kind, depth, meta) VALUES(?,?,?,?) ON CONFLICT(url) DO NOTHING`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	added := 0
	for _, t := range ts {
		meta := ""
		if len(t.Meta) > 0 {
			b, _ := json.Marshal(t.Meta)
			meta = string(b)
		}
		res, err := stmt.ExecContext(ctx, t.URL, string(t.Kind), t.Depth, meta)
		if err != nil {
			return added, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			added++
		}
	}
	return added, tx.Commit()
}

func (d *DB) Lease(ctx context.Context, n int) ([]adapter.Target, error) {
	if n <= 0 {
		n = 1
	}
	rows, err := d.db.QueryContext(ctx, `
		UPDATE frontier SET state='leased', updated_at=datetime('now')
		WHERE url IN (SELECT url FROM frontier WHERE state='pending' ORDER BY depth, rowid LIMIT ?)
		RETURNING url, kind, depth, meta`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []adapter.Target
	for rows.Next() {
		var t adapter.Target
		var kind, meta string
		if err := rows.Scan(&t.URL, &kind, &t.Depth, &meta); err != nil {
			return nil, err
		}
		t.Kind = adapter.Kind(kind)
		if meta != "" {
			json.Unmarshal([]byte(meta), &t.Meta)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (d *DB) Complete(ctx context.Context, rawURL string, meta engine.FetchMeta) error {
	state := "done"
	if meta.Err != "" {
		state = "failed"
	}
	_, err := d.db.ExecContext(ctx, `
		UPDATE frontier SET state=?, attempts=attempts+?, status=?, error=?,
		       etag=COALESCE(NULLIF(?,''), etag),
		       last_modified=COALESCE(NULLIF(?,''), last_modified),
		       fetched_at=?, updated_at=datetime('now')
		WHERE url=?`,
		state, max(meta.Attempts, 1), meta.Status, nullIfEmpty(meta.Err),
		meta.ETag, meta.LastModified, now(), rawURL)
	if err != nil {
		return err
	}
	d.mu.Lock()
	runID := d.runID
	d.mu.Unlock()
	_, err = d.db.ExecContext(ctx,
		`INSERT INTO fetch_log(run_id, url, status, bytes, ms, attempts, error) VALUES(?,?,?,?,?,?,?)`,
		runID, rawURL, meta.Status, meta.Bytes, meta.Duration.Milliseconds(), meta.Attempts,
		nullIfEmpty(meta.Err))
	return err
}

func (d *DB) Release(ctx context.Context, rawURL string) error {
	_, err := d.db.ExecContext(ctx,
		`UPDATE frontier SET state='pending', updated_at=datetime('now') WHERE url=? AND state='leased'`, rawURL)
	return err
}

// ReleaseAllLeased returns stranded leases to the pending pool, for --resume after a crash.
func (d *DB) ReleaseAllLeased(ctx context.Context) (int64, error) {
	res, err := d.db.ExecContext(ctx, `UPDATE frontier SET state='pending' WHERE state='leased'`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *DB) Pending(ctx context.Context) (int, error) {
	var n int
	err := d.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM frontier WHERE state='pending'`).Scan(&n)
	return n, err
}

func (d *DB) Validators(ctx context.Context, rawURL string) (string, string, bool) {
	var etag, lastMod sql.NullString
	var state string
	err := d.db.QueryRowContext(ctx,
		`SELECT COALESCE(etag,''), COALESCE(last_modified,''), state FROM frontier WHERE url=?`, rawURL).
		Scan(&etag, &lastMod, &state)
	if err != nil || (etag.String == "" && lastMod.String == "") {
		return "", "", false
	}
	return etag.String, lastMod.String, true
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// ---- shared value tables -----------------------------------------------------

func (d *DB) colorID(ctx context.Context, tx *sql.Tx, hex string, l, c, h float64) (int64, error) {
	hex = strings.ToLower(strings.TrimSpace(hex))
	if hex == "" {
		return 0, nil
	}
	d.mu.Lock()
	if id, ok := d.colors[hex]; ok {
		d.mu.Unlock()
		return id, nil
	}
	d.mu.Unlock()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO colors(hex, oklch_l, oklch_c, oklch_h) VALUES(?,?,?,?) ON CONFLICT(hex) DO NOTHING`,
		hex, l, c, h); err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM colors WHERE hex=?`, hex).Scan(&id); err != nil {
		return 0, err
	}
	d.mu.Lock()
	d.colors[hex] = id
	d.mu.Unlock()
	return id, nil
}

func (d *DB) fontID(ctx context.Context, tx *sql.Tx, family, source string) (int64, error) {
	family = strings.TrimSpace(family)
	if family == "" {
		return 0, nil
	}
	key := family + "|" + source
	d.mu.Lock()
	if id, ok := d.fonts[key]; ok {
		d.mu.Unlock()
		return id, nil
	}
	d.mu.Unlock()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO fonts(family, source) VALUES(?,?) ON CONFLICT(family, source) DO NOTHING`,
		family, source); err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM fonts WHERE family=? AND source=?`, family, source).Scan(&id); err != nil {
		return 0, err
	}
	d.mu.Lock()
	d.fonts[key] = id
	d.mu.Unlock()
	return id, nil
}

// valueID interns a scalar into one of the single-column value tables.
func (d *DB) valueID(ctx context.Context, tx *sql.Tx, table string, value any, extra ...any) (int64, error) {
	key := fmt.Sprintf("%s|%v", table, value)
	d.mu.Lock()
	if id, ok := d.values[key]; ok {
		d.mu.Unlock()
		return id, nil
	}
	d.mu.Unlock()

	var err error
	if table == "gradients" {
		typ := ""
		if len(extra) > 0 {
			typ, _ = extra[0].(string)
		}
		_, err = tx.ExecContext(ctx,
			`INSERT INTO gradients(value, type) VALUES(?,?) ON CONFLICT(value) DO NOTHING`, value, typ)
	} else {
		_, err = tx.ExecContext(ctx,
			fmt.Sprintf(`INSERT INTO %s(value) VALUES(?) ON CONFLICT(value) DO NOTHING`, table), value)
	}
	if err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT id FROM %s WHERE value=?`, table), value).Scan(&id); err != nil {
		return 0, err
	}
	d.mu.Lock()
	d.values[key] = id
	d.mu.Unlock()
	return id, nil
}

func jsonOrEmpty(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	s := string(b)
	if s == "null" {
		return ""
	}
	return s
}

// dropCaches clears interned ids after a rolled-back transaction.
func (d *DB) dropCaches() {
	d.mu.Lock()
	defer d.mu.Unlock()
	clear(d.colors)
	clear(d.fonts)
	clear(d.values)
}
