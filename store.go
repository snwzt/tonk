package main

import (
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"net/http"
)

// websocket

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// messages

type WsMsg struct {
	SDP *webrtc.SessionDescription `json:"sdp"`
	ICE *webrtc.ICECandidateInit   `json:"ice"`
}

type DataChanMsg State

// state

// Tank represents the Tank object in JSON
type Tank struct {
	XPosition     float64 `json:"_xPosition"`
	YPosition     float64 `json:"_yPosition"`
	RotationAngle int     `json:"_rotationAngle"`
	Dir           string  `json:"_dir"`
	ShieldMode    bool    `json:"_shieldMode"`
	MineMode      bool    `json:"_mineMode"`
	RocketMode    bool    `json:"_rocketMode"`
	BurstMode     bool    `json:"_burstMode"`
}

// Shot represents the Shot object in JSON
type Shot struct {
	Name        string  `json:"name"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Vx          float64 `json:"vx"`
	Vy          float64 `json:"vy"`
	Angle       int     `json:"angle"`
	CurrentFire string  `json:"currentFire"`
}

// Mine represents the Mine object in JSON
type Mine struct {
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

// State represents the overall state in JSON
type State struct {
	UserTanks map[string]Tank `json:"userTanks"`
	Shots     []Shot          `json:"shots"`
	Mines     []Mine          `json:"mines"`
}

// store

var (
	state = State{
		UserTanks: make(map[string]Tank),
		Shots:     make([]Shot, 0),
		Mines:     make([]Mine, 0),
	}
	//gameMap = make(make([]int, 300), 200)
)
