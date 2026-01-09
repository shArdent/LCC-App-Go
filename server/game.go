package server

import "sync"

type GameState struct {
	BuzzOpen bool
	WinnerID string
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"teamName"`
}

var (
	State   = GameState{}
	Teams   = map[string]Team{}
	Clients = map[string]Client{}
	Mutex   sync.Mutex
)

func ResetGame() {
	State.BuzzOpen = false
	State.WinnerID = ""
}
