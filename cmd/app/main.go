package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	gomments "github.com/arizard/gomment"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type gommentsEnv struct {
	required map[string]string
	optional map[string]string
}

func main() {
	env := gommentsEnv{
		required: map[string]string{
			"PORT": "",
		},
		optional: map[string]string{
			"BASE_URL": "",
		},
	}

	for k := range env.required {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("missing env variable: %s", k)
		}

		env.required[k] = v
	}

	for k := range env.optional {
		env.optional[k] = os.Getenv(k)
	}

	baseURL := env.optional["BASE_URL"]
	log.Printf("base url is %q", baseURL)

	_, err := os.Stat("/root/data/gomments.db")
	if os.IsNotExist(err) {
		log.Println("db does not exist, creating new")
	}

	db, err := sql.Open("sqlite3", "/root/data/gomments.db")
	if err != nil {
		log.Fatalf("opening db: %s", err)
		return
	}

	driver, err := sqlite3.WithInstance(
		db,
		&sqlite3.Config{
			DatabaseName: "gomments",
		},
	)
	if err != nil {
		log.Fatalf("creating driver: %s", err)
		return
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		log.Fatalf("creating migrations: %s", err)
		return
	}

	if err := m.Up(); err != nil {
		if err.Error() != "no change" {
			log.Fatalf("migrating up: %s", err)
		}
	}

	dbx := sqlx.NewDb(db, "sqlite3")

	svc := gomments.New(dbx)

	router := gin.Default()
	router.SetTrustedProxies(nil)
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = []string{"http://localhost:1313", "https://less.coffee"}
	corsCfg.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	router.Use(cors.New(corsCfg))

	commentsRouter := router.Group(baseURL)

	commentsRouter.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	commentsRouter.GET("/articles/:article/replies", func(c *gin.Context) {
		ctx := context.Background()
		resp, err := svc.GetReplies(ctx, gomments.GetRepliesRequest{Article: c.Param("article")})
		if err != nil {
			c.AbortWithError(err.Status(), err)
			return
		}

		c.JSON(
			http.StatusOK,
			struct {
				Replies gomments.Replies `json:"replies"`
			}{
				Replies: resp.Replies,
			},
		)
	})

	commentsRouter.POST("/articles/:article/replies", func(c *gin.Context) {
		var req gomments.SubmitReplyRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		req.ReplyArticle = c.Param("article")

		ctx := context.Background()
		resp, err := svc.SubmitReply(
			ctx,
			req,
		)
		if err != nil {
			c.AbortWithError(err.Status(), err)
			return
		}

		c.JSON(
			http.StatusOK,
			struct {
				Reply gomments.Reply `json:"reply"`
			}{
				Reply: resp.Reply,
			},
		)
	})

	if err := router.Run(fmt.Sprintf(":%s", env.required["PORT"])); err != nil {
		log.Fatalln(err.Error())
		return
	}
}
