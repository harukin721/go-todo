package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Todo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Body      string `json:"body"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
