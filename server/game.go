package server

import "sync"

type GameState struct {
	BuzzOpen        bool
	WinnerID        string
	Session         string
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"teamName"`
}

var (
	State   = GameState{Session: "WAJIB"}
	Teams   = map[string]Team{}
	Clients = map[string]Client{}
	Mutex   sync.Mutex
)

func ResetGame() {
	State.BuzzOpen = false
	State.WinnerID = ""
}
