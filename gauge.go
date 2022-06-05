package metrics

const kindGauge = ":gauge"

// Gauge implements a numerical value that can be set
// directly. Use a gauge for measured values like memory
// usage or active requests.
type Gauge struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Value float64 `json:"value"`
	Count uint64  `json:"count"`
}

// NewGauge returns a new gauge intialized at value.
func NewGauge(value float64) *Gauge {
	m := &Gauge{}
	if value != 0 {
		m.Set(value)
	}
	return m
}

// Set sets the gauge value.
func (m *Gauge) Set(value float64) {
	if value < m.Min || m.Count == 0 {
		m.Min = value
	}
	if value > m.Max || m.Count == 0 {
		m.Max = value
	}
	m.Value = value
	m.Count++
}
