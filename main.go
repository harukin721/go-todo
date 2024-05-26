package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"
)

// go:embed static
var static embed.FS

// go:embed templates/*
var templates embed.FS

// +------------+-----------------+------+-----+-------------------+-------------------+
// | Field      | Type            | Null | Key | Default           | Extra             |
// +------------+-----------------+------+-----+-------------------+-------------------+
// | id         | bigint unsigned | NO   | PRI | NULL              | auto_increment    |
// | name       | text            | NO   |     | NULL              |                   |
// | body       | text            | NO   |     | NULL              |                   |
// | done       | tinyint(1)      | NO   |     | 0                 |                   |
// | created_at | timestamp       | YES  |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
// | updated_at | timestamp       | YES  |     | NULL              |                   |
// | deleted_at | timestamp       | YES  |     | NULL              |                   |
// +------------+-----------------+------+-----+-------------------+-------------------+
// 7 rows in set (0.00 sec)

type Todo struct {
	// bun は、Go の構造体を SQL クエリに変換するためのライブラリ
	bun.BaseModel `bun:"table:todos,alias:t"`

	ID        int       `bun:"type:serial,pk"`
	Name      string    `bun:"type:text,notnull"`
	Body      string    `bun:"type:text,notnull"`
	Done      bool      `bun:"type:boolean,notnull,default:false"`
	CreatedAt time.Time `bun:"type:timestamp,default:current_timestamp"`
	UpdatedAt time.Time `bun:"type:timestamp,nullzero"`
	DeletedAt time.Time `bun:"type:timestamp,soft_delete,nullzero"`
}

type Data struct {
	Todos  []Todo
	Errors []error
}

// テンプレートエンジンの設定
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func formatDateTime(t time.Time) string {
	// 時刻がゼロ値の場合は空文字を返す
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
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
	_, err = db.NewCreateTable().IfNotExists().Model((*Todo)(nil)).Exec(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// サーバー起動
	e := echo.New()

	// テンプレートエンジンの設定
	e.Renderer = &Template{
		// テンプレートファイルを読み込む
		templates: template.Must(template.New("").
			Funcs(template.FuncMap{"formatDateTime": formatDateTime}).
			ParseFS(templates, "templates/*html")),
	}

	e.GET("/", func(c echo.Context) error {
		var todos []Todo
		// テーブルからデータを取得
		ctx := context.Background()
		// SELECT * FROM todos
		err := db.NewSelect().Model(&todos).Scan(ctx)
		if err != nil {
			e.Logger.Error(err)
			return c.Render(http.StatusBadRequest, "index", Data{
				Errors: []error{errors.New("データの取得に失敗しました")},
			})
		}
		return c.Render(http.StatusOK, "index", Data{Todos: todos})
	})
	e.POST("/", func(c echo.Context) error {
		var todo Todo
		// フォームからデータを取得
		errs := echo.FormFieldBinder(c).
			Int("id", &todo.ID).
			String("name", &todo.Name).
			String("body", &todo.Body).
			Bool("done", &todo.Done).
			BindErrors()
		if errs != nil {
			e.Logger.Error(err)
			return c.Render(http.StatusBadRequest, "index", Data{Errors: errs})
		} else if todo.ID == 0 {
			// テーブルにデータを追加
			ctx := context.Background()
			if todo.Body == "" {
				err = errors.New("内容を入力してください")
			} else {
				// INSERT INTO todos (name, body) VALUES (?, ?)
				_, err = db.NewInsert().Model(&todo).Exec(ctx)
				if err != nil {
					e.Logger.Error(err)
					err = errors.New("データの追加に失敗しました")
				}
			}
		} else {
			ctx := context.Background()
			if c.FormValue("delete") != "" {
				_, err = db.NewDelete().Model(&todo).Where("id = ?", todo.ID).Exec(ctx)
			} else {
				var orig Todo
				err = db.NewSelect().Model(&orig).Where("id = ?", todo.ID).Scan(ctx)
				if err == nil {
					orig.Done = todo.Done
					_, err = db.NewUpdate().Model(&orig).Where("id = ?", todo.ID).Exec(ctx)
				}
			}
			if err != nil {
				e.Logger.Error(err)
				err = errors.New("失敗")
			}
		}
		if err != nil {
			return c.Render(http.StatusBadRequest, "index", Data{Errors: []error{err}})
		}
		return c.Redirect(http.StatusFound, "/")
	})
	staticFs, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}
	fileServer := http.FileServer(http.FileSystem(http.FS(staticFs)))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", fileServer)))
	e.Logger.Fatal(e.Start(":1323"))
}
