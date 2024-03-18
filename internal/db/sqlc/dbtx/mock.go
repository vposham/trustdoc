package dbtx

import "context"

// MockStore struct provides all the mocked business DB transactions. It implements StoreIf
type MockStore struct {
	saveDocMetaFn         func(ctx context.Context, in DocMeta) error
	getDocMetaFn          func(ctx context.Context, docId string) (DocMeta, error)
	getDocMetaByDocHashFn func(ctx context.Context, docMd5Hash string) (DocMeta, error)
}

var _ StoreIf = (*MockStore)(nil)

// SaveDocMeta - mock implementation of it for unit testing
func (m MockStore) SaveDocMeta(ctx context.Context, in DocMeta) error {
	if m.saveDocMetaFn != nil {
		return m.saveDocMetaFn(ctx, in)
	}
	return nil
}

// GetDocMeta - mock implementation of it for unit testing
func (m MockStore) GetDocMeta(ctx context.Context, docId string) (DocMeta, error) {
	if m.getDocMetaFn != nil {
		return m.getDocMetaFn(ctx, docId)
	}
	return DocMeta{}, nil
}

func (m MockStore) GetDocMetaByHash(ctx context.Context, docMd5Hash string) (DocMeta, error) {
	if m.getDocMetaByDocHashFn != nil {
		return m.getDocMetaByDocHashFn(ctx, docMd5Hash)
	}
	return DocMeta{}, nil
}
