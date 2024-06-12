package httpserver

import "github.com/labstack/echo/v4"

type server struct {
	port     string
	engine   *echo.Echo
	handlers handlers
}
