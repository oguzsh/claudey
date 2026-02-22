// Package datetime provides date/time string formatters for session files.
package datetime

import "time"

// DateString returns the current date in YYYY-MM-DD format.
func DateString() string {
	return time.Now().Format("2006-01-02")
}

// TimeString returns the current time in HH:MM format.
func TimeString() string {
	return time.Now().Format("15:04")
}

// DateTimeString returns the current datetime in YYYY-MM-DD HH:MM:SS format.
func DateTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}




