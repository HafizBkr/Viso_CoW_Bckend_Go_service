package models

import "github.com/gorilla/websocket"

type Participant struct {
	Conn          *websocket.Conn
	UserID        string
	Username      string
	Role          string
	AudioMuted    bool
	VideoOff      bool
	ScreenSharing bool
}
