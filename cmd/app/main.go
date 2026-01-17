package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"context"
	"net/http"

	"github.com/arizard/gomments"
	"github.com/arizard/gomments/internal"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/secure"
	"github.com/gin-gonic/gin"
)

func mustGetEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env variable: %s", k)
	}
	return v
}

func main() {
	settings := struct {
		port        string
		baseURL     string
		allowOrigin string
		cors        cors.Config
	}{
		port:        mustGetEnv("PORT"),
		baseURL:     os.Getenv("BASE_URL"),
		allowOrigin: os.Getenv("ALLOW_ORIGIN"),
		cors:        cors.DefaultConfig(),
	}

	if settings.allowOrigin != "" {
		settings.cors.AllowOrigins = []string{settings.allowOrigin, "https://less.coffee"}
	} else {
		settings.cors.AllowOrigins = []string{"https://less.coffee"}
	}
	settings.cors.AllowMethods = []string{"GET", "POST", "OPTIONS"}

	log.Printf("base url is %q", settings.baseURL)

	router := gin.Default()
	router.MaxMultipartMemory = 1 << 20 // 1 MB
	router.SetTrustedProxies(nil)

	router.Use(secure.Secure(secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
	}))
	router.Use(cors.New(settings.cors))
	router.Use(internal.NewClientIPRateLimiterMiddleware(10))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dbx, err := internal.InitSQLiteDatabase("/home/appuser/data/gomments.db")
	if err != nil {
		log.Fatalf("getting migrated dbx: %s", err)
		return
	}
	svc := gomments.New(ctx, dbx)

	rg := router.Group(settings.baseURL)
	rg.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	rg.GET("/articles/:article/replies", func(c *gin.Context) {
		resp, err := svc.GetReplies(ctx, gomments.GetRepliesRequest{Article: c.Param("article")})
		if err != nil {
			var gsErr *gomments.ServiceError
			if errors.As(err, &gsErr) {
				c.AbortWithError(gsErr.Status(), err)
			} else {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	rg.POST("/articles/:article/replies", func(c *gin.Context) {
		var req gomments.SubmitReplyRequest
		c.BindJSON(&req)

		req.Article = c.Param("article")
		if len(req.Article) > 1024 {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("article id too long"))
		}
		resp, err := svc.SubmitReply(ctx, req)
		if err != nil {
			var gsErr *gomments.ServiceError
			if errors.As(err, &gsErr) {
				c.AbortWithError(gsErr.Status(), err)
			} else {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	rg.GET("/articles/replies/stats", func(c *gin.Context) {
		req := gomments.GetStatsByArticlesRequest{
			Articles: c.QueryArray("article"),
		}
		resp, err := svc.GetStatsByArticles(ctx, req)
		if err != nil {
			var gsErr *gomments.ServiceError
			if errors.As(err, &gsErr) {
				c.AbortWithError(gsErr.Status(), err)
			} else {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	rg.POST("/sessions", func(c *gin.Context) {
		resp, err := svc.CreateSession(ctx)
		if err != nil {
			var gsErr *gomments.ServiceError
			if errors.As(err, &gsErr) {
				c.AbortWithError(gsErr.Status(), err)
			} else {
				c.AbortWithError(http.StatusInternalServerError, err)
			}
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	if err := router.Run(fmt.Sprintf(":%s", settings.port)); err != nil {
		log.Fatalln(err.Error())
		return
	}
}
