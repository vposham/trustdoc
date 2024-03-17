package svrtiming

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

const (
	hdrKey = "Server-Timing"

	// Specified server-timing-param-name values.

	paramNameDesc = "desc"
	paramNameDur  = "dur"
)

var reNumber = regexp.MustCompile(`^\d+\.?\d*$`)

// Header represents a collection of metrics that can be encoded as
// a Server-Timing header value.
//
// The functions for working with metrics are concurrency-safe to make
// it easy to record metrics from goroutines. If you want to avoid the
// lock overhead, you can access the Metrics field directly.
//
// The functions for working with metrics are also usable on a nil
// Header pointer. This allows functions that use FromContext to get the
// *Header value to skip nil-checking and use it as normal. On a nil
// *Header, Metrics are not recorded.
type Header struct {
	// Metrics is the list of metrics in the header.
	Metrics []*Metric

	// The lock that is held when Metrics is being modified. This
	// ONLY NEEDS TO BE SET WHEN working with Metrics directly. If using
	// the functions on the struct, the lock is managed automatically.
	sync.Mutex
}

// NewMetric creates a new Metric and adds it to this header.
func (h *Header) NewMetric(name string) *Metric {
	return h.Add(&Metric{Name: name})
}

// Add adds the given metric to the header.
//
// This function is safe to call concurrently.
func (h *Header) Add(m *Metric) *Metric {
	if h == nil {
		return m
	}

	h.Lock()
	defer h.Unlock()
	h.Metrics = append(h.Metrics, m)
	return m
}

// String returns the valid Server-Timing header value that can be
// sent in an HTTP response.
func (h *Header) String() string {
	if h == nil {
		return ""
	}
	parts := make([]string, 0, len(h.Metrics))
	for _, m := range h.Metrics {
		parts = append(parts, m.String())
	}

	return strings.Join(parts, ",")
}

// headerEncodeParam encodes a key/value pair as a proper `key=value`
// syntax, using double-quotes if necessary.
func headerEncodeParam(key, value string) string {
	// The only case we currently don't quote is numbers. We can make this
	// smarter in the future.
	if reNumber.MatchString(value) {
		return fmt.Sprintf(`%s=%s`, key, value)
	}

	return fmt.Sprintf(`%s=%q`, key, value)
}
