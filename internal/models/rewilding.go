package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Rewilding struct {
	RewildingID       primitive.ObjectID   `bson:"_id,omitempty" json:"rewilding_id"`
	RewildingType     string               `bson:"rewilding_type,omitempty" json:"rewilding_type"`
	RewildingTypeData RefRewildingTypes    `bson:"rewilding_type_data,omitempty" json:"rewilding_type_data"`
	RewildingCity     string               `bson:"rewilding_city,omitempty" json:"rewilding_city"`
	RewildingArea     string               `bson:"rewilding_area,omitempty" json:"rewilding_area"`
	RewildingName     string               `bson:"rewilding_name,omitempty" json:"rewilding_name"`
	RewildingRating   int                  `bson:"rewilding_rating,omitempty" json:"rewilding_rating"`
	RewildingLat      primitive.Decimal128 `bson:"rewilding_lat,omitempty" json:"rewilding_lat"`
	RewildingLng      primitive.Decimal128 `bson:"rewilding_lng,omitempty" json:"rewilding_lng"`
}
