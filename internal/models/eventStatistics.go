package models

type EventStatistics struct {
	EventPeriod EventStatisticsId `bson:"_id,omitempty" json:"event_period"`
	EventCount  int               `bson:"event_count" json:"event_count"`
}

type EventStatisticsId struct {
	Month int `bson:"month" json:"month"`
	Year  int `bson:"year" json:"year"`
}
