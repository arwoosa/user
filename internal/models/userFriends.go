package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserFriends struct {
	UserFriendsId         primitive.ObjectID `bson:"_id,omitempty" json:"user_friends_id"`
	UserFriendsStatus     *int               `bson:"user_friends_status,omitempty" json:"user_friends_status"`
	UserFriendsUser1      primitive.ObjectID `bson:"user_friends_user_1,omitempty" json:"user_friends_user_1"`
	UserFriendsUser2      primitive.ObjectID `bson:"user_friends_user_2,omitempty" json:"user_friends_user_2"`
	UserFriendsCreatedAt  primitive.DateTime `bson:"user_friends_created_at,omitempty" json:"user_friends_created_at"`
	UserFriendsIsOfficial *bool              `bson:"user_friends_is_official,omitempty" json:"user_friends_is_official"`
	UserFriendsDetail     *UsersAggBreathing `bson:"user_friends_detail,omitempty" json:"user_friends_detail,omitempty"`
}
