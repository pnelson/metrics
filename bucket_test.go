package metrics

import (
	"reflect"
	"testing"
)

func TestNewDefaultLatencyBuckets(t *testing.T) {
	have := NewDefaultLatencyBuckets()
	want := defaultLatencyBucketValues
	if len(have) != len(want) {
		t.Fatalf("length\nhave %v\nwant %v", len(have), len(want))
	}
	if have[0].Value != want[0] {
		t.Fatalf("at 0\nhave %v\nwant %v", have[0].Value, want[0])
	}
	i := len(want) - 1
	if have[i].Value != want[i] {
		t.Fatalf("at %d\nhave %v\nwant %v", i, have[i].Value, want[i])
	}
}

func TestNewLinearBuckets(t *testing.T) {
	tests := []struct {
		offset float64
		width  float64
		count  int
		want   []Bucket
	}{
		{5, 15, 4, []Bucket{{5, 0}, {20, 0}, {35, 0}, {50, 0}}},
	}
	for _, tt := range tests {
		have := NewLinearBuckets(tt.offset, tt.width, tt.count)
		if !reflect.DeepEqual(have, tt.want) {
			t.Fatalf("mismatched buckets\nhave %v\nwant %v", have, tt.want)
		}
	}
}

// https://cloud.google.com/logging/docs/logs-based-metrics/distribution-metrics
func TestNewExponentialBuckets(t *testing.T) {
	tests := []struct {
		scale  float64
		growth float64
		count  int
		want   []Bucket
	}{
		{3, 2, 4, []Bucket{{3, 0}, {6, 0}, {12, 0}, {24, 0}}},
		{1, 1.5, 6, []Bucket{{1, 0}, {1.5, 0}, {2.25, 0}, {3.375, 0}, {5.0625, 0}, {7.59375, 0}}},
	}
	for _, tt := range tests {
		have := NewExponentialBuckets(tt.scale, tt.growth, tt.count)
		if !reflect.DeepEqual(have, tt.want) {
			t.Fatalf("NewExponentialBuckets(%v, %v, %d)\nhave %v\nwant %v", tt.scale, tt.growth, tt.count, have, tt.want)
		}
	}
}
