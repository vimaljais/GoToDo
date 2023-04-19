package routes

import (
	"context"
	"fmt"
	db "ginserver/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
}

type ResponseUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func UserRoutes(r *gin.Engine) {
	user := r.Group("/user")
	{
		user.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "userroute",
			})
		})

		user.GET("/list", func(c *gin.Context) {

			var users []bson.M

			collection := db.Client.Database("todo").Collection("users")
			cursor, err := collection.Find(context.TODO(), bson.D{})
			err = cursor.All(context.TODO(), &users)
			if err != nil {
				panic(err)
			}
			//filter deleted
			for i := 0; i < len(users); i++ {
				if users[i]["deleted"] == true {
					users = append(users[:i], users[i+1:]...)
					i--
				}
			}

			var resUsers []ResponseUser
			for _, user := range users {
				var singleUser ResponseUser
				singleUser.ID = user["id"].(string)
				singleUser.Username = user["username"].(string)
				resUsers = append(resUsers, singleUser)
			}

			c.JSON(http.StatusOK, resUsers)
		})

		user.POST("/change-password", func(c *gin.Context) {
			// Get username and old and new passwords from request

			var req struct {
				Username    string `json:"username" binding:"required"`
				OldPassword string `json:"old_password" binding:"required"`
				NewPassword string `json:"new_password" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "BAD_REQUEST", "message": err.Error()})
				return
			}
			fmt.Print(req)

			// Find user in database
			collection := db.Client.Database("todo").Collection("users")
			var user User
			err := collection.FindOne(context.TODO(), bson.D{
				{Key: "username", Value: req.Username},
			}).Decode(&user)

			// Check if user exists
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "USER_NOT_FOUND",
					"message": "User not found",
				})
				return
			}

			// Check if old password is correct
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "INVALID_PASSWORD",
					"message": "Invalid old password",
				})
				return
			}

			// Hash new password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "INTERNAL_SERVER_ERROR",
					"message": "Failed to hash new password",
				})
				return
			}

			// Update user's password in database
			_, err = collection.UpdateOne(context.TODO(), bson.D{
				{Key: "username", Value: req.Username},
			}, bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "password", Value: string(hashedPassword)},
				}},
			})
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "INTERNAL_SERVER_ERROR",
					"message": "Failed to update password",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Password updated successfully",
			})
		})

		user.DELETE("/delete/:id", func(c *gin.Context) {

			collection := db.Client.Database("todo").Collection("users")
			res, err := collection.UpdateOne(context.TODO(), bson.D{}, bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "deleted", Value: true},
					{Key: "updated_at", Value: time.Now()},
				}},
			})
			if err != nil {
				panic(err)
			}
			c.JSON(http.StatusOK, res)
		})

	}
}
