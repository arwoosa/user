package helpers

import (
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func MongoZeroID(a primitive.ObjectID) bool {
	zeroValue, _ := primitive.ObjectIDFromHex("000000000000000000000000")
	return a == zeroValue
}

func RandomString(length int) string {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}

func MongoTimestampToTime(datetime primitive.DateTime) time.Time {
	return datetime.Time()
}
