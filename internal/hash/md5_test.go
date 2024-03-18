package hash

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/log"
)

func TestMain(m *testing.M) {
	ce := os.Getenv("appEnv")
	defer func() {
		_ = os.Setenv("appEnv", ce)
	}()
	_ = os.Setenv("appEnv", "test")
	ctx := context.Background()
	_ = config.Load(ctx, "../../config")
	_ = log.Load(ctx)

	os.Exit(m.Run())
}

func TestMd5_Hash(t *testing.T) {
	type args struct {
		ctx context.Context
		in  io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				in:  strings.NewReader("hello world"),
			},
			want:    "5eb63bbbe01eeed093cb22bb8f5acdc3",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Md5{}
			got, err := m.Hash(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hash() got = %v, want %v", got, tt.want)
			}
		})
	}
}
