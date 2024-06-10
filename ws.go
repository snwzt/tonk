package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"log"
)

func handleWebSocketMessage(ws *websocket.Conn, pc *webrtc.PeerConnection, msg []byte) {
	var wsMsg WsMsg
	if err := json.Unmarshal(msg, &wsMsg); err != nil {
		log.Println("Unmarshal error:", err)
		return
	}

	if wsMsg.SDP != nil {
		if wsMsg.SDP.Type == webrtc.SDPTypeAnswer {
			if err := pc.SetRemoteDescription(*wsMsg.SDP); err != nil {
				log.Println("SetRemoteDescription error:", err)
				return
			}
			log.Println("answer received")
		} else if wsMsg.SDP.Type == webrtc.SDPTypeOffer {
			if err := pc.SetRemoteDescription(*wsMsg.SDP); err != nil {
				log.Println("SetRemoteDescription error:", err)
				return
			}
			answer, err := pc.CreateAnswer(nil)
			if err != nil {
				log.Println("CreateAnswer error:", err)
				return
			}
			if err := pc.SetLocalDescription(answer); err != nil {
				log.Println("SetLocalDescription error:", err)
				return
			}
			log.Println("answer generated")
			ws.WriteJSON(WsMsg{SDP: &answer})
		}
	} else if wsMsg.ICE != nil {
		if err := pc.AddICECandidate(*wsMsg.ICE); err != nil {
			log.Println("AddICECandidate error:", err)
		} else {
			log.Println("ice added")
		}
	} else {
		log.Println("Other msg received from websocket")
	}
}
