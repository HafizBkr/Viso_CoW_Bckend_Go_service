package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
    ID            primitive.ObjectID `bson:"_id" json:"id"`
    Email         string             `bson:"email" json:"email"`
    Username      string             `bson:"username" json:"username"`
    Avatar        string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
    Bio           string             `bson:"bio,omitempty" json:"bio,omitempty"`
    Location      string             `bson:"location,omitempty" json:"location,omitempty"`
    EmailVerified bool               `bson:"emailVerified" json:"emailVerified"`
    OnlineStatus  bool               `bson:"onlineStatus" json:"onlineStatus"`
    VideoEnabled  bool               `bson:"videoEnabled" json:"videoEnabled"`
    AudioEnabled  bool               `bson:"audioEnabled" json:"audioEnabled"`
    CreatedAt     primitive.DateTime `bson:"createdAt" json:"createdAt"`
}
