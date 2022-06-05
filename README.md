# metrics

Package metrics aggregates metrics over regular intervals.

## Usage

Initialize a new metrics instance with 10 second aggregated intervals over a 1
hour window. Optionally emit runtime stats.

```go
m := metrics.New(time.Hour, 10*time.Second)
go m.MemStats(time.Second)
```

Use `Add` for monotonically increasing `Counter` values such as errors or
execution counts.

```go
m.Add([]string{"errors"}, 1)
```

Use `Set` for measured `Gauge` values such as memory usage or active requests.

```go
var n uint64
m.Set([]string{"active"}, float64(atomic.AddUint64(&n, 1)))
defer func() { m.Set([]string{"active"}, float64(atomic.AddUint64(&n, ^uint64(0)))) }()
```

Use `Mod` to simplify a `Gauge` that increments or decrements like the above.
The gauge will increment or decrement from zero if there is no previous value
within the entire stored window.

```go
m.Mod([]string{"active"}, 1)
defer m.Mod([]string{"active"}, -1)
```

Use `Put` for sample values distributed over the `Histogram` bucket values such
as latency. The `Histogram` also stores a count of the samples.

```go
m.Put([]string{"latency"}, 1)
```

Use `Timer` as a shorthand for `Put` to sample the duration elapsed in
milliseconds from a starting `time.Time` value.

```go
func (m *M) Method() {
  defer m.Timer([]string{"method", "latency"}, time.Now())
  // ...
}
```
