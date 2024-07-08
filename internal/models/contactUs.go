package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type ContactUs struct {
	ContactUsId        primitive.ObjectID `bson:"_id,omitempty" json:"contact_us_id"`
	ContactUsEmail     string             `bson:"contact_us_email,omitempty" json:"contact_us_email"`
	ContactUsMessage   string             `bson:"contact_us_message,omitempty" json:"contact_us_message"`
	ContactUsCreatedAt primitive.DateTime `bson:"contact_us_created_at,omitempty" json:"contact_us_created_at"`
}
