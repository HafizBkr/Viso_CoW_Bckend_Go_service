package middleware

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func HasWorkspaceAccess(mongoClient *mongo.Client, userID, workspaceID string) (bool, error) {
	workspaceObjID, err := primitive.ObjectIDFromHex(workspaceID)
	if err != nil {
		return false, err
	}
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}
	workspaceFilter := bson.M{"_id": workspaceObjID}
	count, err := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("workspaces").CountDocuments(context.Background(), workspaceFilter)
	if err != nil || count == 0 {
		return false, nil
	}
	memberFilter := bson.M{
		"workspace":      workspaceObjID,
		"user":           userObjID,
		"inviteAccepted": true,
	}
	count, err = mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("workspacemembers").CountDocuments(context.Background(), memberFilter)
	if err != nil || count == 0 {
		return false, nil
	}

	return true, nil
}
