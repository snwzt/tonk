package httpserver

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"tonk/models"
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
	left(echo.Context) error

	webrtcInitialize(*websocket.Conn) (*webrtc.PeerConnection, *webrtc.DataChannel, error)
}

type handle struct {
	Logger *zerolog.Logger

	CreateSession  chan string
	IntroToSession chan string
	//MsgFromSession chan string
	//MsgToSession   chan string
}

func (h *handle) health(c echo.Context) error {
	return c.String(http.StatusOK, "healthy")
}

func (h *handle) home(c echo.Context) error {
	return c.Render(http.StatusOK, "home.html", nil)
}

func (h *handle) join(c echo.Context) error {
	gameID := c.Param("id")
	return c.Render(http.StatusOK, "join.html", map[string]string{"gameID": gameID})
}

func (h *handle) createGame(c echo.Context) error {
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

	// signal manager channel to start a game session goroutine with name of current game

	return c.Render(http.StatusOK, "game", map[string]string{
		"username":      createGameRequest.PlayerName,
		"userID":        userID,
		"gameUrl":       gameUrl, // copyable link will be shown in leaderboard
		"userTankColor": createGameRequest.TankBaseColor + "-" + createGameRequest.TankTopColor,
		"wsUrl":         wsUrl,
	})
}

func (h *handle) joinGame(c echo.Context) error {
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

func (h *handle) wsConn(c echo.Context) error {
	//userID := c.Param("userID")

	ws, err := upgrade.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	defer ws.Close()

	// here somehow tell the game session goroutine to read from userID channel?
	// I am an idiot

	// webrtc init
	pc, datChan, err := h.webrtcInitialize(ws)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	defer pc.Close()

	var wg sync.WaitGroup

	wg.Add(3)

	// websocket reader
	go func() {
		defer wg.Done()
	}()

	// websocket writer
	go func() {
		defer wg.Done()
	}()

	// webrtc reader
	datChan.OnMessage(func(msg webrtc.DataChannelMessage) {

	})

	// webrtc writer
	go func() {
		defer wg.Done()
	}()

	wg.Wait()

	return nil
}

func (h *handle) left(c echo.Context) error {
	return c.Render(http.StatusOK, "left", nil)
}

func (h *handle) webrtcInitialize(ws *websocket.Conn) (*webrtc.PeerConnection, *webrtc.DataChannel, error) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	defer peerConnection.Close()
	if err != nil {
		return nil, nil, err
	}

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		h.Logger.Info().Msg("Peer Connection State has changed: " + s.String())
	})

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			j := candidate.ToJSON()
			ws.WriteJSON(models.WsMessage{
				Type: "WebrtcExchange",
				Msg:  models.WebrtcExchange{ICE: &j},
			})
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

	response := models.WsMessage{
		Type: "WebrtcExchange",
		Msg:  models.WebrtcExchange{SDP: &offer},
	}
	if err := ws.WriteJSON(response); err != nil {
		log.Println("WebSocket write error:", err)
	}

	h.Logger.Info().Msg("Offer sent")

	return peerConnection, dataChannel, nil
}
