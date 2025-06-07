package middleware

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func HasWorkspaceAccess(mongoClient *mongo.Client, userID, workspaceID string) (bool, error) {
	// Vérifie que workspaceID et userID sont des ObjectId valides
	workspaceObjID, err := primitive.ObjectIDFromHex(workspaceID)
	if err != nil {
		return false, err
	}
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	// Vérifie que le workspace existe
	workspaceFilter := bson.M{"_id": workspaceObjID}
	count, err := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("workspaces").CountDocuments(context.Background(), workspaceFilter)
	if err != nil || count == 0 {
		return false, nil // workspace not found
	}

	// Vérifie que le user est membre du workspace et que inviteAccepted = true
	memberFilter := bson.M{
		"workspace":      workspaceObjID,
		"user":           userObjID,
		"inviteAccepted": true,
	}
	count, err = mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("workspacemembers").CountDocuments(context.Background(), memberFilter)
	if err != nil || count == 0 {
		return false, nil // not a member or not accepted
	}

	return true, nil
}
