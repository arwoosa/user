package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type EventParticipants struct {
	EventParticipantsId              primitive.ObjectID `bson:"_id,omitempty" json:"event_participants_id"`
	EventParticipantsEvent           primitive.ObjectID `bson:"event_participants_event,omitempty" json:"event_participants_event"`
	EventParticipantsUser            primitive.ObjectID `bson:"event_participants_user,omitempty" json:"event_participants_user"`
	EventParticipantsStatus          int64              `bson:"event_participants_status" json:"event_participants_status"`
	EventParticipantsIsPaid          int64              `bson:"event_participants_is_paid,omitempty" json:"event_participants_is_paid"`
	EventParticipantsPaidAmount      float64            `bson:"event_participants_paid_amount,omitempty" json:"event_participants_paid_amount"`
	EventParticipantsPaidAt          string             `bson:"event_participants_paid_at,omitempty" json:"event_participants_paid_at"`
	EventParticipantsPaymentRequest  string             `bson:"event_participants_payment_request,omitempty" json:"event_participants_payment_request"`
	EventParticipantsPaymentResponse string             `bson:"event_participants_payment_response,omitempty" json:"event_participants_payment_response"`
	EventParticipantsExperience      string             `bson:"event_participants_experience,omitempty" json:"event_participants_experience"`
	EventParticipantsCreatedBy       primitive.ObjectID `bson:"event_participants_created_by,omitempty" json:"event_participants_created_by"`
	EventParticipantsCreatedAt       primitive.DateTime `bson:"event_participants_created_at,omitempty" json:"event_participants_created_at"`
	EventParticipantsEventDetail     *Events            `bson:"event_participants_event_detail,omitempty" json:"event_participants_event_detail,omitempty"`
}

type EventParticipantsRanking struct {
	EventParticipantsRankingId       primitive.ObjectID `bson:"_id,omitempty" json:"event_participants_event"`
	EventParticipantsExperienceCount int                `bson:"event_participants_experience_count,omitempty" json:"event_participants_ranking"`
}
