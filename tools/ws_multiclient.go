package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
	From   string      `json:"from,omitempty"`
	Target string      `json:"target,omitempty"`
}

func runClient(wg *sync.WaitGroup, id int, roomID, token, serverAddr string, messages []WSMessage) {
	defer wg.Done()
	u := url.URL{
		Scheme:   "ws",
		Host:     serverAddr,
		Path:     fmt.Sprintf("/ws/room/%s", roomID),
		RawQuery: "token=" + token,
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("[Client %d] dial error: %v", id, err)
		return
	}
	defer c.Close()

	// Goroutine de réception
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("[Client %d] read error: %v", id, err)
				return
			}
			fmt.Printf("[Client %d] Received: %s\n", id, message)
		}
	}()

	// Envoi des messages
	for _, msg := range messages {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		b, _ := json.Marshal(msg)
		err := c.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			log.Printf("[Client %d] write error: %v", id, err)
			return
		}
		fmt.Printf("[Client %d] Sent: %s\n", id, b)
	}
	// Laisse le client ouvert un peu pour recevoir les réponses
	time.Sleep(2 * time.Second)
}

func main() {
	roomID := "5fb92b92-ce70-4e50-a791-8ec89dc5f040"
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2ODdhZGUzYmQwMGZlODExYjk4ZTk5M2EiLCJlbWFpbCI6ImJvdWthcmloYWZpejRAZ21haWwuY29tIiwidXNlcm5hbWUiOiJNaWNoZWwiLCJpYXQiOjE3NTMwMDIwNzgsImV4cCI6MTc1MzE3NDg3OH0.xI4osaAVyOsRnfwTc8sTyp_gZi0jD5Rh4dqVul_Mrrk"
	serverAddr := "localhost:8081" // ou ton adresse réelle

	// Messages à envoyer par chaque client
	messages := []WSMessage{
		{Type: "chat", Data: map[string]interface{}{"message": "hello from client"}},
		{Type: "offer", Data: map[string]interface{}{"sdp": "fakeSDP"}},
		{Type: "candidate", Data: map[string]interface{}{"candidate": "fakeICE"}},
	}

	numClients := 3 // Nombre de clients à lancer

	var wg sync.WaitGroup
	for i := 1; i <= numClients; i++ {
		wg.Add(1)
		go runClient(&wg, i, roomID, token, serverAddr, messages)
	}
	wg.Wait()
	fmt.Println("All clients finished.")
}
