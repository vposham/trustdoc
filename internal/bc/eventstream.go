package bc

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vposham/trustdoc/log"
	"go.uber.org/zap"
)

func (k *Kaleido) ListenForEvents(ctx context.Context) {
	logger := log.GetLogger(ctx)
	logger.Info("listening for events")
	c, _, err := websocket.DefaultDialer.Dial(k.wsUrl, nil)
	if err != nil {
		logger.Fatal("dial wss err:", zap.Error(err))
	}
	defer c.Close()

	done := make(chan struct{})

	// Concurrently handle incoming messages
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logger.Info("read:", zap.Error(err))
				return
			}
			logger.Info("message received", zap.Any("msg", message))
		}
	}()

	// Send a message every 3 seconds
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			message := `{"type":"listen"}`
			err := c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				logger.Error("writeErr:", zap.Error(err))
			}
		}
	}
}
