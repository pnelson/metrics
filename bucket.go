package metrics

import (
	"fmt"
	"math"
)

// Bucket is a count of samples that fall between the
// range of the values from the previous bucket, if any,
// up to and including the value of this bucket.
type Bucket struct {
	Value float64
	Count uint64
}

// MarshalJSON implements the json.Marshaler interface.
func (b Bucket) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("[%v,%d]", b.Value, b.Count)), nil
}

// defaultLatencyBucketValues represents the default bucket
// values for measuring response time latency.
var defaultLatencyBucketValues = []float64{1, 3, 5, 7, 10, 15, 20, 25, 30, 35, 40, 45, 50, 60, 70, 80, 90, 100, 125, 150, 175, 200, 225, 250, 275, 300, 350, 400, 450, 500, 600, 700, 800, 900, 1000, 1250, 1500, 1750, 2000, 2250, 2500, 2750, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000}

// NewDefaultLatencyBuckets returns a set of buckets suitable
// for measuring response time latency. Buckets are skewed
// with lower widths towards smaller values for greater
// accuracy at the levels where it matters, and gradually
// higher widths at already generally unacceptable values.
func NewDefaultLatencyBuckets() []Bucket {
	buckets := make([]Bucket, len(defaultLatencyBucketValues))
	for i, v := range defaultLatencyBucketValues {
		buckets[i] = Bucket{Value: v}
	}
	return buckets
}

// NewLinearBuckets returns a set of buckets with equal width.
func NewLinearBuckets(offset, width float64, count int) []Bucket {
	buckets := make([]Bucket, count)
	for i := 0; i < count; i++ {
		buckets[i] = Bucket{Value: offset + width*float64(i)}
	}
	return buckets
}

// NewExponentialBuckets returns a set of buckets with widths
// increasing for higher values.
func NewExponentialBuckets(scale, growth float64, count int) []Bucket {
	buckets := make([]Bucket, count)
	for i := 0; i < count; i++ {
		buckets[i] = Bucket{Value: scale * math.Pow(growth, float64(i))}
	}
	return buckets
}
