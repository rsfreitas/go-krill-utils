package converters

import (
	"time"

	timestamp "google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertFromTimestampToTime(value *timestamp.Timestamp) time.Time {
	var t time.Time
	if value != nil {
		t = value.AsTime()
	}

	return t
}

func ConvertFromTimestampToTimePointer(value *timestamp.Timestamp) *time.Time {
	if value == nil {
		return nil
	}

	t := ConvertFromTimestampToTime(value)
	return &t
}

// TimeToTimestamp converts a *time.Time to a Protobuf timestamp.
func TimeToTimestamp(t *time.Time) *timestamp.Timestamp {
	if t == nil {
		return nil
	}

	return timestamp.New(*t)
}
