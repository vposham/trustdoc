package dbtx

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	_ = config.Load(ctx, "../../../resources")
	_ = log.Load(ctx)
	_ = Load(ctx)

	os.Exit(m.Run())
}

// skipCI is in place to skip unit tests in environments  where local DB isnt feasible
// todo use in future
func skipCI(t *testing.T) {

	timeout := time.Second
	dbHost := "localhost"
	dbPort := "5432"

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(dbHost, dbPort), timeout)
	if err != nil {
		t.Skip("skipping " + t.Name() + " unit test as connecting on " + dbHost + " on port" +
			dbPort + "resulted in " + err.Error())
	}
	if conn != nil {
		defer func(conn net.Conn) {
			_ = conn.Close()
		}(conn)
		_, _ = fmt.Println("running unit test with db as ", net.JoinHostPort(dbHost, dbPort))
	}
}

func TestGetDbStore(t *testing.T) {
	assert.NotNil(t, GetDbStore())
}
