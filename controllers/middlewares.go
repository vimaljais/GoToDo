package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "AUTH-1",
				"message": "Authorization token is missing",
			})
			c.Abort()
			return
		}

		tokenString = strings.ReplaceAll(tokenString, "Bearer ", "")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid token signing method")
			}
			return []byte("your-secret-key"), nil // replace with your actual secret key
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "AUTH-2",
				"message": "Invalid authorization token",
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// extract access token and set it in the context
			accessToken := claims["access_token"].(string)
			c.Set("access_token", accessToken)
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
