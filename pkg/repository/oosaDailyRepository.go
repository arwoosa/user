package repository

import (
	"context"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OosaDailyRepository struct{}

func (r OosaDailyRepository) Watched(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)

	now := time.Now()

	periodStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	periodEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	var todayExp models.Exp
	filter := bson.D{
		{Key: "exp_user", Value: userDetail.UsersId},
		{Key: "exp_source", Value: helpers.EXP_OOSA_DAILY},
		{Key: "exp_created_at", Value: bson.D{
			{Key: "$gte", Value: primitive.NewDateTimeFromTime(periodStart)},
			{Key: "$lt", Value: primitive.NewDateTimeFromTime(periodEnd)},
		}},
	}
	config.DB.Collection("Exp").FindOne(context.TODO(), filter).Decode(&todayExp)
	if !helpers.MongoZeroID(todayExp.ExpId) {
		helpers.ResponseBadRequestError(c, "Exp already awarded for today")
		return
	}
	helpers.ExpAward(c, helpers.EXP_OOSA_DAILY, 1, primitive.NilObjectID)
	helpers.ResponseSuccessMessage(c, "Exp awarded")
}
