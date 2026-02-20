package models

import (
"time"
"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
Name      string             `bson:"name" json:"name"`
Email     string             `bson:"email" json:"email"`
Phone     string             `bson:"phone" json:"phone"`
Level     string             `bson:"level" json:"level"`
Password  string             `bson:"password" json:"-"`
Code      string             `bson:"code,omitempty" json:"code,omitempty"`
Verified  bool               `bson:"verified" json:"verified"`
CreatedAt time.Time          `bson:"created_at" json:"created_at"`
UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// Add this Admin struct - keep it separate from User
type Admin struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Email     string             `bson:"email" json:"email"`
    Password  string             `bson:"password" json:"password"`
    Name      string             `json:"name" bson:"name"`
    CreatedAt primitive.DateTime `bson:"createdAt" json:"createdAt"`
    IsActive  bool               `bson:"isActive" json:"isActive"`
}