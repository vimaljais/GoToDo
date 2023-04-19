// routes/notes.go
package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	db "ginserver/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID        primitive.ObjectID `json:"_id" json:"id"`
	Title     string             `json:"title"`
	Content   string             `json:"content"`
	UserID    string             `json:"user_id"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type ToUpdateNote struct {
	ID      string `json:"_id" json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func Todo(r *gin.Engine) {
	todo := r.Group("/todo")
	{
		todo.POST("/create", func(c *gin.Context) {
			var req Note
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "VALIDATEERR-1",
						"message": "Invalid inputs. Please check your inputs"})
				return
			}
			// err = client.Set("id", jsonData, 0).Err()
			fmt.Print(req.UserID)
			collection := db.Client.Database("todo").Collection("users")

			//find user by id
			id, _ := primitive.ObjectIDFromHex(req.UserID)

			var user bson.M
			err := collection.FindOne(context.TODO(), bson.D{
				{Key: "_id", Value: id},
			}).Decode(&user)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "VALIDATEERR-2",
					"message": "User not found"})

				return
			}

			collection = db.Client.Database("todo").Collection("notes")
			res, err := collection.InsertOne(context.TODO(), bson.D{
				{Key: "title", Value: req.Title},
				{Key: "content", Value: req.Content},
				{Key: "user_id", Value: req.UserID},
				{Key: "created_at", Value: time.Now()},
				{Key: "updated_at", Value: time.Now()},
			})
			if err != nil {

				panic(err)
			}

			c.JSON(http.StatusOK, res)
		})
		todo.GET("/list", func(c *gin.Context) {
			var notes []bson.M

			collection := db.Client.Database("todo").Collection("notes")
			cursor, err := collection.Find(context.TODO(), bson.D{})
			err = cursor.All(context.TODO(), &notes)
			if err != nil {
				panic(err)
			}
			c.JSON(http.StatusOK, notes)
		})
		todo.PUT("/update/:id", func(c *gin.Context) {
			var req ToUpdateNote
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest,
					gin.H{
						"error":   "VALIDATEERR-1",
						"message": "Invalid inputs. Please check your inputs"})
				return
			}

			id, _ := primitive.ObjectIDFromHex(c.Param("id"))
			filter := bson.M{"_id": id}
			update := bson.M{
				"$set": bson.M{
					"title":      req.Title,
					"content":    req.Content,
					"updated_at": time.Now(),
				},
			}

			collection := db.Client.Database("todo").Collection("notes")
			res, err := collection.UpdateOne(context.TODO(), filter, update)
			if err != nil {
				panic(err)
			}
			c.JSON(http.StatusOK, res)
		})
		todo.DELETE("/delete/:id", func(c *gin.Context) {
			id, _ := primitive.ObjectIDFromHex(c.Param("id"))
			filter := bson.M{"_id": id}

			collection := db.Client.Database("todo").Collection("notes")
			res, err := collection.DeleteOne(context.TODO(), filter)
			if err != nil {
				panic(err)
			}
			c.JSON(http.StatusOK, res)
		})
	}
}
