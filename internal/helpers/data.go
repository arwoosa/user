package helpers

import (
	"math/rand"
	"time"

	"github.com/dlclark/regexp2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	REGEX_USERNAME = "^(?=[a-zA-Z._]{2,20}$)(?=(.*?[A-Z]){1})(?=(.?[a-z]){1}).*"
	REGEX_NAME     = "^(?=[\u4e00-\u9fa5a-zA-Z ]{1,20}$)(?=(.*?[A-Z]){1})(?=(.*?[a-z]){1}).*"
	REGEX_PASSWORD = "^(?=[a-zA-Z0-9]{6,12}$)(?=(.*?[A-Z]){1})(?=(.*?[a-z]){1})(?=.*?[0-9]).*"
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

func MonthInterval(y int, m time.Month) (firstDay, lastDay time.Time) {
	firstDay = time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	lastDay = time.Date(y, m+1, 1, 0, 0, 0, -1, time.UTC)
	return firstDay, lastDay
}

func RegexCompare(regexString string, stringTest string) bool {
	re := regexp2.MustCompile(regexString, 0)
	isMatch, _ := re.MatchString(stringTest)
	return isMatch
}
