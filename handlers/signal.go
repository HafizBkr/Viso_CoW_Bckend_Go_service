package handlers

import (
	"context"
	"go-visio-service/models"
	"go-visio-service/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func SignalHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("id")
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		userID, username, _ := utils.ParseUserFromJWT(token)
		log.Printf("Parsed JWT: userID=%s, username=%s", userID, username)

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}
		defer ws.Close()

		roomsMu.Lock()
		room, ok := rooms[roomID]
		if !ok {
			if mongoClient != nil {
				var dbRoom models.Room
				err := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("rooms").FindOne(context.Background(), bson.M{"roomId": roomID}).Decode(&dbRoom)
				if err == nil {
					room = &dbRoom
					room.Participants = make(map[string]*models.Participant)
					rooms[roomID] = room
				} else {
					roomsMu.Unlock()
					ws.WriteJSON(WSMessage{Type: "error", Data: "Room not found"})
					return
				}
			}
		}
		role := "participant"
		if userID == room.AdminID {
			role = "admin"
		}
		participant := &models.Participant{
			Conn:     ws,
			UserID:   userID,
			Username: username,
			Role:     role,
		}
		room.Participants[userID] = participant
		roomsMu.Unlock()

		broadcast(roomID, WSMessage{Type: "join", Data: map[string]string{"user": username}})
		sendParticipantList(room)

		if mongoClient != nil {
			cur, err := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("messages").
				Find(context.Background(), bson.M{"roomId": roomID})
			if err == nil {
				var messages []models.Message
				if err := cur.All(context.Background(), &messages); err == nil {
					for _, m := range messages {
						ws.WriteJSON(WSMessage{Type: "chat", Data: map[string]interface{}{
							"user":    m.Username,
							"message": m.Content,
						}})
					}
				}
			}
		}

		for {
			var msg WSMessage
			if err := ws.ReadJSON(&msg); err != nil {
				break
			}
			handleWSMessage(roomID, userID, msg)
		}

		roomsMu.Lock()
		delete(room.Participants, userID)
		roomsMu.Unlock()
		broadcast(roomID, WSMessage{Type: "leave", Data: map[string]string{"user": username}})
		sendParticipantList(room)
	}
}

func broadcast(roomID string, msg WSMessage) {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	room, ok := rooms[roomID]
	if !ok {
		return
	}
	for _, p := range room.Participants {
		p.Conn.WriteJSON(msg)
	}
}

func sendParticipantList(room *models.Room) {
	var list []map[string]interface{}
	for _, p := range room.Participants {
		list = append(list, map[string]interface{}{
			"userID":        p.UserID,
			"username":      p.Username,
			"role":          p.Role,
			"audioMuted":    p.AudioMuted,
			"videoOff":      p.VideoOff,
			"screenSharing": p.ScreenSharing,
		})
	}
	for _, p := range room.Participants {
		p.Conn.WriteJSON(WSMessage{Type: "participants", Data: list})
	}
}

func handleWSMessage(roomID, senderID string, msg WSMessage) {
	roomsMu.Lock()
	room := rooms[roomID]
	sender := room.Participants[senderID]
	roomsMu.Unlock()

	switch msg.Type {
	case "offer", "answer", "candidate":
		broadcast(roomID, msg)
	case "chat":
		if mongoClient != nil {
			m := models.Message{
				RoomID:    roomID,
				UserID:    sender.UserID,
				Username:  sender.Username,
				Content:   msg.Data.(map[string]interface{})["message"].(string),
				CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
			}
			mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("messages").InsertOne(context.Background(), m)
		}
		broadcast(roomID, WSMessage{Type: "chat", Data: map[string]interface{}{
			"user":    sender.Username,
			"message": msg.Data.(map[string]interface{})["message"],
		}})
	case "mute":
		sender.AudioMuted = true
		broadcast(roomID, WSMessage{Type: "mute", Data: map[string]string{"user": sender.Username}})
		sendParticipantList(room)
	case "unmute":
		sender.AudioMuted = false
		broadcast(roomID, WSMessage{Type: "unmute", Data: map[string]string{"user": sender.Username}})
		sendParticipantList(room)
	case "video_on":
		sender.VideoOff = false
		broadcast(roomID, WSMessage{Type: "video_on", Data: map[string]string{"user": sender.Username}})
		sendParticipantList(room)
	case "video_off":
		sender.VideoOff = true
		broadcast(roomID, WSMessage{Type: "video_off", Data: map[string]string{"user": sender.Username}})
		sendParticipantList(room)
	case "screen_share_start":
		sender.ScreenSharing = true
		broadcast(roomID, WSMessage{Type: "screen_share_start", Data: map[string]string{"user": sender.Username}})
		sendParticipantList(room)
	case "screen_share_stop":
		sender.ScreenSharing = false
		broadcast(roomID, WSMessage{Type: "screen_share_stop", Data: map[string]string{"user": sender.Username}})
		sendParticipantList(room)
	case "kick":
		if sender.Role == "admin" {
			targetID := msg.Data.(map[string]interface{})["userID"].(string)
			roomsMu.Lock()
			target, ok := room.Participants[targetID]
			roomsMu.Unlock()
			if ok {
				target.Conn.Close()
				roomsMu.Lock()
				delete(room.Participants, targetID)
				roomsMu.Unlock()
				broadcast(roomID, WSMessage{Type: "kick", Data: map[string]string{"user": target.Username}})
				sendParticipantList(room)
			}
		}
	}
}
