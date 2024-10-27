package sessions

import "time"

func getCurrentAndNextYearTime() (string, string) {
	currentTime := time.Now().UTC()
	nextYearTime := currentTime.AddDate(1, 0, 0)
	currentTimeFormatted := currentTime.Format("2006-01-02T15:04:05.999999999Z")
	nextYearTimeFormatted := nextYearTime.Format("2006-01-02T15:04:05.999999999Z")
	return currentTimeFormatted, nextYearTimeFormatted
}
