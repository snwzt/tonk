package models

type CreateGameRequest struct {
	PlayerName    string `form:"playerName"`
	TankTopColor  string `form:"tankTopColor"`
	TankBaseColor string `form:"tankBaseColor"`
}

type JoinGameRequest struct {
	PlayerName    string `form:"playerName"`
	TankTopColor  string `form:"tankTopColor"`
	TankBaseColor string `form:"tankBaseColor"`
}
