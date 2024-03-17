package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestBadLogConf(t *testing.T) {
	mock := ZapAppConf{}
	lgr := mock.CustomizeLogger(&zap.Config{
		EncoderConfig: zapcore.EncoderConfig{
			EncodeTime: nil,
			TimeKey:    "fcfhgv",
		}})
	assert.NotNil(t, lgr, "bad zap config should default zap prod config")
}
func TestFetchLogLevels(t *testing.T) {
	toTestLvls := []string{Debug, Info, Warn, Error, DPanic, Panic, Fatal, ""}
	for _, v := range toTestLvls {
		if fetchLogLevel(v).String() == "" {
			t.Errorf("bad zap config level should default to InfoLevel")
		}
	}
}
func TestDefaultConfig(t *testing.T) {
	got := DefaultConfig()
	if got == nil {
		t.Errorf("DefaultConfig() should give valid base zap config")
	}
}
