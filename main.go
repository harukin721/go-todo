package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type Todo struct {
	// bun は、Go の構造体を SQL クエリに変換するためのライブラリ
	bun.BaseModel `bun:"table:todos,alias:t"`

	ID        int       `bun:"type:serial,pk,autoincr"`
	Name      string    `bun:"type:text,notnull"`
	Body      string    `bun:"type:text,notnull"`
	Done      bool      `bun:"type:boolean,notnull,default:false"`
	CreatedAt time.Time `bun:"type:timestamp,default:current_timestamp"`
	UpdatedAt time.Time `bun:"type:timestamp,nullzero"`
	DeletedAt time.Time `bun:"type:timestamp,soft_delete,nullzero"`
}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
