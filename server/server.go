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
	Mutex.Unlock()

	ip := utils.GetLocalIP()

	Broadcast("STATE_SYNC",
		map[string]any{
			"buzzOpen":    State.BuzzOpen,
			"winner":      nil,
			"teams":       TeamsToArray(),
			"sessionName": State.Session,
		})

	Broadcast("SERVER_INFO",
		map[string]any{
			"port": 3000,
			"ip":   ip,
		})
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
				"buzzOpen":    State.BuzzOpen,
				"winner":      nil,
				"teams":       TeamsToArray(),
				"sessionName": State.Session,
			})
		case "CHANGE_SESSION":
			sessionName := data.(string)
			Mutex.Lock()
			State.Session = sessionName
			Mutex.Unlock()
			Broadcast("STATE_SYNC", map[string]any{
				"buzzOpen":    State.BuzzOpen,
				"winner":      nil,
				"teams":       TeamsToArray(),
				"sessionName": State.Session,
			})
		case "OPEN_BUZZ":
			ResetGame()
			Mutex.Lock()
			State.BuzzOpen = true
			targetTime := time.Now().Add(10 * time.Second).UnixMilli()
			Mutex.Unlock()

			Broadcast("BUZZ_OPEN", nil)
			Broadcast("START_COUNTDOWN", map[string]any{
				"targetTime": targetTime,
			})

		case "GET_STATE":
			Broadcast("STATE_SYNC",
				map[string]any{
					"buzzOpen":    State.BuzzOpen,
					"winner":      nil,
					"teams":       TeamsToArray(),
					"sessionName": State.Session,
				})
		case "START_COUNTDOWN":
			Mutex.Lock()
			targetTime := time.Now().Add(10 * time.Second).UnixMilli()
			Mutex.Unlock()

			Broadcast("START_COUNTDOWN", map[string]any{
				"targetTime": targetTime,
			})
		case "RESET_COUNTDOWN":
			Broadcast("RESET_COUNTDOWN", nil)
		case "BUZZ":
			Mutex.Lock()
			if !State.BuzzOpen || State.WinnerID != "" {
				Mutex.Unlock()
				continue
			}

			State.WinnerID = id
			State.BuzzOpen = false
			winner := Teams[id]
			targetTime := time.Now().Add(10 * time.Second).UnixMilli()
			Mutex.Unlock()

			Broadcast("BUZZ_WINNER", map[string]any{
				"teamName": winner.Name,
				"time":     time.Now().UnixMilli(),
			})

			Broadcast("START_COUNTDOWN", map[string]any{
				"targetTime": targetTime,
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
		"buzzOpen":    State.BuzzOpen,
		"winner":      nil,
		"teams":       TeamsToArray(),
		"sessionName": State.Session,
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
