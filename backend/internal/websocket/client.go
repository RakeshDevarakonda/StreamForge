package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// pongWait is the time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// pingPeriod sends pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// sendBufferSize is the capacity of the binary video chunk channel (optimized for low memory usage)
	sendBufferSize = 64
)

// Client represents a connected WebSocket viewer watching a stream
type Client struct {
	Conn     *websocket.Conn
	Send     chan []byte
	closeOnce sync.Once
}

// NewClient creates an initialized WebSocket client
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
		Send: make(chan []byte, sendBufferSize),
	}
}

// Close ensures the connection and cleanup handler fire strictly once
func (c *Client) Close(onClose func()) {
	c.closeOnce.Do(func() {
		_ = c.Conn.Close()
		if onClose != nil {
			onClose()
		}
	})
}

// WritePump pumps video packets and heartbeats from the Send channel to the WebSocket connection
func (c *Client) WritePump(onClose func()) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close(onClose)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
			}
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump reads control messages (pong, close) from the WebSocket connection
func (c *Client) ReadPump(onClose func()) {
	defer func() {
		c.Close(onClose)
	}()

	c.Conn.SetReadLimit(4096)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Read error: %v", err)
			}
			break
		}
	}
}

