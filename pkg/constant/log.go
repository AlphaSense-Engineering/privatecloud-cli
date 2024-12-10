// Package constant is the package that contains the constant variables.
package constant

import "time"

const (
	// LogMsgPodCreated is the message that is logged when the pod is created.
	LogMsgPodCreated = "created %s/%s Pod"

	// LogMsgPodDeleted is the message that is logged when the pod is deleted.
	LogMsgPodDeleted = "deleted %s/%s Pod"
)

// LogDefaultTimeFunc is the default time function for the logger.
var LogDefaultTimeFunc = func(t time.Time) time.Time {
	return t.UTC()
}
