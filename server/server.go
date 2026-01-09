package server

import (
	"log"
	"net/http"
	"os"
	"time"
	"path/filepath"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	id := NewClientID()

	Mutex.Lock()
	Clients[id] = Client{ID: id, Conn: conn}
	Mutex.Unlock()

	for {
		var msg map[string]any
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		event := msg["event"].(string)
		data := msg["data"]

		switch event {

		case "JOIN":
			name := data.(string)
			Mutex.Lock()
			if len(Teams) >= 4 {
				conn.WriteJSON(map[string]any{"event": "ROOM_FULL"})
				Mutex.Unlock()
				continue
			}
			Teams[id] = Team{ID: id, Name: name}
			Mutex.Unlock()
			Broadcast("TEAM_LIST", Teams)

		case "OPEN_BUZZ":
			ResetGame()
			State.BuzzOpen = true
			Broadcast("BUZZ_OPEN", nil)

		case "BUZZ":
			Mutex.Lock()
			if !State.BuzzOpen || State.WinnerID != "" {
				Mutex.Unlock()
				continue
			}
			State.WinnerID = id
			State.BuzzOpen = false
			winner := Teams[id]
			Mutex.Unlock()

			Broadcast("BUZZ_WINNER", map[string]any{
				"teamName": winner.Name,
				"time":     time.Now().UnixMilli(),
			})

		case "RESET":
			ResetGame()
			Broadcast("RESET", nil)
		}
	}

	Mutex.Lock()
	delete(Clients, id)
	delete(Teams, id)
	Mutex.Unlock()

	Broadcast("TEAM_LIST", Teams)
}

func StartServer() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", wsHandler)

	buildDir := "build"
	fs := http.FileServer(http.Dir(buildDir))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(buildDir, r.URL.Path)

		if _, err := os.Stat(path); err == nil {
			fs.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(buildDir, "index.html"))
	}))

	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	go func() {
		log.Println("Server running at :3000")
		server.ListenAndServe()
	}()

	return server
}
