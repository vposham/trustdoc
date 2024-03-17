package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAll(t *testing.T) {

	t.Run("GetAllWithoutLoad", func(t *testing.T) {
		assert.Nil(t, GetAll())
	})

	t.Run("GetAllSuccess", func(t *testing.T) {
		t.Setenv(envDetectorKey, "test")
		assert.NoError(t, Load(context.Background(), "../config"))
		assert.NotNil(t, GetAll())
	})
}
