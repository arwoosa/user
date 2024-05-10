package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserFollowings struct {
	UserFollowingsId        primitive.ObjectID `bson:"_id,omitempty" json:"user_followings_id"`
	UserFollowingsUser      primitive.ObjectID `bson:"user_followings_user,omitempty" json:"user_followings_user"`
	UserFollowingsFollowing primitive.ObjectID `bson:"user_followings_following,omitempty" json:"user_followings_following"`
	UserFollowingsCreatedAt primitive.DateTime `bson:"user_followings_created_at,omitempty" json:"user_followings_created_at"`
}
