package handlers

import (
	"context"
	"errors"
	"fmt"
	"go-visio-service/models"
	"go-visio-service/utils"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"go-visio-service/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin:     func(r *http.Request) bool { return true },
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	roomsMu     sync.RWMutex
	rooms       = make(map[string]*models.Room)
	mongoClient *mongo.Client
)

type WSMessage struct {
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
	From   string      `json:"from,omitempty"`
	Target string      `json:"target,omitempty"`
}

func SetMongoClient(client *mongo.Client) {
	mongoClient = client
}

func SignalHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("id")
		log.Printf("[SignalHandler] Tentative de connexion à la room: %s", roomID)

		// Extraction du token JWT
		token := extractToken(c)
		userID, username, _ := utils.ParseUserFromJWT(token)
		if userID == "" || username == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Upgrade WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}
		defer conn.Close()

		// Configurer les timeouts
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// Gestion de la room
		room, err := getOrCreateRoom(roomID)
		if err != nil {
			conn.WriteJSON(WSMessage{Type: "error", Data: fmt.Sprintf("Error: %s", err.Error())})
			return
		}

		// Vérification des permissions
		if err := checkWorkspaceAccess(room, userID); err != nil {
			conn.WriteJSON(WSMessage{Type: "error", Data: err.Error()})
			return
		}

		// Création du participant
		participant := createParticipant(conn, userID, username, room.AdminID)

		// Ajout du participant à la room
		addParticipantToRoom(room, userID, participant)
		defer removeParticipantFromRoom(room, userID, username)

		// Envoi des informations initiales
		sendInitialInfo(conn, room, userID, username)

		// Boucle de lecture des messages
		handleMessages(conn, roomID, userID, username)
	}
}

// Fonctions utilitaires

func extractToken(c *gin.Context) string {
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}
	return token
}

func getOrCreateRoom(roomID string) (*models.Room, error) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	if room, exists := rooms[roomID]; exists {
		return room, nil
	}

	if mongoClient == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var dbRoom models.Room
	err := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("rooms").
		FindOne(context.Background(), bson.M{"roomId": roomID}).Decode(&dbRoom)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}

	dbRoom.Participants = make(map[string]*models.Participant)
	rooms[roomID] = &dbRoom
	return &dbRoom, nil
}

func checkWorkspaceAccess(room *models.Room, userID string) error {
	if room.WorkspaceID == "" || mongoClient == nil {
		return nil
	}

	hasAccess, err := middleware.HasWorkspaceAccess(mongoClient, userID, room.WorkspaceID)
	if err != nil || !hasAccess {
		return errors.New("access denied to workspace")
	}
	return nil
}

func createParticipant(ws *websocket.Conn, userID, username, adminID string) *models.Participant {
	role := "participant"
	if userID == adminID {
		role = "admin"
	}
	return &models.Participant{
		Conn:     ws,
		UserID:   userID,
		Username: username,
		Role:     role,
	}
}

func addParticipantToRoom(room *models.Room, userID string, participant *models.Participant) {
	roomsMu.Lock()
	room.Participants[userID] = participant
	roomsMu.Unlock()

	log.Printf("[JOIN] User %s (%s) joined room %s", userID, participant.Username, room.RoomID)

	broadcast(room.RoomID, WSMessage{
		Type: "peer_joined",
		From: userID,
		Data: map[string]string{
			"userID":   userID,
			"username": participant.Username,
		},
	})
	sendParticipantList(room)
}

func removeParticipantFromRoom(room *models.Room, userID, username string) {
	roomsMu.Lock()
	delete(room.Participants, userID)
	empty := len(room.Participants) == 0
	if empty {
		delete(rooms, room.RoomID)
	}
	roomsMu.Unlock()

	log.Printf("[LEAVE] User %s (%s) left room %s", userID, username, room.RoomID)

	broadcast(room.RoomID, WSMessage{
		Type: "peer_left",
		From: userID,
		Data: map[string]string{
			"userID":   userID,
			"username": username,
		},
	})
	sendParticipantList(room)
}

func sendInitialInfo(ws *websocket.Conn, room *models.Room, userID, username string) {
	// Envoyer les pairs existants
	roomsMu.RLock()
	var existingPeers []map[string]string
	for id, p := range room.Participants {
		if id != userID {
			existingPeers = append(existingPeers, map[string]string{
				"userID":   id,
				"username": p.Username,
			})
		}
	}
	roomsMu.RUnlock()

	if len(existingPeers) > 0 {
		ws.WriteJSON(WSMessage{
			Type: "existing_peers",
			Data: existingPeers,
		})
	}

	// Envoyer l'historique des messages
	if mongoClient != nil {
		sendMessageHistory(ws, room.RoomID)
	}
}

func sendMessageHistory(ws *websocket.Conn, roomID string) {
	cur, err := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("messages").
		Find(context.Background(), bson.M{"roomId": roomID})
	if err != nil {
		return
	}

	var messages []models.Message
	if err := cur.All(context.Background(), &messages); err == nil {
		for _, m := range messages {
			ws.WriteJSON(WSMessage{
				Type: "chat",
				Data: map[string]interface{}{
					"user":    m.Username,
					"message": m.Content,
					"time":    m.CreatedAt.Time().Format(time.RFC3339),
				},
			})
		}
	}
}

func handleMessages(ws *websocket.Conn, roomID, userID, username string) {
	for {
		var msg WSMessage
		if err := ws.ReadJSON(&msg); err != nil {
			log.Printf("[WS] Read error from %s (%s): %v", userID, username, err)
			break
		}

		msg.From = userID
		if err := handleWSMessage(roomID, userID, msg); err != nil {
			log.Printf("[WS] Error handling message from %s (%s): %v", userID, username, err)
			break
		}
	}
}

func handleWSMessage(roomID, senderID string, msg WSMessage) error {
	roomsMu.RLock()
	room, exists := rooms[roomID]
	sender, senderExists := room.Participants[senderID]
	roomsMu.RUnlock()

	if !exists || !senderExists {
		return fmt.Errorf("room or participant not found")
	}

	switch msg.Type {
	case "offer", "answer", "candidate":
		if err := validateWebRTCMessage(msg); err != nil {
			return err
		}
		routeWebRTCMessage(roomID, msg)

	case "chat":
		return handleChatMessage(roomID, sender, msg)

	case "mute", "unmute", "video_on", "video_off", "screen_share_start", "screen_share_stop":
		updateParticipantState(room, senderID, msg.Type)
		broadcast(roomID, msg)
		sendParticipantList(room)

	case "kick":
		return handleKickCommand(room, sender, msg)

	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

	return nil
}

func validateWebRTCMessage(msg WSMessage) error {
	if _, ok := msg.Data.(map[string]interface{}); !ok {
		return fmt.Errorf("invalid WebRTC message format")
	}
	return nil
}

func routeWebRTCMessage(roomID string, msg WSMessage) {
	if msg.Target == "" || msg.Target == "broadcast" {
		broadcast(roomID, msg)
	} else {
		sendToParticipant(roomID, msg.Target, msg)
	}
}

func handleChatMessage(roomID string, sender *models.Participant, msg WSMessage) error {
	messageData, ok := msg.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid chat message format")
	}

	messageText, ok := messageData["message"].(string)
	if !ok || messageText == "" {
		return fmt.Errorf("empty chat message")
	}

	if mongoClient != nil {
		_, err := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("messages").InsertOne(
			context.Background(),
			models.Message{
				RoomID:    roomID,
				UserID:    sender.UserID,
				Username:  sender.Username,
				Content:   messageText,
				CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
			},
		)
		if err != nil {
			log.Printf("Failed to save chat message: %v", err)
		}
	}

	broadcast(roomID, WSMessage{
		Type: "chat",
		From: sender.UserID,
		Data: map[string]interface{}{
			"user":    sender.Username,
			"message": messageText,
			"time":    time.Now().Format(time.RFC3339),
		},
	})

	return nil
}

func updateParticipantState(room *models.Room, userID, action string) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	participant := room.Participants[userID]
	if participant == nil {
		return
	}

	switch action {
	case "mute":
		participant.AudioMuted = true
	case "unmute":
		participant.AudioMuted = false
	case "video_on":
		participant.VideoOff = false
	case "video_off":
		participant.VideoOff = true
	case "screen_share_start":
		participant.ScreenSharing = true
	case "screen_share_stop":
		participant.ScreenSharing = false
	}
}

func handleKickCommand(room *models.Room, sender *models.Participant, msg WSMessage) error {
	if sender.Role != "admin" {
		return fmt.Errorf("unauthorized")
	}

	targetID, ok := msg.Data.(map[string]interface{})["userID"].(string)
	if !ok {
		return fmt.Errorf("invalid target")
	}

	roomsMu.Lock()
	target, exists := room.Participants[targetID]
	roomsMu.Unlock()

	if exists {
		target.Conn.Close()
		roomsMu.Lock()
		delete(room.Participants, targetID)
		empty := len(room.Participants) == 0
		if empty {
			delete(rooms, room.RoomID)
		}
		roomsMu.Unlock()

		log.Printf("[KICK] User %s (%s) kicked by %s (%s) from room %s", targetID, target.Username, sender.UserID, sender.Username, room.RoomID)

		broadcast(room.RoomID, WSMessage{
			Type: "peer_kicked",
			From: sender.UserID,
			Data: map[string]string{
				"targetID": targetID,
				"username": target.Username,
			},
		})
		sendParticipantList(room)
	}

	return nil
}

// broadcast envoie à tous les participants sauf l'expéditeur, et nettoie les connexions mortes
func broadcast(roomID string, msg WSMessage) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	room, exists := rooms[roomID]
	if !exists {
		return
	}

	for userID, p := range room.Participants {
		if p.UserID != msg.From {
			if err := p.Conn.WriteJSON(msg); err != nil {
				log.Printf("[WS] Failed to broadcast to %s (%s): %v. Removing from room.", p.UserID, p.Username, err)
				p.Conn.Close()
				delete(room.Participants, userID)
			}
		}
	}
	if len(room.Participants) == 0 {
		delete(rooms, room.RoomID)
	}
}

// sendToParticipant envoie à un participant ciblé, et nettoie la connexion morte si besoin
func sendToParticipant(roomID, targetID string, msg WSMessage) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	room, exists := rooms[roomID]
	if !exists {
		return
	}

	target, exists := room.Participants[targetID]
	if exists {
		if err := target.Conn.WriteJSON(msg); err != nil {
			log.Printf("[WS] Failed to send to %s (%s): %v. Removing from room.", targetID, target.Username, err)
			target.Conn.Close()
			delete(room.Participants, targetID)
		}
	}
	if len(room.Participants) == 0 {
		delete(rooms, room.RoomID)
	}
}

func sendParticipantList(room *models.Room) {
	roomsMu.RLock()
	defer roomsMu.RUnlock()

	var participants []map[string]interface{}
	for _, p := range room.Participants {
		participants = append(participants, map[string]interface{}{
			"userID":        p.UserID,
			"username":      p.Username,
			"role":          p.Role,
			"audioMuted":    p.AudioMuted,
			"videoOff":      p.VideoOff,
			"screenSharing": p.ScreenSharing,
		})
	}

	msg := WSMessage{Type: "participants", Data: participants}
	for _, p := range room.Participants {
		if err := p.Conn.WriteJSON(msg); err != nil {
			log.Printf("[WS] Failed to send participant list to %s (%s): %v", p.UserID, p.Username, err)
		}
	}
}
