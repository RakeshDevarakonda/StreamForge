package websocket

import (
	"log"
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

	// sendBufferSize is the capacity of the binary video chunk channel
	sendBufferSize = 512
)

// Client represents a connected WebSocket viewer watching a stream
type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

// NewClient creates an initialized WebSocket client
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
		Send: make(chan []byte, sendBufferSize),
	}
}

// WritePump pumps video packets and heartbeats from the Send channel to the WebSocket connection
func (c *Client) WritePump(onClose func()) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
		if onClose != nil {
			onClose()
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send raw binary video frame chunk
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
			// Send heartbeat ping to keep connection alive through proxies/NATs
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
		_ = c.Conn.Close()
		if onClose != nil {
			onClose()
		}
	}()

	c.Conn.SetReadLimit(4096)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Read messages to process socket close / ping-pong controls
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Read error: %v", err)
			}
			break
		}
	}
}
