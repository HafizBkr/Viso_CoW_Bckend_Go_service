package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID    string             `bson:"roomId" json:"roomId"`
	UserID    string             `bson:"userId" json:"userId"`
	Username  string             `bson:"username" json:"username"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
}
