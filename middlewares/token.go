package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func getSecretKey() []byte {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		// If the secret key is not set, panic with an error message
		panic("JWT secret key not found in environment variables")
	}
	return []byte(secretKey)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenCookie, err := c.Cookie("todo_token")
		fmt.Print(tokenCookie)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "AUTH-1",
				"message": "Authorization token is missing",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenCookie, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid token signing method")
			}
			return []byte(getSecretKey()), nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "AUTH-2",
				"message": "Invalid authorization token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "AUTH-3",
				"message": "Invalid authorization token",
			})
			c.Abort()
			return
		}

		if user_id, ok := claims["user_id"].(string); ok {
			c.Set("user_id", user_id)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "AUTH-3",
				"message": "Invalid authorization token",
			})
			c.Abort()
			return
		}

	}
}
