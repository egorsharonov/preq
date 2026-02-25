package mnpevent

import "github.com/google/uuid"

func NewEventID() string {
	return uuid.NewString()
}
