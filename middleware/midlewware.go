package middleware

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func HasWorkspaceAccess(mongoClient *mongo.Client, userID, workspaceID string) (bool, error) {
	db := mongoClient.Database(os.Getenv("MONGO_DBNAME"))
	ctx := context.Background()

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("Invalid userID hex: %v", err)
		return false, err
	}

	workspaceObjID, err := primitive.ObjectIDFromHex(workspaceID)
	if err != nil {
		log.Printf("Invalid workspaceID hex: %v", err)
		return false, err
	}

	log.Printf("Checking workspace existence for ID %s", workspaceID)
	count, err := db.Collection("workspaces").CountDocuments(ctx, bson.M{"_id": workspaceObjID})
	if err != nil {
		log.Printf("Error checking workspace existence: %v", err)
		return false, err
	}
	if count == 0 {
		log.Printf("No workspace found with that ID")
		return false, nil
	}

	log.Printf("Checking if user %s is creator", userID)
	creatorCount, err := db.Collection("workspaces").CountDocuments(ctx, bson.M{"_id": workspaceObjID, "createdBy": userObjID})
	if err != nil {
		log.Printf("Error checking creator: %v", err)
		return false, err
	}
	if creatorCount > 0 {
		log.Printf("User is creator of workspace")
		return true, nil
	}

	log.Printf("Checking if user %s is a member", userID)
	memberCount, err := db.Collection("workspacemembers").CountDocuments(ctx, bson.M{
		"workspace":      workspaceObjID,
		"user":           userObjID,
		"inviteAccepted": true,
	})
	if err != nil {
		log.Printf("Error checking membership: %v", err)
		return false, err
	}
	if memberCount > 0 {
		log.Printf("User is a member of workspace")
		return true, nil
	}

	log.Printf("User has no access to workspace")
	return false, nil
}
