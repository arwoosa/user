package repository

import (
	"context"
	"net/http"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserStatisticsRepository struct{}

func (t UserStatisticsRepository) Retrieve(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)

	response := gin.H{
		"statistics_last_rewilding":     "",
		"statistics_rewilding_by_month": []map[string]any{},
	}

	var lastRewilding []models.Events
	lastRewildingAgg := mongo.Pipeline{
		bson.D{{
			Key: "$match", Value: bson.M{"event_participants_user": userDetail.UsersId, "event_participants_status": 1},
		}},
		bson.D{{
			Key: "$lookup", Value: bson.M{
				"from":         "Events",
				"localField":   "event_participants_event",
				"foreignField": "_id",
				"as":           "Events",
			},
		}},
		bson.D{{
			Key: "$unwind", Value: bson.M{"path": "$Events", "preserveNullAndEmptyArrays": true},
		}},
		bson.D{{
			Key: "$replaceRoot", Value: bson.M{"newRoot": "$Events"},
		}},
		bson.D{
			{Key: "$sort", Value: bson.M{"events_date": -1}},
		},
		bson.D{{Key: "$limit", Value: 1}},
	}
	lastRewildingCursor, _ := config.DB.Collection("EventParticipants").Aggregate(context.TODO(), lastRewildingAgg)
	lastRewildingCursor.All(context.TODO(), &lastRewilding)

	if len(lastRewilding) > 0 {
		response["statistics_last_rewilding"] = lastRewilding[0].EventsDate
	}

	var rewildingByMonth []models.EventStatistics
	rewildingByMonthAgg := mongo.Pipeline{
		bson.D{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "month", Value: bson.D{{Key: "$month", Value: "$events_date"}}},
					{Key: "year", Value: bson.D{{Key: "$year", Value: "$events_date"}}},
				}},
				{Key: "event_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			},
		}},
		bson.D{
			{Key: "$sort", Value: bson.M{"_id": 1}},
		},
	}

	rewildingByMonthCursor, err := config.DB.Collection("Events").Aggregate(context.TODO(), rewildingByMonthAgg)
	rewildingByMonthCursor.All(context.TODO(), &rewildingByMonth)

	if err != nil {
		helpers.ResponseError(c, err.Error())
		return
	} else {
		response["statistics_rewilding_by_month"] = rewildingByMonth
	}

	c.JSON(http.StatusOK, response)
}
