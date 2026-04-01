// File writing helpers shared across package types.

package espanso

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// writeFile creates dir/filename and writes content from a WriterTo.
func writeFile(dir, filename string, wt io.WriterTo) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	path := filepath.Join(dir, filename)
	f, err := os.Create(path) //nolint:gosec // paths are caller-controlled
	if err != nil {
		return fmt.Errorf("create file %s: %w", path, err)
	}

	writeErr := writeAndSync(f, wt)
	closeErr := f.Close()

	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return fmt.Errorf("close file %s: %w", path, closeErr)
	}
	return nil
}

// encodeYAML marshals v to YAML with 2-space indent and writes to w.
func encodeYAML(w io.Writer, v any) (int64, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return 0, fmt.Errorf("encode yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return 0, fmt.Errorf("close yaml encoder: %w", err)
	}
	n, err := w.Write(buf.Bytes())
	if err != nil {
		return int64(n), fmt.Errorf("write yaml: %w", err)
	}
	return int64(n), nil
}

func writeAndSync(f *os.File, wt io.WriterTo) error {
	if _, err := wt.WriteTo(f); err != nil {
		return fmt.Errorf("write content: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync file: %w", err)
	}
	return nil
}
