package router

import (
	"net/http"

	"gin-demo/internal/handler"

	"github.com/gin-gonic/gin"
)

func New(postHandler *handler.PostHandler) *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	v1 := router.Group("/api/v1")
	{
		posts := v1.Group("/posts")
		{
			posts.POST("", postHandler.Create)
			posts.GET("", postHandler.List)
			posts.GET("/:id", postHandler.GetByID)
			posts.PATCH("/:id", postHandler.Update)
			posts.DELETE("/:id", postHandler.Delete)
		}
	}

	return router
}
