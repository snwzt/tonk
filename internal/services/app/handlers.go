package app

import "github.com/labstack/echo/v4"

type handler interface {
	health(echo.Context) error
}
