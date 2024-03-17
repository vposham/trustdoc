package dbtx

import "context"

type DocMeta struct {
	DocId string
}

func (store *Store) SaveDocMeta(ctx context.Context, in DocMeta) error {
	// TODO implement me
	panic("implement me")
}

func (store *Store) GetDocMeta(ctx context.Context, docId string) (DocMeta, error) {
	// TODO implement me
	panic("implement me")
}
