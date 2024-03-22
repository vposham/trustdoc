package dbtx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveDocMeta(t *testing.T) {
	mockStore := &MockStore{}

	// Test case 1: SaveDocMetaFn is not set
	err := mockStore.SaveDocMeta(context.Background(), DocMeta{})
	assert.NoError(t, err)

	// Test case 2: SaveDocMetaFn is set
	mockStore.saveDocMetaFn = func(_ context.Context, _ DocMeta) error {
		return nil
	}
	err = mockStore.SaveDocMeta(context.Background(), DocMeta{})
	assert.NoError(t, err)
}

func TestGetDocMeta(t *testing.T) {
	mockStore := &MockStore{}

	// Test case 1: GetDocMetaFn is not set
	result, err := mockStore.GetDocMeta(context.Background(), "123")
	assert.NoError(t, err)
	assert.Equal(t, DocMeta{}, result)

	// Test case 2: GetDocMetaFn is set
	expectedDocMeta := DocMeta{}
	mockStore.getDocMetaFn = func(_ context.Context, _ string) (DocMeta, error) {
		return expectedDocMeta, nil
	}
	result, err = mockStore.GetDocMeta(context.Background(), "123")
	assert.NoError(t, err)
	assert.Equal(t, expectedDocMeta, result)
}
