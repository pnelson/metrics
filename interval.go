package metrics

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Window represents the entire window of intervals.
type Window struct {
	Duration  int        `json:"duration"`
	Intervals []Interval `json:"intervals"`
}

// Interval represents an aggregated interval for the window.
type Interval struct {
	mu      sync.RWMutex
	time    time.Time
	metrics map[string]any
}

// newInterval returns a new interval of aggregated metrics.
func newInterval(t time.Time) *Interval {
	return &Interval{
		time:    t,
		metrics: make(map[string]any),
	}
}

// Time returns the interval time.
func (i *Interval) Time() time.Time {
	return i.time
}

// Counter returns the counter at key if it exists.
func (i *Interval) Counter(key []string) Counter {
	k := keyPath(key) + kindCounter
	m := Counter{}
	v, ok := i.metrics[k]
	if !ok {
		return m
	}
	c := v.(*Counter)
	m.Value = c.Value
	m.Count = c.Count
	return m
}

// Gauge returns the gauge at key if it exists.
func (i *Interval) Gauge(key []string) Gauge {
	k := keyPath(key) + kindGauge
	m := Gauge{}
	v, ok := i.metrics[k]
	if !ok {
		return m
	}
	g := v.(*Gauge)
	m.Min = g.Min
	m.Max = g.Max
	m.Value = g.Value
	m.Count = g.Count
	return m
}

// Histogram returns the histogram at key if it exists.
func (i *Interval) Histogram(key []string) Histogram {
	k := keyPath(key) + kindHistogram
	m := Histogram{}
	v, ok := i.metrics[k]
	if !ok {
		return m
	}
	h := v.(*Histogram)
	m.Min = h.Min
	m.Max = h.Max
	m.Count = h.Count
	m.Dropped = h.Dropped
	m.Buckets = make([]Bucket, len(h.Buckets))
	copy(m.Buckets, h.Buckets)
	return m
}

// MarshalJSON implements the json.Marshaler interface.
func (i *Interval) MarshalJSON() ([]byte, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	return json.Marshal(struct {
		Time    int64          `json:"time"`
		Metrics map[string]any `json:"metrics"`
	}{
		Time:    i.time.UnixMilli(),
		Metrics: i.metrics,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (i *Interval) UnmarshalJSON(b []byte) error {
	m := struct {
		Time    int64                      `json:"time"`
		Metrics map[string]json.RawMessage `json:"metrics"`
	}{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	metrics := make(map[string]any)
	for k, v := range m.Metrics {
		n := strings.Index(k, ":")
		switch k[n:] {
		case kindCounter:
			metrics[k] = new(Counter)
		case kindGauge:
			metrics[k] = new(Gauge)
		case kindHistogram:
			metrics[k] = new(Histogram)
		default:
			panic(fmt.Errorf("metrics: unexpected metric type '%s'", k[n+1:]))
		}
		err = json.Unmarshal(v, metrics[k])
		if err != nil {
			return err
		}
	}
	i.time = time.UnixMilli(m.Time)
	i.metrics = metrics
	return nil
}
