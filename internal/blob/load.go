// Package blob will contain the interfaces and implementations of blob store
package blob

import (
	"context"
	"fmt"
	"sync"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/vposham/trustdoc/config"
)

var (
	onceInit      = new(sync.Once)
	concreteImpls = make(map[string]any)
)

const (
	// blobExecKey implementation makes calls to blob store
	blobExecKey = "MinioConfiguredExecKey"
)

// Load enables us inject this package as dependency from its parent
func Load(ctx context.Context) error {
	var appErr error
	onceInit.Do(func() {
		appErr = loadImpls(ctx)
	})
	return appErr
}

func loadImpls(_ context.Context) error {
	props := config.GetAll()
	if concreteImpls[blobExecKey] == nil {
		blobStoreUrl := props.MustGetString("minio.endpoint.url")
		keyId := props.MustGetString("minio.access.key.id")
		secretAccessKey := props.MustGetString("minio.access.secret.key")
		useSsl := props.MustGetBool("minio.use.ssl")
		bucketName := props.MustGetString("minio.app.bucket.name")
		minioClient, err := minio.New(blobStoreUrl, &minio.Options{
			Creds:  credentials.NewStaticV4(keyId, secretAccessKey, ""),
			Secure: useSsl,
		})
		if err != nil {
			return fmt.Errorf("failed to create minio client - %w", err)
		}
		var blobExec OpsIf = &Minio{
			bucketName: bucketName,
			client:     minioClient,
		}
		concreteImpls[blobExecKey] = blobExec
	}
	return nil
}

// GetBlobStore is used to get blob store
func GetBlobStore() OpsIf {
	targetImpl := blobExecKey
	v := concreteImpls[targetImpl]
	return v.(OpsIf)
}
