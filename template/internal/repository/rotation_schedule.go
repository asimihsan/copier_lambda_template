package repository

import (
	"context"
	"time"
)

type RotationScheduleEvent struct {
	ScheduledHour  string    `dynamodbav:"scheduled_hour"` // partition key
	EventID        string    `dynamodbav:"event_id"`       // sort key, (rotation_id#action#timestamp)
	RotationID     string    `dynamodbav:"rotation_id"`
	Action         string    `dynamodbav:"action"` // "reminder" or "handover"
	SlackChannelID string    `dynamodbav:"slack_channel_id"`
	NextOwner      string    `dynamodbav:"next_owner"`
	EventTime      time.Time `dynamodbav:"event_time"`
	ExpiresAt      int64     `dynamodbav:"expires_at"`
}

// RotationScheduleRepository defines operations on rotation schedule events
type RotationScheduleRepository interface {
	// AddEvent adds a new event to the rotation schedule
	AddEvent(ctx context.Context, event RotationScheduleEvent) error
	
	// GetEventsByHour retrieves all events scheduled for a specific hour
	GetEventsByHour(ctx context.Context, hour string) ([]RotationScheduleEvent, error)
}
