package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Workspace struct {
    ID          primitive.ObjectID `bson:"_id" json:"id"`
    Name        string             `bson:"name" json:"name"`
    Description string             `bson:"description,omitempty" json:"description,omitempty"`
    Logo        string             `bson:"logo,omitempty" json:"logo,omitempty"`
    CreatedBy   primitive.ObjectID `bson:"createdBy" json:"createdBy"`
    CreatedAt   primitive.DateTime `bson:"createdAt" json:"createdAt"`
}
