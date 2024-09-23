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

func (t StatisticsRepository) RetrieveRankingFeelings(c *gin.Context) {
	var RewildingRankingFeelings []models.RewildingRanking
	eventParticipantsAgg := t.GetRankingPipeline(c, 1)
	eventParticipantsCursor, eventParticipantsErr := config.DB.Collection("EventParticipants").Aggregate(context.TODO(), eventParticipantsAgg)
	eventParticipantsCursor.All(context.TODO(), &RewildingRankingFeelings)

	if eventParticipantsErr != nil {
		helpers.ResponseError(c, eventParticipantsErr.Error())
		return
	}

	if len(RewildingRankingFeelings) == 0 {
		helpers.ResponseNoData(c, "No Data")
		return
	}
	c.JSON(200, RewildingRankingFeelings)
}

func (t StatisticsRepository) RetrieveRankingRewilding(c *gin.Context) {
	var RewildingRankingFeelings []models.RewildingRanking
	eventParticipantsAgg := t.GetRankingPipeline(c, 2)
	eventParticipantsCursor, eventParticipantsErr := config.DB.Collection("EventParticipants").Aggregate(context.TODO(), eventParticipantsAgg)
	eventParticipantsCursor.All(context.TODO(), &RewildingRankingFeelings)

	if eventParticipantsErr != nil {
		helpers.ResponseError(c, eventParticipantsErr.Error())
		return
	}

	if len(RewildingRankingFeelings) == 0 {
		helpers.ResponseNoData(c, "No Data")
		return
	}
	c.JSON(200, RewildingRankingFeelings)
}

func (t StatisticsRepository) GetRankingPipeline(c *gin.Context, groupType int) mongo.Pipeline {
	// 1: Ranking based on feelings, 2: Ranking based on participant count
	rankingFeelings := c.Query("feelings")
	pipeline := mongo.Pipeline{
		bson.D{{
			Key: "$match", Value: bson.M{
				"event_participants_experience": bson.M{"$exists": true, "$eq": rankingFeelings},
			},
		}},
		bson.D{{
			Key: "$lookup", Value: bson.M{
				"from":         "Events",
				"localField":   "event_participants_event",
				"foreignField": "_id",
				"as":           "event_participants_event_detail",
			},
		}},
		bson.D{{
			Key: "$unwind", Value: "$event_participants_event_detail",
		}},
		bson.D{{
			Key: "$replaceRoot", Value: bson.M{
				"newRoot": "$event_participants_event_detail",
			},
		}},
		bson.D{{
			Key: "$lookup", Value: bson.M{
				"from":         "RefRewildingTypes",
				"localField":   "events_type",
				"foreignField": "_id",
				"as":           "events_type_detail",
			},
		}},
		bson.D{{
			Key: "$unwind", Value: bson.M{"path": "$events_type_detail", "preserveNullAndEmptyArrays": true},
		}},
		bson.D{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$events_rewilding"},
				{Key: "rewilding_participants_experience_count", Value: bson.M{"$sum": 1}},
				{Key: "rewilding_type_list", Value: bson.M{"$addToSet": "$events_type_detail.ref_rewilding_types_name"}},
			},
		}},
		bson.D{{
			Key: "$match", Value: bson.M{
				"rewilding_participants_experience_count": bson.M{"$gte": config.APP_LIMIT.MinimumTopRanking},
			},
		}},
		bson.D{{
			Key: "$lookup", Value: bson.M{
				"from":         "Rewilding",
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "events_rewilding_detail",
			},
		}},
		bson.D{{
			Key: "$unwind", Value: "$events_rewilding_detail",
		}},
		bson.D{{
			Key: "$replaceRoot", Value: bson.M{
				"newRoot": bson.M{
					"$mergeObjects": bson.A{
						"$events_rewilding_detail",
						bson.M{"rewilding_type_list": "$rewilding_type_list"},
						bson.M{"rewilding_participants_experience_count": "$rewilding_participants_experience_count"},
					},
				},
			},
		}},
		bson.D{{Key: "$sort", Value: bson.M{"rewilding_participants_experience_count": -1}}},
		bson.D{{Key: "$limit", Value: 5}},
	}

	if groupType == 2 {
		_, pipeline = pipeline[0], pipeline[1:]
	}

	return pipeline
}
