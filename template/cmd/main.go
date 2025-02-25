package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from {{ project_name }} tokenissuer!")
	})

	log.Println("{{ project_name }} tokenissuer listening on :7001...")
	e.Logger.Fatal(e.Start(":7001"))
}
