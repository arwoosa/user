package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type EventTypeGroupStatistics struct {
	EventType     primitive.ObjectID `bson:"_id,omitempty" json:"event_type"`
	EventTypeName string             `bson:"ref_rewilding_types_name,omitempty" json:"event_type_name"`
	EventCount    int                `bson:"ref_rewilding_types_count" json:"event_count"`
}

type EventTypeStatistics struct {
	EventType  primitive.ObjectID `bson:"_id,omitempty" json:"event_type"`
	EventCount int                `bson:"event_count" json:"event_count"`
}

type EventStatistics struct {
	EventPeriod EventStatisticsId `bson:"_id,omitempty" json:"event_period"`
	EventCount  int               `bson:"event_count" json:"event_count"`
}

type EventStatisticsId struct {
	Month int `bson:"month" json:"month"`
	Year  int `bson:"year" json:"year"`
}

type UserStarStatistics struct {
	UserPeriod      EventStatisticsId `bson:"_id,omitempty" json:"user_period"`
	UserCount       int               `bson:"user_count" json:"user_count"`
	UserTotalStar   int               `bson:"total_star" json:"total_star"`
	UserAverageStar float64           `bson:"average_star" json:"average_star"`
}
