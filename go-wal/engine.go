package wal

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	headerSize = 24 // 4 (crc) + 8 (lsn) + 8 (timestamp) + 4 (payload length)
	walExt     = ".wal"
)

// Record represents a single logged entry in the write-ahead log.
type Record struct {
	LSN       uint64
	Timestamp time.Time
	Data      []byte
}

// WAL provides a segmented crash-resilient write-ahead log.
type WAL struct {
	cfg         Config
	mu          sync.Mutex
	currentFile *os.File
	segmentIdx  int
	currentSize int64
	lastLSN     uint64
	closed      bool
	closeChan   chan struct{}
}

// Open initializes or recovers a WAL directory.
func Open(opts ...Option) (*WAL, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("wal: create directory failed: %w", err)
	}

	w := &WAL{
		cfg:       cfg,
		closeChan: make(chan struct{}),
	}

	// Discover existing segments and determine last LSN
	segments, err := w.listSegments()
	if err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		// First segment
		if err := w.rollSegment(1); err != nil {
			return nil, err
		}
	} else {
		latest := segments[len(segments)-1]
		w.segmentIdx = latest.index

		// Replay to find last LSN and verify file state
		lastLSN, err := w.scanSegmentMaxLSN(latest.path)
		if err != nil {
			return nil, err
		}
		w.lastLSN = lastLSN

		// Open latest segment for appending
		f, err := os.OpenFile(latest.path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		fi, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		w.currentFile = f
		w.currentSize = fi.Size()
	}

	if cfg.SyncMode == SyncInterval {
		go w.periodicSyncLoop()
	}

	return w, nil
}

// Append writes data to the log, returning a monotonic Log Sequence Number.
func (w *WAL) Append(data []byte) (uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, ErrWALClosed
	}

	recordLen := int64(headerSize + len(data))
	if w.currentSize+recordLen > w.cfg.MaxSegmentSize {
		if err := w.rollSegment(w.segmentIdx + 1); err != nil {
			return 0, err
		}
	}

	lsn := w.lastLSN + 1
	now := time.Now()
	nowNano := uint64(now.UnixNano())

	var header [headerSize]byte
	binary.BigEndian.PutUint64(header[4:12], lsn)
	binary.BigEndian.PutUint64(header[12:20], nowNano)
	binary.BigEndian.PutUint32(header[20:24], uint32(len(data)))

	// Compute CRC32
	crc := crc32.NewIEEE()
	crc.Write(header[4:24])
	crc.Write(data)
	binary.BigEndian.PutUint32(header[0:4], crc.Sum32())

	// Write header + payload
	if _, err := w.currentFile.Write(header[:]); err != nil {
		return 0, err
	}
	if len(data) > 0 {
		if _, err := w.currentFile.Write(data); err != nil {
			return 0, err
		}
	}

	w.currentSize += recordLen
	w.lastLSN = lsn

	if w.cfg.SyncMode == SyncAlways {
		if err := w.currentFile.Sync(); err != nil {
			return 0, err
		}
	}

	return lsn, nil
}

// Replay reads records starting from fromLSN and invokes fn for each record.
func (w *WAL) Replay(fromLSN uint64, fn func(Record) error) error {
	w.mu.Lock()
	// Flush any pending buffer
	if w.currentFile != nil {
		_ = w.currentFile.Sync()
	}
	w.mu.Unlock()

	segments, err := w.listSegments()
	if err != nil {
		return err
	}

	for _, seg := range segments {
		f, err := os.Open(seg.path)
		if err != nil {
			return err
		}

		err = func() error {
			defer f.Close()
			var header [headerSize]byte

			for {
				_, err := io.ReadFull(f, header[:])
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					return nil // Segment ended cleanly
				}
				if err != nil {
					return err
				}

				expectedCRC := binary.BigEndian.Uint32(header[0:4])
				lsn := binary.BigEndian.Uint64(header[4:12])
				tsNano := binary.BigEndian.Uint64(header[12:20])
				payloadLen := binary.BigEndian.Uint32(header[20:24])

				payload := make([]byte, payloadLen)
				if payloadLen > 0 {
					if _, err := io.ReadFull(f, payload); err != nil {
						return ErrCorruptRecord
					}
				}

				// Validate CRC
				crc := crc32.NewIEEE()
				crc.Write(header[4:24])
				crc.Write(payload)
				if crc.Sum32() != expectedCRC {
					return ErrCorruptRecord
				}

				if lsn >= fromLSN {
					rec := Record{
						LSN:       lsn,
						Timestamp: time.Unix(0, int64(tsNano)),
						Data:      payload,
					}
					if err := fn(rec); err != nil {
						return err
					}
				}
			}
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

// LastLSN returns the latest assigned sequence number.
func (w *WAL) LastLSN() uint64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastLSN
}

// TruncateBefore removes older segment files where all records have LSN < minLSN.
func (w *WAL) TruncateBefore(minLSN uint64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	segments, err := w.listSegments()
	if err != nil {
		return err
	}

	// Do not delete the active latest segment
	for i := 0; i < len(segments)-1; i++ {
		seg := segments[i]
		maxLSN, err := w.scanSegmentMaxLSN(seg.path)
		if err != nil {
			continue
		}
		if maxLSN < minLSN {
			_ = os.Remove(seg.path)
		}
	}
	return nil
}

// Close safely flushes and closes the WAL.
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true
	close(w.closeChan)

	if w.currentFile != nil {
		_ = w.currentFile.Sync()
		return w.currentFile.Close()
	}
	return nil
}

// -------------------------------------------------------------
// Internal Segment Helpers
// -------------------------------------------------------------

type segmentInfo struct {
	index int
	path  string
}

func (w *WAL) listSegments() ([]segmentInfo, error) {
	entries, err := os.ReadDir(w.cfg.Dir)
	if err != nil {
		return nil, err
	}

	var segs []segmentInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), walExt) {
			continue
		}
		base := strings.TrimSuffix(entry.Name(), walExt)
		idx, err := strconv.Atoi(base)
		if err != nil {
			continue
		}
		segs = append(segs, segmentInfo{
			index: idx,
			path:  filepath.Join(w.cfg.Dir, entry.Name()),
		})
	}

	sort.Slice(segs, func(i, j int) bool {
		return segs[i].index < segs[j].index
	})
	return segs, nil
}

func (w *WAL) rollSegment(newIdx int) error {
	if w.currentFile != nil {
		_ = w.currentFile.Sync()
		_ = w.currentFile.Close()
	}

	fileName := fmt.Sprintf("%08d%s", newIdx, walExt)
	filePath := filepath.Join(w.cfg.Dir, fileName)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	w.currentFile = f
	w.segmentIdx = newIdx
	w.currentSize = 0
	return nil
}

func (w *WAL) scanSegmentMaxLSN(path string) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var header [headerSize]byte
	var maxLSN uint64

	for {
		_, err := io.ReadFull(f, header[:])
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return maxLSN, err
		}

		lsn := binary.BigEndian.Uint64(header[4:12])
		payloadLen := binary.BigEndian.Uint32(header[20:24])

		if lsn > maxLSN {
			maxLSN = lsn
		}

		// Skip payload
		if payloadLen > 0 {
			if _, err := f.Seek(int64(payloadLen), io.SeekCurrent); err != nil {
				break
			}
		}
	}
	return maxLSN, nil
}

func (w *WAL) periodicSyncLoop() {
	ticker := time.NewTicker(w.cfg.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.closeChan:
			return
		case <-ticker.C:
			w.mu.Lock()
			if w.currentFile != nil && !w.closed {
				_ = w.currentFile.Sync()
			}
			w.mu.Unlock()
		}
	}
}
