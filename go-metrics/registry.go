package gometrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type metricEntry struct {
	name       string
	help       string
	metricType MetricType
	labels     Labels
	counter    *Counter
	gauge      *Gauge
	histogram  *Histogram
}

// Registry stores and formats registered metrics for Prometheus scrapers.
type Registry struct {
	mu      sync.RWMutex
	metrics map[string]*metricEntry
}

// NewRegistry initializes an empty metrics Registry.
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]*metricEntry),
	}
}

// DefaultRegistry is the standard global registry.
var DefaultRegistry = NewRegistry()

func (r *Registry) makeKey(name string, labels Labels) string {
	return name + labels.Format()
}

// Counter registers or retrieves an existing Counter with matching name and labels.
func (r *Registry) Counter(name, help string, labels Labels) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(name, labels)
	if entry, ok := r.metrics[key]; ok && entry.counter != nil {
		return entry.counter
	}

	c := NewCounter()
	r.metrics[key] = &metricEntry{
		name:       name,
		help:       help,
		metricType: TypeCounter,
		labels:     labels,
		counter:    c,
	}
	return c
}

// Gauge registers or retrieves an existing Gauge with matching name and labels.
func (r *Registry) Gauge(name, help string, labels Labels) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(name, labels)
	if entry, ok := r.metrics[key]; ok && entry.gauge != nil {
		return entry.gauge
	}

	g := NewGauge()
	r.metrics[key] = &metricEntry{
		name:       name,
		help:       help,
		metricType: TypeGauge,
		labels:     labels,
		gauge:      g,
	}
	return g
}

// Histogram registers or retrieves an existing Histogram with matching name and labels.
func (r *Registry) Histogram(name, help string, buckets []float64, labels Labels) *Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(name, labels)
	if entry, ok := r.metrics[key]; ok && entry.histogram != nil {
		return entry.histogram
	}

	h := NewHistogram(buckets)
	r.metrics[key] = &metricEntry{
		name:       name,
		help:       help,
		metricType: TypeHistogram,
		labels:     labels,
		histogram:  h,
	}
	return h
}

// Gather serializes all metrics into official Prometheus text exposition format (version 0.0.4).
func (r *Registry) Gather() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Group by name for HELP and TYPE headers
	grouped := make(map[string][]*metricEntry)
	for _, entry := range r.metrics {
		grouped[entry.name] = append(grouped[entry.name], entry)
	}

	names := make([]string, 0, len(grouped))
	for name := range grouped {
		names = append(names, name)
	}
	sort.Strings(names)

	var sb strings.Builder

	for _, name := range names {
		entries := grouped[name]
		if len(entries) == 0 {
			continue
		}

		first := entries[0]
		if first.help != "" {
			fmt.Fprintf(&sb, "# HELP %s %s\n", name, first.help)
		}
		fmt.Fprintf(&sb, "# TYPE %s %s\n", name, first.metricType)

		for _, entry := range entries {
			labelStr := entry.labels.Format()

			switch entry.metricType {
			case TypeCounter:
				fmt.Fprintf(&sb, "%s%s %g\n", name, labelStr, entry.counter.Value())

			case TypeGauge:
				fmt.Fprintf(&sb, "%s%s %g\n", name, labelStr, entry.gauge.Value())

			case TypeHistogram:
				count, sum, buckets, counts := entry.histogram.Snapshot()

				for i, b := range buckets {
					bLabels := make(Labels, len(entry.labels)+1)
					for k, v := range entry.labels {
						bLabels[k] = v
					}
					bLabels["le"] = fmt.Sprintf("%g", b)
					fmt.Fprintf(&sb, "%s_bucket%s %d\n", name, bLabels.Format(), counts[i])
				}

				// +Inf bucket
				infLabels := make(Labels, len(entry.labels)+1)
				for k, v := range entry.labels {
					infLabels[k] = v
				}
				infLabels["le"] = "+Inf"
				fmt.Fprintf(&sb, "%s_bucket%s %d\n", name, infLabels.Format(), count)

				fmt.Fprintf(&sb, "%s_sum%s %g\n", name, labelStr, sum)
				fmt.Fprintf(&sb, "%s_count%s %d\n", name, labelStr, count)
			}
		}
	}

	return sb.String()
}
