package routes

import (
	"context"
	"errors"
	"fmt"
	db "ginserver/config"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// Claims defines the JWT claims structure.
type Claims struct {
	User_id string `json:"user_id"`
	jwt.StandardClaims
}

func GetSecretToken() []byte {

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

func CreateToken(user_id string) (string, error) {
	// Create JWT token
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires after 24 hours
	claims := &Claims{
		User_id: user_id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(GetSecretToken()))
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

func AuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
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
			token, err := CreateToken(req.ID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{
						"error":   "VALIDATEERR-5",
						"message": "Internal server error"})
				return
			}

			// Set cookie
			cookie := &http.Cookie{
				Name:     "todo_token",
				Value:    token,
				Expires:  time.Now().Add(24 * time.Hour),
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			}
			http.SetCookie(c.Writer, cookie)

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
				{Key: "deleted", Value: false},
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

			token, err := CreateToken(existingUser.ID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{
						"error":   "VALIDATEERR-5",
						"message": "Internal server error"})
				return
			}

			// Set cookie
			cookie := &http.Cookie{
				Name:     "todo_token",
				Value:    token,
				Expires:  time.Now().Add(24 * time.Hour),
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			}
			http.SetCookie(c.Writer, cookie)

			c.JSON(http.StatusOK, gin.H{
				"token": token,
			})
		})

	}
}
