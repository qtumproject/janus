package server

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

func pingPong(ctx context.Context, ws *websocket.Conn, writeMutex *sync.Mutex) func() {
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	ticker := time.NewTicker(pingPeriod)

	go func(ctx context.Context, t *time.Ticker) {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				writeMutex.Lock()
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				err := ws.WriteMessage(websocket.PingMessage, nil)
				writeMutex.Unlock()
				if err != nil {
					ws.Close()
					return
				}
			}
		}
	}(ctx, ticker)

	return ticker.Stop
}
