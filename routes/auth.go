package routes

import (
	"context"
	"errors"
	"fmt"
	db "ginserver/config"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// Claims defines the JWT claims structure.
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func CreateToken(username string) (string, error) {
	// Create JWT token
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires after 24 hours
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("SECRET_KEY"))
}

func VerifyToken(tokenString string) (*Claims, error) {
	// Verify JWT token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("SECRET_KEY"), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "MISSING_AUTH_TOKEN", "message": "Authorization token is missing."})
			return
		}

		// Verify token
		claims, err := VerifyToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "INVALID_AUTH_TOKEN", "message": "Authorization token is invalid."})
			return
		}

		// Set user in context
		c.Set("username", claims.Username)
		c.Next()
	}
}

func AuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	auth.Use(AuthMiddleware())
	{
		auth.POST("/register", func(c *gin.Context) {
			var req User
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "VALIDATEERR-1",
						"message": "Invalid inputs. Please check your inputs"})
				return
			}

			// Check if user already exists
			collection := db.Client.Database("todo").Collection("users")
			var existingUser bson.M
			err := collection.FindOne(context.TODO(), bson.D{
				{Key: "username", Value: req.Username},
			}).Decode(&existingUser)

			if err == nil {
				c.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "VALIDATEERR-2",
						"message": "User already exists"})
				return
			}

			// Hash password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{
						"error":   "VALIDATEERR-3",
						"message": "Internal server error"})
				return
			}
			req.ID = uuid.New().String()
			req.Password = string(hashedPassword)
			req.Type = "user"
			req.CreatedAt = time.Now()
			req.UpdatedAt = time.Now()
			req.Deleted = false

			// Create user
			result, err := collection.InsertOne(context.TODO(), req)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{
						"error":   "VALIDATEERR-4",
						"message": "Internal server error"})
				return
			}
			fmt.Println(result.InsertedID)

			// Create JWT token
			token, err := CreateToken(req.Username)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{
						"error":   "VALIDATEERR-5",
						"message": "Internal server error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"token": token,
			})
		})
		auth.POST("/login", func(c *gin.Context) {
			var req User
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "VALIDATEERR-1",
					"message": "Invalid inputs. Please check your inputs",
				})
				return
			}

			// Find user by username
			collection := db.Client.Database("todo").Collection("users")
			var existingUser User
			err := collection.FindOne(context.TODO(), bson.D{
				{Key: "username", Value: req.Username},
			}).Decode(&existingUser)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":   "INVALID_CREDENTIALS",
					"message": "Invalid username or password",
				})
				return
			}

			// Check password
			if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error":   "INVALID_CREDENTIALS",
					"message": "Invalid username or password",
				})
				return
			}

			// Create JWT token
			token, err := CreateToken(existingUser.Username)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "VALIDATEERR-5",
					"message": "Internal server error",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"token": token,
			})
		})

	}
}
