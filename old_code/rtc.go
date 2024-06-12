package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"log"
)

func NewPr(ws *websocket.Conn) (*webrtc.PeerConnection, *webrtc.DataChannel, error) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("Peer Connection State has changed: %s\n", s.String())
	})

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			j := candidate.ToJSON()
			ws.WriteJSON(WsMsg{ICE: &j})
		}
		log.Println("ice sent")
	})

	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		peerConnection.Close()
		return nil, nil, err
	}

	dataChannel.OnOpen(func() {
		log.Println("Data channel opened")
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		var dataChanMsg DataChanMsg
		if err := json.Unmarshal(msg.Data, &dataChanMsg); err != nil {
			log.Println("unmarshal error: " + err.Error())
		}

		state.UserTanks = dataChanMsg.UserTanks
		state.Shots = dataChanMsg.Shots
		state.Mines = dataChanMsg.Mines

		// move this to ticker maybe

		message, err := json.Marshal(state)
		if err != nil {
			log.Println("Marshal error:", err)
			return
		}

		if dataChannel != nil && dataChannel.ReadyState() == webrtc.DataChannelStateOpen {
			dataChannel.SendText(string(message))
		}
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		peerConnection.Close()
		return nil, nil, err
	}

	log.Println("offer created")

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		peerConnection.Close()
		return nil, nil, err
	}

	response := WsMsg{SDP: &offer}
	if err := ws.WriteJSON(response); err != nil {
		log.Println("WebSocket write error:", err)
	}

	log.Println("offer sent")

	return peerConnection, dataChannel, nil
}
