package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"
)

// +-------+-----------------+------+-----+---------+----------------+
// | Field | Type            | Null | Key | Default | Extra          |
// +-------+-----------------+------+-----+---------+----------------+
// | id    | bigint unsigned | NO   | PRI | NULL    | auto_increment |
// | name  | text            | NO   |     | NULL    |                |
// | body  | text            | NO   |     | NULL    |                |
// | done  | tinyint(1)      | NO   |     | 0       |                |
// +-------+-----------------+------+-----+---------+----------------+
// 4 rows in set (0.01 sec)

type Todo struct {
	// bun は、Go の構造体を SQL クエリに変換するためのライブラリ
	bun.BaseModel `bun:"table:todos,alias:t"`

	ID   int    `bun:"type:serial,pk"`
	Name string `bun:"type:text,notnull"`
	Body string `bun:"type:text,notnull"`
	Done bool   `bun:"type:boolean,notnull,default:false"`
}

func main() {

	// .env ファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// MySQL に接続
	sqldb, err := sql.Open("mysql", os.Getenv("DATABASE_URL")) // DATABASE_URL は .env ファイルに記述
	if err != nil {
		log.Fatal(err)
	}
	defer sqldb.Close()

	// bun に sqldb を渡して DB オブジェクトを作成
	db := bun.NewDB(sqldb, mysqldialect.New())
	defer db.Close()

	// bun にクエリフックを追加
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
	))

	// テーブル作成
	ctx := context.Background()
	_, err = db.NewCreateTable().Model((*Todo)(nil)).Exec(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// サーバー起動
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
