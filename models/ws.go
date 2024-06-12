package models

import "github.com/pion/webrtc/v4"

type WsMessage struct {
	SDP *webrtc.SessionDescription `json:"sdp"`
	ICE *webrtc.ICECandidateInit   `json:"ice"`
}
