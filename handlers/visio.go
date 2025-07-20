package handlers

import (
	"context"
	"go-visio-service/middleware"
	"go-visio-service/models"
	"go-visio-service/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Déclarations globales utilisées dans tout le package handlers

func CreateRoomHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		if token == "" {
			c.JSON(401, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}

		userID, _, _ := utils.ParseUserFromJWT(token)
		if userID == "" {
			c.JSON(401, gin.H{"error": "Invalid or expired JWT"})
			return
		}
		type CreateRoomPayload struct {
			WorkspaceID string `json:"workspaceId"`
		}
		var payload CreateRoomPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		workspaceID := payload.WorkspaceID
		if workspaceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing workspaceId"})
			return
		}

		hasAccess, err := middleware.HasWorkspaceAccess(mongoClient, userID, workspaceID)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid workspace or user ID"})
			return
		}
		if !hasAccess {
			c.JSON(403, gin.H{"error": "You do not have access to this workspace"})
			return
		}

		roomID := uuid.New().String()
		room := &models.Room{
			RoomID:       roomID,
			WorkspaceID:  workspaceID,
			AdminID:      userID,
			CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
			Participants: make(map[string]*models.Participant),
		}

		roomsMu.Lock()
		rooms[roomID] = room
		roomsMu.Unlock()

		if mongoClient != nil {
			_, err := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("rooms").InsertOne(context.Background(), room)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist room"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"roomId": roomID})
	}
}

func ListRoomsByWorkspaceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.Param("workspaceId")
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		userID, _, _ := utils.ParseUserFromJWT(token)
		if userID == "" {
			c.JSON(401, gin.H{"error": "Invalid or expired JWT"})
			return
		}

		hasAccess, err := middleware.HasWorkspaceAccess(mongoClient, userID, workspaceID)
		if err != nil || !hasAccess {
			c.JSON(403, gin.H{"error": "Access denied to workspace"})
			return
		}

		var roomsList []models.Room
		cursor, err := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("rooms").Find(
			context.Background(),
			bson.M{"workspaceId": workspaceID},
		)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to fetch rooms"})
			return
		}
		if err := cursor.All(context.Background(), &roomsList); err != nil {
			c.JSON(500, gin.H{"error": "Failed to decode rooms"})
			return
		}

		// Optionnel : ne retourne que les champs utiles
		var result []gin.H
		for _, r := range roomsList {
			result = append(result, gin.H{
				"roomId":    r.RoomID,
				"adminId":   r.AdminID,
				"createdAt": r.CreatedAt,
			})
		}
		c.JSON(200, result)
	}
}
