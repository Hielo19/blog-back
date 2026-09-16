package router

import (
	"net/http"

	"gin-demo/internal/handler"

	"github.com/gin-gonic/gin"
)

func New(
	postHandler *handler.PostHandler,
	userHandler *handler.UserHandler,
	categoryHandler *handler.CategoryHandler,
) *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandler.Create)
		}

		categories := v1.Group("/categories")
		{
			categories.POST("", categoryHandler.Create)
			categories.GET("", categoryHandler.List)
			categories.GET("/:id", categoryHandler.GetByID)
			categories.PATCH("/:id", categoryHandler.Update)
			categories.DELETE("/:id", categoryHandler.Delete)
		}

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
