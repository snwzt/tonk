package server

import "github.com/labstack/echo/v4"

type Server struct {
	port     string
	engine   *echo.Echo
	handlers handlers
}

func NewServer(port string, engine *echo.Echo, handlers handlers) *Server {
	return &Server{
		port:     port,
		engine:   engine,
		handlers: handlers,
	}
}
