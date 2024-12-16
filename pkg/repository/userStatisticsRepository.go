package repository

import (
	"context"
	"math"
	"net/http"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserStatisticsRepository struct{}

func (t UserStatisticsRepository) Retrieve(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)

	response := gin.H{
		"statistics_last_rewilding":     "",
		"statistics_rewilding_by_month": []map[string]any{},
		"stars_last_achieved_month":     0,                  // [x] If the user received stars within the last month
		"stars_months_last_achieved":    0,                  // [x] If the user has not received stars in the last month
		"oosa_star_per_user":            []map[string]any{}, // [x] 2.1 OOSA Platform Average
		"oosa_star_current_user":        []map[string]any{}, // [x] 2.2 User’s Monthly Star Count
		"user_star_per_year":            0,                  // [x] 3.1 User’s Average Star Count per Year: Total stars received by the user to date divided by the number of years they have been using the app.
		"user_star_by_type":             []map[string]any{}, // [x] 3.2 Rewilding-Type Statistics: Cumulative star count for the user under each rewilding type (retrieved from /references via rewilding_types).
	}

	var EventParticipants models.EventParticipants

	filterAchievementBase := bson.D{
		{Key: "event_participants_user", Value: userDetail.UsersId},
		{Key: "event_participants_status", Value: 1},
		{Key: "event_participants_star_type", Value: bson.M{"$exists": true}},
	}

	filterAchievementUser := append(filterAchievementBase, primitive.E{Key: "event_participants_achievement_unlocked_at", Value: bson.M{"$exists": true}})
	opts := options.FindOne().SetSort(bson.D{{Key: "event_participants_achievement_unlocked_at", Value: -1}})
	checkLastStar := config.DB.Collection("EventParticipants").
		FindOne(context.TODO(), filterAchievementUser, opts).
		Decode(&EventParticipants)

	if checkLastStar != mongo.ErrNoDocuments {
		/*lastStarDiff := time.Since(EventParticipants.EventParticipantsAchievementUnlockedAt.Time())
		response["stars_months_last_achieved"] = int64(lastStarDiff.Hours() / 24 / 30)

		monthFrom := time.Now().Add(-24 * time.Hour * 30)

		optsStarWithinLastMonth := options.Count().SetHint("_id_")
		filterStarWithinLastMonth := append(filterAchievementBase, primitive.E{Key: "event_participants_achievement_unlocked_at", Value: bson.M{"$exists": true, "$gte": primitive.NewDateTimeFromTime(monthFrom)}})
		countStarWithinLastMonth, _ := config.DB.Collection("EventParticipants").CountDocuments(context.TODO(), filterStarWithinLastMonth, optsStarWithinLastMonth)

		response["stars_within_last_month"] = countStarWithinLastMonth*/

		response["stars_last_achieved_month"] = EventParticipants.EventParticipantsAchievementUnlockedAt.Time().Month()
	}

	userMemberLength := time.Since(userDetail.UsersCreatedAt.Time())
	memberYears := int64(math.Ceil(userMemberLength.Hours() / 24 / 365))
	optsStarUser := options.Count().SetHint("_id_")
	countStarUser, _ := config.DB.Collection("EventParticipants").CountDocuments(context.TODO(), filterAchievementBase, optsStarUser)
	response["user_star_per_year"] = countStarUser / memberYears

	var userByMonth []models.UserStarStatistics
	userByMonthAgg := mongo.Pipeline{
		bson.D{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "month", Value: bson.D{{Key: "$month", Value: "$users_created_at"}}},
					{Key: "year", Value: bson.D{{Key: "$year", Value: "$users_created_at"}}},
				}},
				{Key: "user_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			},
		}},
		bson.D{
			{Key: "$sort", Value: bson.M{"_id": 1}},
		},
	}
	userByMonthCursor, _ := config.DB.Collection("Users").Aggregate(context.TODO(), userByMonthAgg)
	userByMonthCursor.All(context.TODO(), &userByMonth)

	monthIndex := map[int]time.Month{
		1:  time.January,
		2:  time.February,
		3:  time.March,
		4:  time.April,
		5:  time.May,
		6:  time.June,
		7:  time.July,
		8:  time.August,
		9:  time.September,
		10: time.October,
		11: time.November,
		12: time.December,
	}

	monthStart := monthIndex[userByMonth[0].UserPeriod.Month]
	firstMonth := time.Date(userByMonth[0].UserPeriod.Year, monthStart, 1, 0, 0, 0, 0, time.UTC)

	monthCount := int(time.Since(firstMonth).Hours() / 24 / 30)

	var monthlyStarCount []models.EventStatistics
	monthlyStarCountAgg := mongo.Pipeline{
		bson.D{{
			Key: "$match", Value: bson.M{
				"event_participants_status":    1,
				"event_participants_star_type": bson.M{"$exists": true},
			},
		}},
		bson.D{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "month", Value: bson.D{{Key: "$month", Value: "$event_participants_achievement_unlocked_at"}}},
					{Key: "year", Value: bson.D{{Key: "$year", Value: "$event_participants_achievement_unlocked_at"}}},
				}},
				{Key: "event_count", Value: bson.D{{Key: "$count", Value: bson.D{}}}},
			},
		}},
	}
	monthlyStarCountCursor, _ := config.DB.Collection("EventParticipants").Aggregate(context.TODO(), monthlyStarCountAgg)
	monthlyStarCountCursor.All(context.TODO(), &monthlyStarCount)

	var monthlyStarCurrentUserCount []models.EventStatistics
	monthlyStarCurrentUserAgg := mongo.Pipeline{
		bson.D{{
			Key: "$match", Value: bson.M{
				"event_participants_user":      userDetail.UsersId,
				"event_participants_status":    1,
				"event_participants_star_type": bson.M{"$exists": true},
			},
		}},
		bson.D{{
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "month", Value: bson.D{{Key: "$month", Value: "$event_participants_achievement_unlocked_at"}}},
					{Key: "year", Value: bson.D{{Key: "$year", Value: "$event_participants_achievement_unlocked_at"}}},
				}},
				{Key: "event_count", Value: bson.D{{Key: "$count", Value: bson.D{}}}},
			},
		}},
	}
	monthlyStarCurrentUserCursor, _ := config.DB.Collection("EventParticipants").Aggregate(context.TODO(), monthlyStarCurrentUserAgg)
	monthlyStarCurrentUserCursor.All(context.TODO(), &monthlyStarCurrentUserCount)

	var monthlyRewildingGroupCurrentUserCount []models.EventTypeGroupStatistics
	var monthlyRewildingCurrentUserCount []models.EventTypeStatistics
	monthlyRewildingCurrentUserAgg := mongo.Pipeline{
		bson.D{{
			Key: "$match", Value: bson.M{
				"event_participants_user":      userDetail.UsersId,
				"event_participants_status":    1,
				"event_participants_star_type": bson.M{"$exists": true},
			},
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
			Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$Events.events_type"},
				{Key: "event_count", Value: bson.D{{Key: "$count", Value: bson.D{}}}},
			},
		}},
	}
	monthlyRewildingCurrentUserCursor, _ := config.DB.Collection("EventParticipants").Aggregate(context.TODO(), monthlyRewildingCurrentUserAgg)
	monthlyRewildingCurrentUserCursor.All(context.TODO(), &monthlyRewildingCurrentUserCount)

	var RefRewildingTypes []models.RefRewildingTypes
	cursor, err := config.DB.Collection("RefRewildingTypes").Find(context.TODO(), bson.D{})
	cursor.All(context.TODO(), &RefRewildingTypes)

	for _, v := range RefRewildingTypes {
		count := 0
		for _, v1 := range monthlyRewildingCurrentUserCount {
			if v1.EventType == v.RefRewildingTypesId {
				count = v1.EventCount
			}
		}
		monthlyRewildingGroupCurrentUserCount = append(monthlyRewildingGroupCurrentUserCount, models.EventTypeGroupStatistics{
			EventType:     v.RefRewildingTypesId,
			EventTypeName: v.RefRewildingTypesName,
			EventCount:    count,
		})
	}

	var timeIndex []string
	var userStatisticsByMonth []models.UserStarStatistics
	var userStatisticsCurrentUserByMonth []models.EventStatistics
	userCountIdx := map[string]int{}
	userCountAccumulated := map[string]int{}
	monthlyStarCountIdx := map[string]int{}
	monthlyStarCurrentUserCountIdx := map[string]int{}

	for _, v := range userByMonth {
		idx := strconv.Itoa(v.UserPeriod.Year) + helpers.PadLeft(strconv.Itoa(v.UserPeriod.Month), 2, "0")
		userCountIdx[idx] = v.UserCount
	}

	for _, v := range monthlyStarCount {
		idx := strconv.Itoa(v.EventPeriod.Year) + helpers.PadLeft(strconv.Itoa(v.EventPeriod.Month), 2, "0")
		monthlyStarCountIdx[idx] = v.EventCount
	}

	for _, v := range monthlyStarCurrentUserCount {
		idx := strconv.Itoa(v.EventPeriod.Year) + helpers.PadLeft(strconv.Itoa(v.EventPeriod.Month), 2, "0")
		monthlyStarCurrentUserCountIdx[idx] = v.EventCount
	}

	accumulated := 0
	for i := 0; i <= monthCount; i++ {
		addedMonth := firstMonth.AddDate(0, i, 0)
		timeIdx := addedMonth.Format("200601")

		addedValue := userCountIdx[timeIdx]
		starValue := monthlyStarCountIdx[timeIdx]

		if i > 0 {
			prevTimeIndex := timeIndex[i-1]
			accumulated = userCountAccumulated[prevTimeIndex]
		}

		userCountAccumulated[timeIdx] = accumulated + addedValue
		userStatisticsByMonth = append(userStatisticsByMonth, models.UserStarStatistics{
			UserPeriod: models.EventStatisticsId{
				Month: int(addedMonth.Month()),
				Year:  int(addedMonth.Year()),
			},
			UserCount:       accumulated + addedValue,
			UserTotalStar:   starValue,
			UserAverageStar: float64(starValue) / float64(accumulated+addedValue),
		})

		userStatisticsCurrentUserByMonth = append(userStatisticsCurrentUserByMonth, models.EventStatistics{
			EventPeriod: models.EventStatisticsId{
				Month: int(addedMonth.Month()),
				Year:  int(addedMonth.Year()),
			},
			EventCount: monthlyStarCurrentUserCountIdx[timeIdx],
		})

		timeIndex = append(timeIndex, timeIdx)
	}

	response["oosa_star_per_user"] = userStatisticsByMonth
	response["oosa_star_current_user"] = userStatisticsCurrentUserByMonth
	response["user_star_by_type"] = monthlyRewildingGroupCurrentUserCount

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
