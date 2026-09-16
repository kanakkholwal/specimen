// Package fsstore lays the archive out on disk: one directory per style, grouped by the
// site origin, plus a run-wide archive of raw payloads.
package fsstore

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kanakkholwal/design-supply/internal/normalize"
)

type FS struct{ root string }

func New(root string) (*FS, error) {
	if root == "" {
		root = "data"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	for _, d := range []string{abs, filepath.Join(abs, "sites"), filepath.Join(abs, "raw")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, err
		}
	}
	return &FS{root: abs}, nil
}

func (f *FS) Root() string { return f.root }

// StyleDir is the canonical directory for a style: data/sites/<origin>/<uuid>/.
func (f *FS) StyleDir(origin, id string) string {
	if origin == "" {
		origin = "unknown"
	}
	return filepath.Join(f.root, "sites", normalize.PathSafe(origin), normalize.PathSafe(id))
}

func (f *FS) RawDir() string { return filepath.Join(f.root, "raw") }

// WriteFile writes to a temporary file and renames, so a crash never leaves a half file.
func (f *FS) WriteFile(dir, name string, data []byte, gz bool) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if gz && !strings.HasSuffix(name, ".gz") {
		name += ".gz"
	}
	final := filepath.Join(dir, name)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	write := func() error {
		if gz {
			zw := gzip.NewWriter(tmp)
			if _, err := zw.Write(data); err != nil {
				return err
			}
			return zw.Close()
		}
		_, err := tmp.Write(data)
		return err
	}
	if err := write(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return "", err
	}
	if err := os.Rename(tmpName, final); err != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("fsstore: rename %s: %w", final, err)
	}
	return final, nil
}

// Rel reports a path relative to the archive root, for storing in the database.
func (f *FS) Rel(path string) string {
	rel, err := filepath.Rel(f.root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
