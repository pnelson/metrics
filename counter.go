package metrics

const kindCounter = ":counter"

// Counter implements a monotonically increasing counter that
// is reset to zero at the beginning of each interval. Use a
// counter for errors or execution counts.
type Counter struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Value float64 `json:"value"`
	Count uint64  `json:"count"`
}

// NewCounter returns a new counter initialized at value.
func NewCounter(value float64) *Counter {
	m := &Counter{}
	if value != 0 {
		m.Add(value)
	}
	return m
}

// Add adds the value.
func (m *Counter) Add(value float64) {
	if value < m.Min || m.Count == 0 {
		m.Min = value
	}
	if value > m.Max || m.Count == 0 {
		m.Max = value
	}
	m.Value += value
	m.Count++
}
