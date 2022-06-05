package metrics

import (
	"path"
	"runtime"
	"sort"
	"sync"
	"time"
)

// Metrics represents the central manager of metrics activity.
type Metrics struct {
	mu        sync.RWMutex
	window    time.Duration
	interval  time.Duration
	buckets   map[string][]Bucket
	intervals []*Interval
}

// New returns a new metrics manager.
func New(window, interval time.Duration) *Metrics {
	m := &Metrics{
		window:    window,
		interval:  interval,
		buckets:   make(map[string][]Bucket),
		intervals: make([]*Interval, 1, window/interval),
	}
	t := time.Now().Truncate(interval).Add(interval)
	m.intervals[0] = newInterval(t)
	// Force create every interval.
	go func() {
		for {
			select {
			case <-time.After(interval):
				t = t.Add(interval)
				i := newInterval(t)
				m.mu.Lock()
				if len(m.intervals) == cap(m.intervals) {
					copy(m.intervals, m.intervals[1:])
					m.intervals[len(m.intervals)-1] = i
				} else {
					m.intervals = append(m.intervals, i)
				}
				m.mu.Unlock()
			}
		}
	}()
	return m
}

// Add adds value to key.
func (m *Metrics) Add(key []string, value float64) {
	k := keyPath(key) + kindCounter
	m.mu.RLock()
	defer m.mu.RUnlock()
	i := m.intervals[len(m.intervals)-1]
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.metrics[k]
	if !ok {
		i.metrics[k] = NewCounter(value)
		return
	}
	v.(*Counter).Add(value)
}

// Set sets key to value.
func (m *Metrics) Set(key []string, value float64) {
	k := keyPath(key) + kindGauge
	m.mu.RLock()
	defer m.mu.RUnlock()
	i := m.intervals[len(m.intervals)-1]
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.metrics[k]
	if !ok {
		i.metrics[k] = NewGauge(value)
		return
	}
	v.(*Gauge).Set(value)
}

// Mod modifies key by relative value. Use this for gauges that increment
// or decrement instead of maintaining a persistent value. The gauge will
// increment or decrement from zero if there is no previous value within
// the entire stored window.
func (m *Metrics) Mod(key []string, value float64) {
	k := keyPath(key) + kindGauge
	m.mu.RLock()
	defer m.mu.RUnlock()
	i := m.intervals[len(m.intervals)-1]
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.metrics[k]
	if !ok {
		prev := m.prevGaugeValue(k)
		i.metrics[k] = NewGauge(prev + value)
		return
	}
	g := v.(*Gauge)
	g.Set(g.Value + value)
}

func (m *Metrics) prevGaugeValue(key string) float64 {
	if len(m.intervals) < 2 {
		return 0
	}
	for n := len(m.intervals) - 2; n >= 0; n-- {
		i := m.intervals[n]
		i.mu.RLock()
		v, ok := i.metrics[key]
		if !ok {
			i.mu.RUnlock()
			continue
		}
		prev := v.(*Gauge).Value
		i.mu.RUnlock()
		return prev
	}
	return 0
}

// Put adds value as a sample for key.
func (m *Metrics) Put(key []string, value float64) {
	k := keyPath(key) + kindHistogram
	m.mu.RLock()
	defer m.mu.RUnlock()
	i := m.intervals[len(m.intervals)-1]
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.metrics[k]
	if !ok {
		buckets := m.bucketsFor(k)
		v := NewHistogram(buckets)
		v.Put(value)
		i.metrics[k] = v
		return
	}
	v.(*Histogram).Put(value)
}

// Timer adds the elapsed duration in milliseconds as a sample for key.
func (m *Metrics) Timer(key []string, t time.Time) {
	ms := time.Since(t).Milliseconds()
	m.Put(key, float64(ms))
}

// Buckets sets the initial buckets for histograms at the key prefix.
func (m *Metrics) Buckets(key []string, buckets []Bucket) {
	k := keyPath(key)
	m.mu.Lock()
	m.buckets[k] = buckets
	m.mu.Unlock()
}

// bucketsFor returns the histogram buckets using
// a longest prefix match from the configured buckets.
func (m *Metrics) bucketsFor(s string) []Bucket {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.buckets) == 0 {
		return NewDefaultLatencyBuckets()
	}
	keys := make([]string, 0, len(m.buckets))
	for k := range m.buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	longestPrefix := ""
	for _, k := range keys {
		n := 0
		for i := range k {
			if i >= len(s) || k[i] != s[i] {
				break
			}
			n++
		}
		if n > len(longestPrefix) {
			longestPrefix = k
		}
	}
	if longestPrefix == "" {
		return NewDefaultLatencyBuckets()
	}
	return m.buckets[longestPrefix]
}

// MemStats records runtime memory allocator metric values at interval d.
func (m *Metrics) MemStats(d time.Duration) {
	var stats runtime.MemStats
	for {
		select {
		case <-time.After(d):
			runtime.ReadMemStats(&stats)
			m.Set([]string{"memstats", "goroutines"}, float64(runtime.NumGoroutine())) // count
			m.Set([]string{"memstats", "total"}, float64(stats.Sys))                   // bytes
			m.Set([]string{"memstats", "alloc"}, float64(stats.Alloc))                 // bytes
			m.Set([]string{"memstats", "count"}, float64(stats.HeapObjects))           // count
		}
	}
}

// Window returns a copy of the windowed intervals.
func (m *Metrics) Window() Window {
	m.mu.RLock()
	defer m.mu.RUnlock()
	view := Window{
		Duration:  int(m.window.Seconds()),
		Intervals: make([]Interval, len(m.intervals)),
	}
	for n, i := range m.intervals {
		metrics := make(map[string]any)
		// The current interval needs a read lock.
		// The defer won't run until the method returns
		// but it only runs on the last iteration.
		if n == len(m.intervals)-1 {
			i.mu.RLock()
			defer i.mu.RUnlock()
		}
		for k, v := range i.metrics {
			switch t := v.(type) {
			case *Counter:
				c := new(Counter)
				*c = *t
				metrics[k] = c
			case *Gauge:
				g := new(Gauge)
				*g = *t
				metrics[k] = g
			case *Histogram:
				h := new(Histogram)
				*h = *t
				metrics[k] = h
			default:
				panic("metrics: unexpected metric type")
			}
		}
		view.Intervals[n] = Interval{time: i.time, metrics: metrics}
	}
	return view
}

// keyPath returns a rooted path to the key.
func keyPath(key []string) string {
	if len(key) == 0 {
		panic("metrics: key required")
	}
	key[0] = "/" + key[0]
	return path.Join(key...)
}
