package hash

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"

	"go.uber.org/zap"

	"github.com/vposham/trustdoc/log"
)

type Md5 struct{}

var _ Hasher = (*Md5)(nil)

func (m Md5) Hash(ctx context.Context, in io.Reader) (string, error) {
	logger := log.GetLogger(ctx)
	hash := md5.New()
	n, err := io.Copy(hash, in)
	if err != nil {
		return "", err
	}
	out := hex.EncodeToString(hash.Sum(nil))
	logger.Info("md5 hash generated", zap.Int64("bytesHashed", n), zap.String("hash", out))
	return out, nil
}
