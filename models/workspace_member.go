package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type WorkspaceMember struct {
	ID              primitive.ObjectID `bson:"_id" json:"id"`
	Workspace       primitive.ObjectID `bson:"workspace" json:"workspace"`
	User            primitive.ObjectID `bson:"user,omitempty" json:"user,omitempty"`
	Email           string             `bson:"email" json:"email"`
	Role            string             `bson:"role" json:"role"`
	InvitedBy       primitive.ObjectID `bson:"invitedBy" json:"invitedBy"`
	InviteAccepted  bool               `bson:"inviteAccepted" json:"inviteAccepted"`
	CurrentPosition *Position          `bson:"currentPosition,omitempty" json:"currentPosition,omitempty"`
	LastActive      primitive.DateTime `bson:"lastActive" json:"lastActive"`
	CreatedAt       primitive.DateTime `bson:"createdAt" json:"createdAt"`
}

type Position struct {
	X float64 `bson:"x" json:"x"`
	Y float64 `bson:"y" json:"y"`
}
