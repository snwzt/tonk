package server

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"
	"log"
	"net/http"
	"os"
	"strings"
	"tonk/models"
	"tonk/store"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type handlers interface {
	health(echo.Context) error

	home(echo.Context) error
	join(echo.Context) error

	createGame(echo.Context) error
	joinGame(echo.Context) error
	wsConn(echo.Context) error
	leaderboard(echo.Context) error
	left(echo.Context) error

	handleWebSocketMessage(*websocket.Conn, *webrtc.PeerConnection, []byte)
	webrtcInitialize(*websocket.Conn) (*webrtc.PeerConnection, *webrtc.DataChannel, error)
}

type Handle struct {
	Logger *zerolog.Logger
	Store  store.Store
}

func (h *Handle) health(c echo.Context) error {
	return c.String(http.StatusOK, "healthy")
}

func (h *Handle) home(c echo.Context) error {
	return c.Render(http.StatusOK, "home.html", nil)
}

func (h *Handle) join(c echo.Context) error {
	gameID := c.Param("id")
	return c.Render(http.StatusOK, "join.html", map[string]string{"gameID": gameID})
}

func (h *Handle) createGame(c echo.Context) error {
	var createGameRequest models.CreateGameRequest
	if err := c.Bind(&createGameRequest); err != nil {
		return c.String(http.StatusBadRequest, "invalid request body")
	}

	userID := createGameRequest.PlayerName + "-" + strings.ReplaceAll(uuid.New().String(), "-", "")
	gameID := "game-" + strings.ReplaceAll(uuid.New().String(), "-", "")

	gameUrl := "http://" + c.Request().Host + "/game/" + gameID
	wsUrl := "ws://" + c.Request().Host + "/ws"

	if os.Getenv("SECURE_FLAG") == "1" {
		gameUrl = "https://" + c.Request().Host + "/game/" + gameID
		wsUrl = "wss://" + c.Request().Host + "/ws"
	}

	return c.Render(http.StatusOK, "game", map[string]string{
		"username":      createGameRequest.PlayerName,
		"userID":        userID,
		"gameUrl":       gameUrl, // copyable link will be shown in leaderboard
		"userTankColor": createGameRequest.TankBaseColor + "-" + createGameRequest.TankTopColor,
		"wsUrl":         wsUrl,
	})
}

func (h *Handle) joinGame(c echo.Context) error {
	gameID := c.Param("id")

	var joinGameRequest models.JoinGameRequest
	if err := c.Bind(&joinGameRequest); err != nil {
		return c.String(http.StatusBadRequest, "invalid request body")
	}

	userID := joinGameRequest.PlayerName + "-" + strings.ReplaceAll(uuid.New().String(), "-", "")

	gameUrl := "http://" + c.Request().Host + "/game/" + gameID
	wsUrl := "ws://" + c.Request().Host + "/ws"

	if os.Getenv("SECURE_FLAG") == "1" {
		gameUrl = "https://" + c.Request().Host + "/game/" + gameID
		wsUrl = "wss://" + c.Request().Host + "/ws"
	}

	return c.Render(http.StatusOK, "game", map[string]string{
		"username":      joinGameRequest.PlayerName,
		"userID":        userID,
		"gameUrl":       gameUrl, // copyable link will be shown in leaderboard
		"userTankColor": joinGameRequest.TankBaseColor + "-" + joinGameRequest.TankTopColor,
		"wsUrl":         wsUrl,
	})
}

func (h *Handle) wsConn(c echo.Context) error {
	gameID := c.Param("gameID")
	userID := c.Param("userID")

	ws, err := upgrade.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	defer ws.Close()

	// webrtc init
	pc, datChan, err := h.webrtcInitialize(ws)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	defer pc.Close()

	// webrtc
	datChan.OnMessage(func(msg webrtc.DataChannelMessage) {
		var recState models.State
		if err := json.Unmarshal(msg.Data, &recState); err != nil {
			log.Println("unmarshal error: " + err.Error())
		}

		if err := h.Store.SetReceivedTanks(gameID, userID, recState.UserTanks[userID]); err != nil {
			h.Logger.Err(err).Msg("error while set received tanks")
		}

		if err := h.Store.SetReceivedShots(gameID, userID, recState.Shots); err != nil {
			h.Logger.Err(err).Msg("error while set received shots")
		}

		if err := h.Store.SetReceivedMines(gameID, userID, recState.Mines); err != nil {
			h.Logger.Err(err).Msg("error while set received mines")
		}

		// send state
		state, err := h.Store.GetState(gameID)
		if err != nil {
			h.Logger.Err(err).Msg("error while get state")
		}

		message, err := json.Marshal(state)
		if err != nil {
			h.Logger.Err(err).Msg("marshal error")
			return
		}

		if datChan != nil && datChan.ReadyState() == webrtc.DataChannelStateOpen {
			datChan.SendText(string(message))
		}
	})

	// websocket
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			h.Logger.Info().Msg("read error: " + err.Error())
			break
		}
		h.handleWebSocketMessage(ws, pc, msg)
	}

	return nil
}

func (h *Handle) leaderboard(c echo.Context) error {
	return nil
}

func (h *Handle) left(c echo.Context) error {
	return c.Render(http.StatusOK, "left", nil)
}

func (h *Handle) handleWebSocketMessage(ws *websocket.Conn, pc *webrtc.PeerConnection, msg []byte) {
	var wsMsg models.WsMessage
	if err := json.Unmarshal(msg, &wsMsg); err != nil {
		h.Logger.Err(err).Msg("Unmarshal error")
		return
	}

	if wsMsg.SDP != nil {
		if wsMsg.SDP.Type == webrtc.SDPTypeAnswer {
			if err := pc.SetRemoteDescription(*wsMsg.SDP); err != nil {
				h.Logger.Err(err).Msg("SetRemoteDescription error")
				return
			}
			h.Logger.Info().Msg("Answer received")
		} else if wsMsg.SDP.Type == webrtc.SDPTypeOffer {
			if err := pc.SetRemoteDescription(*wsMsg.SDP); err != nil {
				h.Logger.Err(err).Msg("SetRemoteDescription error")
				return
			}
			answer, err := pc.CreateAnswer(nil)
			if err != nil {
				h.Logger.Err(err).Msg("CreateAnswer error")
				return
			}
			if err := pc.SetLocalDescription(answer); err != nil {
				h.Logger.Err(err).Msg("SetLocalDescription error")
				return
			}
			h.Logger.Info().Msg("Answer generated")
			ws.WriteJSON(models.WsMessage{SDP: &answer})
		}
	} else if wsMsg.ICE != nil {
		if err := pc.AddICECandidate(*wsMsg.ICE); err != nil {
			h.Logger.Err(err).Msg("AddICECandidate error")
		} else {
			h.Logger.Info().Msg("ICE added")
		}
	} else {
		h.Logger.Info().Msg("Other msg received from websocket")
	}
}

func (h *Handle) webrtcInitialize(ws *websocket.Conn) (*webrtc.PeerConnection, *webrtc.DataChannel, error) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		h.Logger.Info().Msg("Peer Connection State has changed: " + s.String())
	})

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			j := candidate.ToJSON()
			ws.WriteJSON(models.WsMessage{ICE: &j})
		}
		h.Logger.Info().Msg("ICE Sent")
	})

	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		peerConnection.Close()
		return nil, nil, err
	}

	dataChannel.OnOpen(func() {
		h.Logger.Info().Msg("Data channel opened")
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		peerConnection.Close()
		return nil, nil, err
	}

	h.Logger.Info().Msg("Offer created")

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		peerConnection.Close()
		return nil, nil, err
	}

	response := models.WsMessage{SDP: &offer}
	if err := ws.WriteJSON(response); err != nil {
		h.Logger.Err(err).Msg("WebSocket write error")
	}

	h.Logger.Info().Msg("Offer sent")

	return peerConnection, dataChannel, nil
}
