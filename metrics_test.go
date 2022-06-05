package metrics_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/pnelson/metrics"
)

var (
	testWindow   = 100 * time.Millisecond
	testInterval = 10 * time.Millisecond
)

func TestMetrics(t *testing.T) {
	m := metrics.New(testWindow, testInterval)
	m.Add([]string{"c"}, 1)
	m.Set([]string{"g"}, 1)
	m.Put([]string{"s"}, 1)
	w := m.Window()
	if len(w.Intervals) != 1 {
		t.Fatalf("should return the current interval at minimum")
	}
}

func TestMetricsAdd(t *testing.T) {
	m := metrics.New(testWindow, testInterval)
	m.Add([]string{"test"}, 1)
	m.Add([]string{"test"}, 2)
	m.Add([]string{"test"}, 3)
	m.Add([]string{"test"}, 5)
	m.Add([]string{"test"}, 8)
	w := m.Window()
	if len(w.Intervals) != 1 {
		t.Fatalf("should only return the current interval")
	}
	i := w.Intervals[0]
	c := i.Counter([]string{"test"})
	if c.Value != 19.0 {
		t.Fatalf("Value\nhave %f\nwant %f", c.Value, 19.0)
	}
	if c.Count != 5 {
		t.Fatalf("Count\nhave %d\nwant %d", c.Count, 5)
	}
	if i.Gauge([]string{"test"}).Count != 0 {
		t.Fatalf("should return a zero gauge")
	}
	if i.Histogram([]string{"test"}).Count != 0 {
		t.Fatalf("should return a zero sample")
	}
}

func TestMetricsSet(t *testing.T) {
	m := metrics.New(testWindow, testInterval)
	m.Set([]string{"test"}, 1)
	m.Set([]string{"test"}, 5)
	m.Set([]string{"test"}, 2)
	m.Set([]string{"test"}, 8)
	m.Set([]string{"test"}, 3)
	w := m.Window()
	if len(w.Intervals) != 1 {
		t.Fatalf("should only return the current interval")
	}
	i := w.Intervals[0]
	g := i.Gauge([]string{"test"})
	if g.Min != 1.0 {
		t.Fatalf("Min\nhave %f\nwant %f", g.Min, 1.0)
	}
	if g.Max != 8.0 {
		t.Fatalf("Max\nhave %f\nwant %f", g.Max, 8.0)
	}
	if g.Value != 3.0 {
		t.Fatalf("Value\nhave %f\nwant %f", g.Value, 3.0)
	}
	if g.Count != 5 {
		t.Fatalf("Count\nhave %d\nwant %d", g.Count, 5)
	}
	if i.Counter([]string{"test"}).Count != 0 {
		t.Fatalf("should return a zero count")
	}
	if i.Histogram([]string{"test"}).Count != 0 {
		t.Fatalf("should return a zero sample")
	}
}

func TestMetricsMod(t *testing.T) {
	m := metrics.New(testWindow, testInterval)
	m.Mod([]string{"test"}, 3)
	m.Mod([]string{"test"}, 2)
	m.Mod([]string{"test"}, -1)
	w := m.Window()
	if len(w.Intervals) != 1 {
		t.Fatalf("should only return the current interval")
	}
	i := w.Intervals[0]
	g := i.Gauge([]string{"test"})
	if g.Value != 4.0 {
		t.Fatalf("Value\nhave %f\nwant %f", g.Value, 4.0)
	}
	<-time.After(3 * testInterval)
	m.Mod([]string{"test"}, 3)
	w = m.Window()
	if len(w.Intervals) < 2 {
		t.Fatalf("should have several intervals")
	}
	i = w.Intervals[len(w.Intervals)-1]
	g = i.Gauge([]string{"test"})
	if g.Min != 7.0 {
		t.Fatalf("Min\nhave %f\nwant %f", g.Min, 7.0)
	}
	if g.Max != 7.0 {
		t.Fatalf("Max\nhave %f\nwant %f", g.Max, 7.0)
	}
	if g.Value != 7.0 {
		t.Fatalf("Value\nhave %f\nwant %f", g.Value, 7.0)
	}
	if g.Count != 1 {
		t.Fatalf("Count\nhave %d\nwant %d", g.Count, 1)
	}
}

func TestMetricsPut(t *testing.T) {
	m := metrics.New(testWindow, testInterval)
	m.Put([]string{"test"}, 1)
	m.Put([]string{"test"}, 2)
	m.Put([]string{"test"}, 3)
	m.Put([]string{"test"}, 30000)
	m.Put([]string{"test"}, 4)
	m.Put([]string{"test"}, 5)
	m.Put([]string{"test"}, 6)
	w := m.Window()
	if len(w.Intervals) != 1 {
		t.Fatalf("should only return the current interval")
	}
	i := w.Intervals[0]
	h := i.Histogram([]string{"test"})
	if h.Min != 1.0 {
		t.Fatalf("Min\nhave %f\nwant %f", h.Min, 1.0)
	}
	if h.Max != 30000.0 {
		t.Fatalf("Max\nhave %f\nwant %f", h.Max, 30000.0)
	}
	if h.Count != 7 {
		t.Fatalf("Count\nhave %d\nwant %d", h.Count, 7)
	}
	if h.Dropped != 1 {
		t.Fatalf("Dropped\nhave %d\nwant %d", h.Dropped, 1)
	}
	buckets := []metrics.Bucket{
		{1.0, 1},
		{3.0, 2},
		{5.0, 2},
		{7.0, 1},
		{10.0, 0},
	}
	for i, b := range buckets {
		if h.Buckets[i] != b {
			t.Fatalf("Buckets[%d]\nhave %v\nwant %v", i, h.Buckets[i], b)
		}
	}
	if i.Counter([]string{"test"}).Count != 0 {
		t.Fatalf("should return a zero count")
	}
	if i.Gauge([]string{"test"}).Count != 0 {
		t.Fatalf("should return a zero gauge")
	}
}

func TestMetricsBuckets(t *testing.T) {
	m := metrics.New(testWindow, testInterval)
	m.Buckets([]string{"test"}, metrics.NewLinearBuckets(1, 1, 3))
	m.Put([]string{"test"}, 1)
	m.Put([]string{"test"}, 2)
	m.Put([]string{"test"}, 3)
	m.Put([]string{"test"}, 4)
	m.Put([]string{"test"}, 5)
	w := m.Window()
	if len(w.Intervals) != 1 {
		t.Fatalf("should only return the current interval")
	}
	i := w.Intervals[0]
	h := i.Histogram([]string{"test"})
	if h.Count != 5 {
		t.Fatalf("Count\nhave %d\nwant %d", h.Count, 5)
	}
	if h.Dropped != 2 {
		t.Fatalf("Dropped\nhave %d\nwant %d", h.Dropped, 2)
	}
}

func TestMetricsWindow(t *testing.T) {
	m := metrics.New(testWindow, testInterval)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case c := <-time.After(time.Millisecond): // 10x per interval
				m.Set([]string{"test"}, float64(c.UnixMilli()))
			case <-done:
				return
			}
		}
	}()
	select {
	case <-time.After(testWindow + 3*testInterval):
		close(done)
	}
	w := m.Window()
	if len(w.Intervals) != int(testWindow/testInterval) {
		t.Fatalf("should cap to window")
	}
	for i := 1; i < len(w.Intervals)-1; i++ {
		v1 := w.Intervals[i-1].Gauge([]string{"test"}).Value
		v2 := w.Intervals[i].Gauge([]string{"test"}).Value
		if v1 == 0 || v2 < v1 {
			t.Fatalf("should have increasing values i=%d\nv1=%d\nv2=%d", i, uint64(v1), uint64(v2))
		}
	}
}

func TestMetricsWindowMarshal(t *testing.T) {
	m := metrics.New(testWindow, testInterval)
	m.Set([]string{"test"}, 1)
	want := m.Window()
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	have := metrics.Window{}
	err = json.Unmarshal(b, &have)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("json\nhave %v\nwant %v", have, want)
	}
}
