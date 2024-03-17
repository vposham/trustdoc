package svrtiming

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	h := new(Header)
	ctx := NewContext(context.Background(), h)
	h2 := FromContext(ctx)
	assert.Equal(t, h, h2, "should have stored value")
}

func TestContext_notSet(t *testing.T) {
	h := FromContext(context.Background())
	assert.Nil(t, h, "should be nil")
}
