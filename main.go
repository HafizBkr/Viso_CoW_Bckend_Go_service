package main

import (
	"context"
	"go-visio-service/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Charger les variables d'environnement
	err := godotenv.Load()
	if err != nil {
		log.Println("Aucun fichier .env trouvé ou erreur de chargement, on continue avec l'environnement système")
	}

	// Connexion MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI n'est pas défini dans l'environnement")
	}

	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Erreur de connexion à MongoDB: %v", err)
	}

	// Injecter le client dans les handlers
	handlers.SetMongoClient(mongoClient)

	// Initialiser Gin
	r := gin.Default()

	// Endpoint de santé
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ✅ Route de test d'accès (à utiliser pour vérifier le middleware)
	r.GET("/debug/workspaces/:userId", func(c *gin.Context) {
		userId := c.Param("userId")
		db := mongoClient.Database(os.Getenv("MONGO_DBNAME"))
		ctx := context.Background()

		objUserId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
			return
		}

		cursor, err := db.Collection("workspaces").Find(ctx, bson.M{"createdBy": objUserId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workspaces"})
			return
		}
		var workspaces []bson.M
		if err = cursor.All(ctx, &workspaces); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse workspaces"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"workspaces": workspaces})
	})

	// Routes de visio
	r.POST("/api/visio/room", handlers.CreateRoomHandler())
	r.GET("/ws/room/:id", handlers.SignalHandler())

	// Lancer le serveur
	err = r.Run(":8081")
	if err != nil {
		log.Fatalf("Erreur de démarrage du serveur : %v", err)
	}
}
