package main

import (
	"context"
	"go-visio-service/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aucun fichier .env trouvé ou erreur de chargement, on continue avec l'environnement système")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI n'est pas défini dans l'environnement")
	}
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Erreur de connexion à MongoDB: %v", err)
	}
	handlers.SetMongoClient(mongoClient)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.POST("/api/visio/room", handlers.CreateRoomHandler())
	r.GET("/ws/room/:id", handlers.SignalHandler())
	r.Run(":8081")
}
