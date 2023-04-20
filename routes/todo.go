// routes/notes.go
package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	db "ginserver/config"
	"ginserver/middlewares"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

const (
	usersCollection = "users"
	notesCollection = "notes"
)

var (
	errInvalidInput = gin.H{
		"error":   "VALIDATEERR-1",
		"message": "Invalid inputs. Please check your inputs",
	}
	errInvalidUser = gin.H{
		"error":   "VALIDATEERR-2",
		"message": "Invalid user",
	}
)

func userExists(userID string) bool {
	collection := db.Client.Database("todo").Collection(usersCollection)
	filter := bson.M{"id": userID, "deleted": false}
	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		panic(err)
	}
	return count > 0
}

func Todo(r *gin.Engine) {
	todo := r.Group("/todo")
	todo.Use(middlewares.AuthMiddleware())
	{
		todo.POST("/create", func(c *gin.Context) {
			var req Note
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, errInvalidInput)
				return
			}

			userID := c.GetString("user_id")

			fmt.Print(userID)
			if !userExists(userID) {
				c.AbortWithStatusJSON(http.StatusBadRequest, errInvalidUser)
				return
			}

			collection := db.Client.Database("todo").Collection(notesCollection)
			res, err := collection.InsertOne(context.TODO(), bson.D{
				{Key: "title", Value: req.Title},
				{Key: "content", Value: req.Content},
				{Key: "user_id", Value: userID},
				{Key: "created_at", Value: time.Now()},
				{Key: "updated_at", Value: time.Now()},
			})
			if err != nil {
				panic(err)
			}

			c.JSON(http.StatusOK, gin.H{
				"id":      res.InsertedID,
				"message": "Note created",
			})
		})
		todo.GET("/list", func(c *gin.Context) {
			var notes []bson.M

			collection := db.Client.Database("todo").Collection("notes")

			user_id := c.GetString("user_id")

			cursor, err := collection.Find(context.TODO(), bson.D{
				{Key: "user_id", Value: user_id},
			})

			err = cursor.All(context.TODO(), &notes)
			if err != nil {
				panic(err)
			}
			c.JSON(http.StatusOK, notes)
		})
		todo.PUT("/update/:id", func(c *gin.Context) {
			id := c.Param("id")
			objectID, err := primitive.ObjectIDFromHex(id)

			var req Note
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "VALIDATEERR-1",
					"message": "Invalid inputs. Please check your inputs",
				})
				return
			}

			if req.Title == "" && req.Content == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "VALIDATEERR-2",
					"message": "Both title and content cannot be empty",
				})
				return
			}

			user_id := c.GetString("user_id")
			collection := db.Client.Database("todo").Collection("notes")
			// Check if the note exists and belongs to the current user
			filter := bson.D{{Key: "_id", Value: objectID}, {Key: "user_id", Value: user_id}}

			var note Note
			err = collection.FindOne(context.Background(), filter).Decode(&note)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
						"error":   "NOTFOUND-1",
						"message": "Note not found",
					})
					return
				}
				panic(err)
			}

			// Update the note
			update := bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "updated_at", Value: time.Now()},
				}},
			}
			if req.Title != "" {
				update[0].Value = append(update[0].Value.(bson.D), bson.E{Key: "title", Value: req.Title})
			}
			if req.Content != "" {
				update[0].Value = append(update[0].Value.(bson.D), bson.E{Key: "content", Value: req.Content})
			}
			res, err := collection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				panic(err)
			}

			if res.ModifiedCount == 0 {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
					"error":   "NOTFOUND-1",
					"message": "Note not found",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Note updated",
			})
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
