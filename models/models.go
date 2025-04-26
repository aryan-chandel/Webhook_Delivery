package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscriber struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	SubscriptionID string             `bson:"sub_id" json:"sub_id"`
	TargetURL      string             `bson:"target_url" json:"target_url"`
	Secret         string             `bson:"secret,omitempty" json:"secret,omitempty"`
	EventTypes     []string            `bson:"event_types,omitempty" json:"event_types,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type DeliveryLog struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	WebhookID      string             `bson:"webhook_id" json:"webhook_id"`             // ID of the original webhook
	SubscriptionID string             `bson:"subscription_id" json:"subscription_id"`   // ID of the subscription
	TargetURL      string             `bson:"target_url" json:"target_url"`             // Redundant copy for easier querying
	Attempted_at   time.Time          `bson:"attempted_at" json:"attempted_at"`         // When this attempt was made
	AttemptNumber  int                `bson:"attempt_number" json:"attempt_number"`     // 1 = initial, 2+ = retries
	Outcome        string             `bson:"outcome" json:"outcome"`                   // Success, Failed Attempt, Failure
	HTTPStatus     int                `bson:"http_status,omitempty" json:"http_status"` // HTTP status from target
}

type DeliveryTask struct {
	LogID          primitive.ObjectID `json:"logid"` //_id of delivery collection for updating log
	WebhookID      string             `json:"webhook_id"`
	SubscriptionID string             `json:"subscription_id"`
	TargetURL      string             `json:"target_url"`
	Payload        []byte             `json:"payload"`
	AttemptNumber  int                `json:"attempt_number"`
}
