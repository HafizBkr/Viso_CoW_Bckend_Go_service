package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Room struct {
	ID           primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
	RoomID       string                  `bson:"roomId" json:"roomId"` // UUID
	WorkspaceID  string                  `bson:"workspaceId" json:"workspaceId"`
	AdminID      string                  `bson:"adminId" json:"adminId"`
	CreatedAt    primitive.DateTime      `bson:"createdAt" json:"createdAt"`
	Participants map[string]*Participant `bson:"-" json:"participants,omitempty"` // Pas stocké en DB
}
