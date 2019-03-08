package models

import "time"

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
