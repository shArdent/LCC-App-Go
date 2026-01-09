package server

import (
	"time"
"github.com/gorilla/websocket")

type Client struct {
	ID   string
	Conn *websocket.Conn
}

func Broadcast(event string, data any) {
	Mutex.Lock()
	defer Mutex.Unlock()

	payload := map[string]any{
		"event": event,
		"data":  data,
	}

	for _, c := range Clients {
		c.Conn.WriteJSON(payload)
	}
}

func NewClientID() string {
	return time.Now().Format("150405.000000")
}
