package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type PasswordResets struct {
	PasswordResetsId        primitive.ObjectID `bson:"_id,omitempty" json:"password_resets_id"`
	PasswordResetsUserId    primitive.ObjectID `bson:"password_resets_user_id,omitempty" json:"password_resets_user_id"`
	PasswordResetsEmail     string             `bson:"password_resets_email,omitempty" json:"password_resets_email"`
	PasswordResetsToken     string             `bson:"password_resets_token,omitempty" json:"password_resets_token"`
	PasswordResetsIsActive  int                `bson:"password_resets_is_active,omitempty" json:"password_resets_is_active"`
	PasswordResetsCreatedAt primitive.DateTime `bson:"password_resets_created_at,omitempty" json:"password_resets_created_at"`
}
