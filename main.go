package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
)

func main() {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet},
	}))

	e.GET("/ws", func(c echo.Context) error {
		ws, err := upgrade.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		defer ws.Close()

		pc, _, err := NewPr(ws)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		defer pc.Close()

		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			handleWebSocketMessage(ws, pc, msg)
		}

		return nil
	})

	e.Logger.Fatal(e.Start(":8080"))
}
