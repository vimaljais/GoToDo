package main

import (
	db "ginserver/config"
	"ginserver/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	db.Connect()

	r := gin.Default()
	routes.PingRoutes(r)
	routes.Todo(r)
	routes.UserRoutes(r)
	routes.AuthRoutes(r)

	r.Run(":8090") // listen and serve on 0.0.0.0:8090
}
