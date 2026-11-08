package gocompress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// Gzip compresses a raw byte slice using gzip.
func Gzip(data []byte, level ...int) ([]byte, error) {
	lvl := gzip.DefaultCompression
	if len(level) > 0 {
		lvl = level[0]
	}

	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, lvl)
	if err != nil {
		return nil, fmt.Errorf("compress: new gzip writer: %w", err)
	}

	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		return nil, fmt.Errorf("compress: write failed: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("compress: close failed: %w", err)
	}

	return buf.Bytes(), nil
}

// Gunzip decompresses a gzip-compressed byte slice.
func Gunzip(compressed []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("compress: new gzip reader: %w", err)
	}
	defer zr.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, zr); err != nil {
		return nil, fmt.Errorf("compress: read decompress: %w", err)
	}

	return buf.Bytes(), nil
}

// Config tunes HTTP response compression.
type Config struct {
	MinSize int // Threshold in bytes; responses smaller than this are not compressed (default 1024).
	Level   int // Compression level (-1 to 9).
}

// DefaultConfig provides optimal balance between compression ratio and CPU usage.
var DefaultConfig = Config{
	MinSize: 1024, // 1 KB
	Level:   gzip.DefaultCompression,
}

// Compressor manages recycled gzip writers to prevent garbage collection spikes.
type Compressor struct {
	cfg      Config
	gzipPool sync.Pool
}

// New constructs a Compressor.
func New(cfg ...Config) *Compressor {
	c := DefaultConfig
	if len(cfg) > 0 {
		c = cfg[0]
		if c.MinSize <= 0 {
			c.MinSize = 1024
		}
	}

	comp := &Compressor{cfg: c}
	comp.gzipPool = sync.Pool{
		New: func() any {
			zw, _ := gzip.NewWriterLevel(io.Discard, c.Level)
			return zw
		},
	}
	return comp
}

type gzipResponseWriter struct {
	http.ResponseWriter
	compressor   *Compressor
	gzipWriter   *gzip.Writer
	buffer       *bytes.Buffer
	wroteHeader  bool
	isGzip       bool
	statusCode   int
	exceededSize bool
}

// Handler returns HTTP compression middleware.
func (c *Compressor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only compress if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") ||
			r.Header.Get("Sec-WebSocket-Key") != "" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Vary", "Accept-Encoding")

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			compressor:     c,
			buffer:         &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}

		defer func() {
			if gzw.isGzip && gzw.gzipWriter != nil {
				_ = gzw.gzipWriter.Close()
				c.gzipPool.Put(gzw.gzipWriter)
			} else if !gzw.wroteHeader && gzw.buffer.Len() > 0 {
				// Buffer was below threshold; flush uncompressed
				w.WriteHeader(gzw.statusCode)
				_, _ = w.Write(gzw.buffer.Bytes())
			}
		}()

		next.ServeHTTP(gzw, r)
	})
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	if g.wroteHeader {
		return
	}
	g.statusCode = code
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.isGzip {
		return g.gzipWriter.Write(b)
	}

	// Buffer writes until MinSize threshold
	if !g.exceededSize && g.buffer.Len()+len(b) < g.compressor.cfg.MinSize {
		return g.buffer.Write(b)
	}

	// Exceeded threshold: initialize gzip and flush buffer
	g.exceededSize = true
	g.isGzip = true

	g.Header().Set("Content-Encoding", "gzip")
	g.Header().Del("Content-Length")
	if !g.wroteHeader {
		g.ResponseWriter.WriteHeader(g.statusCode)
		g.wroteHeader = true
	}

	gw := g.compressor.gzipPool.Get().(*gzip.Writer)
	gw.Reset(g.ResponseWriter)
	g.gzipWriter = gw

	// Flush buffered data
	if g.buffer.Len() > 0 {
		if _, err := gw.Write(g.buffer.Bytes()); err != nil {
			return 0, err
		}
		g.buffer.Reset()
	}

	return gw.Write(b)
}

// Middleware creates a handler from config directly.
func Middleware(cfg ...Config) func(http.Handler) http.Handler {
	c := New(cfg...)
	return c.Handler
}
