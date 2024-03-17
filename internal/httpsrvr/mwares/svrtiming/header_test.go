package svrtiming

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// headerCases contains test cases for the Server-Timing header. This set
// of test cases is used to test both parsing the header value as well as
// generating the correct header value.
var headerCases = []struct {
	Metrics     []*Metric
	HeaderValue string
}{
	{
		Metrics: []*Metric{
			{
				Name:     "sql-1",
				Duration: 100 * time.Millisecond,
				Desc:     "MySQL lookup Server",
			},
		},
		HeaderValue: `sql-1;desc="MySQL lookup Server";dur=100`,
	},

	// Comma in description
	{
		Metrics: []*Metric{
			{
				Name:     "sql-1",
				Duration: 100 * time.Millisecond,
				Desc:     "MySQL, lookup Server",
			},
		},
		HeaderValue: `sql-1;desc="MySQL, lookup Server";dur=100`,
	},

	// Semicolon in description
	{
		Metrics: []*Metric{
			{
				Name:     "sql-1",
				Duration: 100 * time.Millisecond,
				Desc:     "MySQL; lookup Server",
			},
		},
		HeaderValue: `sql-1;desc="MySQL; lookup Server";dur=100`,
	},

	// Description that contains a number
	{
		Metrics: []*Metric{
			{
				Name:     "sql-1",
				Duration: 100 * time.Millisecond,
				Desc:     "GET 200",
			},
		},
		HeaderValue: `sql-1;desc="GET 200";dur=100`,
	},

	// Number that contains floating point
	{
		Metrics: []*Metric{
			{
				Name:     "sql-1",
				Duration: 100100 * time.Microsecond,
				Desc:     "MySQL; lookup Server",
			},
		},
		HeaderValue: `sql-1;desc="MySQL; lookup Server";dur=100.1`,
	},
}

func TestHeaderString(t *testing.T) {
	for _, tt := range headerCases {
		t.Run(tt.HeaderValue, func(t *testing.T) {
			h := &Header{Metrics: tt.Metrics}
			actual := h.String()
			assert.Equal(t, actual, tt.HeaderValue)
		})
	}
}

// Same as TestHeaderString but using the Add method
func TestHeaderAdd(t *testing.T) {
	for _, tt := range headerCases {
		t.Run(tt.HeaderValue, func(t *testing.T) {
			var h Header
			for _, m := range tt.Metrics {
				h.Add(m)
			}
			actual := h.String()
			assert.Equal(t, actual, tt.HeaderValue)
		})
	}
}

func TestNewMetric(t *testing.T) {
	var h Header
	m := h.NewMetric("sql")
	assert.NotNil(t, m, "expected non nil")
}

func TestNilHeader(t *testing.T) {
	var h *Header
	assert.Equal(t, h.String(), "")
	m := h.Add(nil)
	assert.Nil(t, m, "expected nil")
}
