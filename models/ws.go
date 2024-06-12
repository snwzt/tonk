package models

import "github.com/pion/webrtc/v4"

type WsMessage struct {
	Type string      `json:"type"`
	Msg  interface{} `json:"msg"`
}

type WebrtcExchange struct {
	SDP *webrtc.SessionDescription `json:"sdp"`
	ICE *webrtc.ICECandidateInit   `json:"ice"`
}

type Leaderboard map[string]Score

type Score struct {
	Kills int `json:"kills"`
	Death int `json:"death"`
}
