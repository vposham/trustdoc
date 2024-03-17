package svrtiming

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetric_startStop(t *testing.T) {
	var m Metric
	m.WithDesc("test").Start()
	time.Sleep(50 * time.Millisecond)
	m.Stop()

	actual := m.Duration
	assert.NotEqual(t, actual, 0, "duration should be set")
	assert.Less(t, actual, 100*time.Millisecond, "expected duration to be within 100ms")
	assert.Greater(t, actual, 30*time.Millisecond, "expected duration to be more than 30ms")
	assert.Contains(t, m.String(), "test", "expected to have test")
}

func TestMetric_stopNoStart(t *testing.T) {
	var m Metric
	m.Stop()
	actual := m.Duration
	assert.Equal(t, int(actual), 0, "duration should not be set")
}

func TestMetric_Print(t *testing.T) {
	m := &Metric{}
	s := fmt.Sprintf("%#v format", m)
	assert.Contains(t, s, "Duration:0", "duration should be 0")
	m = nil
	s = fmt.Sprintf("%#v format", m)
	assert.Contains(t, s, "nil", "metric should be nil")
}
