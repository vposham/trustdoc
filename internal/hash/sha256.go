package hash

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"go.uber.org/zap"

	"github.com/vposham/trustdoc/log"
)

// Sha256 holds any config which is needed for Hasher implementation
type Sha256 struct{}

// _ maintain this line to force a compilation error when Sha256 does not implement OpsIf
var _ Hasher = (*Sha256)(nil)

// Hash generates sha256 hash of the input data
func (s Sha256) Hash(ctx context.Context, in io.Reader) (string, error) {
	logger := log.GetLogger(ctx)
	hash := sha256.New()
	n, err := io.Copy(hash, in)
	if err != nil {
		return "", err
	}
	out := hex.EncodeToString(hash.Sum(nil))
	logger.Info("sha256 hash generated", zap.Int64("bytesHashed", n), zap.String("hash", out))
	return out, nil
}
