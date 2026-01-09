package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"lcc-go/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TeamsToArray() []map[string]string {
	arr := []map[string]string{}
	for _, t := range Teams {
		arr = append(arr, map[string]string{
			"teamName": t.Name,
		})
	}
	return arr
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

	conn.WriteJSON(map[string]any{
		"event": "STATE_SYNC",
		"data": map[string]any{
			"buzzOpen": State.BuzzOpen,
			"winner":   nil,
			"teams":    TeamsToArray(),
		},
	})

	conn.WriteJSON(map[string]any{
		"event": "SERVER_INFO",
		"data": map[string]any{
			"ip":   utils.GetLocalIP(),
			"port": 3000,
		},
	})
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
			Teams[id] = Team{ID: id, Name: name}
			Mutex.Unlock()

			conn.WriteJSON(map[string]any{
				"event": "JOIN_OK",
			})

			Broadcast("STATE_SYNC", map[string]any{
				"buzzOpen": State.BuzzOpen,
				"winner":   nil,
				"teams":    TeamsToArray(),
			})

		case "OPEN_BUZZ":
			ResetGame()
			State.BuzzOpen = true
			Broadcast("BUZZ_OPEN", nil)

		case "GET_STATE":
			Mutex.Lock()
			conn.WriteJSON(map[string]any{
				"event": "STATE_SYNC",
				"data": map[string]any{
					"buzzOpen": State.BuzzOpen,
					"winner":   nil,
					"teams":    TeamsToArray(),
				},
			})
			Mutex.Unlock()

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

	Broadcast("STATE_SYNC", map[string]any{
		"buzzOpen": State.BuzzOpen,
		"winner":   nil,
		"teams":    TeamsToArray(),
	})
}

func StartServer() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", wsHandler)

	buildDir := "frontend"
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

	ip := utils.GetLocalIP()

	go func() {
		log.Println("Server running at http://" + ip + ":3000")
		server.ListenAndServe()
	}()

	return server
}
