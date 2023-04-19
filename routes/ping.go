package routes

import "github.com/gin-gonic/gin"

func PingRoutes(r *gin.Engine) {
	ping := r.Group("/ping")
	{
		ping.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		ping.GET("/repeated", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"ping": "pong",
			})
		})
	}
}
