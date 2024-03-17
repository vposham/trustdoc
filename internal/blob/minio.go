package blob

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/vposham/trustdoc/log"
	"go.uber.org/zap"
)

// Minio blob store implementation
type Minio struct {
	bucketName string
	client     *minio.Client
}

// Put uploads a document to the Minio blob store
func (m *Minio) Put(ctx context.Context, doc io.Reader, size int64) (docId string, err error) {
	logger := log.GetLogger(ctx)
	docId = uuid.New().String()
	logger.Info("started uploading document", zap.String("docId", docId))
	up, err := m.client.PutObject(ctx, m.bucketName, docId, doc, size, minio.PutObjectOptions{SendContentMd5: true})
	if err != nil {
		err = fmt.Errorf("failed to upload - %w", err)
		logger.Error("failed to upload document", zap.String("docId", docId), zap.Error(err))
		return
	}
	logger.Info("completed uploading document", zap.String("docId", docId), zap.Any("uploadInfo", up))
	return
}

// Get downloads a document from the Minio blob store
func (m *Minio) Get(ctx context.Context, docId string) (doc io.Reader, err error) {
	logger := log.GetLogger(ctx)
	logger.Info("started downloading document", zap.String("docId", docId))
	obj, err := m.client.GetObject(ctx, m.bucketName, docId, minio.GetObjectOptions{})
	if err != nil {
		err = fmt.Errorf("failed to download - %w", err)
		logger.Error("failed to download document", zap.String("docId", docId), zap.Error(err))
		return
	}
	return obj, nil
}
