package metrics

const kindHistogram = ":histogram"

// Histogram represents a distribution of metrics that count
// the number of values that fall within configured buckets.
type Histogram struct {
	Min     float64  `json:"min"`
	Max     float64  `json:"max"`
	Count   uint64   `json:"count"`
	Dropped uint64   `json:"dropped"`
	Buckets []Bucket `json:"buckets"`
}

// NewHistogram returns a new histogram initialized with buckets.
// The default latency buckets will be used if buckets is nil.
func NewHistogram(buckets []Bucket) *Histogram {
	if buckets == nil {
		buckets = NewDefaultLatencyBuckets()
	}
	m := &Histogram{Buckets: buckets}
	for _, b := range buckets {
		m.Count += b.Count
	}
	return m
}

// Put adds value as a sample.
func (m *Histogram) Put(value float64) {
	if value < m.Min || m.Count == 0 {
		m.Min = value
	}
	if value > m.Max || m.Count == 0 {
		m.Max = value
	}
	m.Count++
	for i := range m.Buckets {
		if value <= m.Buckets[i].Value {
			m.Buckets[i].Count++
			return
		}
	}
	m.Dropped++
}

// Percentile returns the value below which a given
// percentage of the samples fall.
func (m Histogram) Percentile(p float64) float64 {
	x0 := 0.0
	sum := uint64(0)
	for _, b := range m.Buckets {
		y0 := float64(sum) / float64(m.Count)
		y1 := float64(sum+b.Count) / float64(m.Count)
		if y1 >= p {
			x1 := b.Value
			return x0 + (x1-x0)*((p-y0)/(y1-y0))
		}
		x0 = b.Value
		sum += b.Count
	}
	return 0
}
