package app

import (
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

type Server struct {
	port    string
	engine  *echo.Echo
	handler handler
	store   store
}

func (svc *Server) Run() {
	svc.engine.Use(middleware.Recover())
	svc.engine.Use(echoprometheus.NewMiddleware("app_service"))
	svc.engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
	}))

	svc.engine.GET("/health", svc.handler.health)
	svc.engine.GET("/metrics", echoprometheus.NewHandler())

	svc.engine.GET("/", nil)
	svc.engine.GET("/join", nil) // query params
	svc.engine.GET("/left", nil)

	apiGroup := svc.engine.Group("/api")
	apiGroup.POST("/createGame", nil) // returns hypermedia block of game
	apiGroup.POST("/joinGame", nil)   // returns hypermedia block of game
	apiGroup.GET("/game/:id", nil)
	apiGroup.GET("/connection/:id", nil)
	apiGroup.GET("/end", nil)
}
