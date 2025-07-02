package internal

import (
	"context"
	"net/http"

	"github.com/arizard/gomments"
	"github.com/gin-gonic/gin"
)

type InitRoutesOptions struct {
	BaseURL string
}

func InitRoutes(router *gin.Engine, svc *gomments.Service, opt InitRoutesOptions) error {
	baseURL := opt.BaseURL

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
			resp,
		)
	})

	commentsRouter.POST("/articles/:article/replies", func(c *gin.Context) {
		var req gomments.SubmitReplyRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		req.Article = c.Param("article")

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
			resp,
		)
	})

	// usage: /articles/replies/stats?article=abc&article=xyz
	commentsRouter.GET("/articles/replies/stats", func(c *gin.Context) {
		req := gomments.GetStatsByArticlesRequest{}
		req.Articles = c.QueryArray("article")

		ctx := context.Background()
		resp, err := svc.GetStatsByArticles(ctx, req)
		if err != nil {
			c.AbortWithError(err.Status(), err)
			return
		}

		c.JSON(
			http.StatusOK,
			resp,
		)
	})
	return nil
}
