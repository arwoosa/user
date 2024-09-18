package repository

import (
	"context"
	"net/http"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StatisticsRepository struct{}

func (t StatisticsRepository) Retrieve(c *gin.Context) {
	y, m, _ := time.Now().Date()
	periodStart, periodEnd := helpers.MonthInterval(y, m)

	response := gin.H{
		"world_rewilding_members":  0,
		"world_rewilding_polaroid": 0,
		"world_rewilding_country":  []map[string]any{},
	}

	// 1
	var eventCountryStatisticCount []models.EventsStatisticCount
	eventCountryStatisticsAgg := mongo.Pipeline{
		bson.D{{
			Key: "$match", Value: bson.M{
				"events_date": bson.M{
					"$gte": primitive.NewDateTimeFromTime(periodStart),
					"$lte": primitive.NewDateTimeFromTime(periodEnd),
				},
			},
		}},
		bson.D{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$events_country_code"},
				{Key: "events_country_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			},
		}},
	}

	eventCountryStatisticsCursor, eventCountryStatisticsErr := config.DB.Collection("Events").Aggregate(context.TODO(), eventCountryStatisticsAgg)
	eventCountryStatisticsCursor.All(context.TODO(), &eventCountryStatisticCount)

	if eventCountryStatisticsErr != nil {
		helpers.ResponseError(c, eventCountryStatisticsErr.Error())
		return
	} else {
		response["world_rewilding_country"] = eventCountryStatisticCount
	}

	// 2
	var eventUserParticipant []models.EventParticipants
	eventUserParticipantAgg := mongo.Pipeline{
		bson.D{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$event_participants_user"},
			},
		}},
	}
	eventUserParticipantCursor, eventUserParticipantErr := config.DB.Collection("EventParticipants").Aggregate(context.TODO(), eventUserParticipantAgg)
	eventUserParticipantCursor.All(context.TODO(), &eventUserParticipant)

	if eventUserParticipantErr != nil {
		helpers.ResponseError(c, eventUserParticipantErr.Error())
		return
	} else {
		response["world_rewilding_members"] = len(eventUserParticipant)
	}

	// 3
	// var eventWithPolaroids []models.Events
	var eventWithPolaroids []models.EventPolaroids
	eventWithPolaroidsAgg := mongo.Pipeline{
		bson.D{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$event_polaroids_event"},
			},
		}},
	}
	eventWithPolaroidsCursor, eventWithPolaroidsErr := config.DB.Collection("EventPolaroids").Aggregate(context.TODO(), eventWithPolaroidsAgg)
	eventWithPolaroidsCursor.All(context.TODO(), &eventWithPolaroids)

	if eventWithPolaroidsErr != nil {
		helpers.ResponseError(c, eventWithPolaroidsErr.Error())
		return
	} else {
		response["world_rewilding_polaroid"] = len(eventWithPolaroids)
	}

	c.JSON(http.StatusOK, response)
}
