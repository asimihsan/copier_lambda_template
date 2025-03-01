package rotationutil

import "time"

// CalculateNextRotationDate returns the next rotation date based on the given lastDate and frequency.
// In the future, you can enhance this function (or add new ones) to support custom durations.
func CalculateNextRotationDate(lastDate time.Time, frequency string) time.Time {
	switch frequency {
	case "daily":
		return lastDate.AddDate(0, 0, 1)
	case "weekly":
		return lastDate.AddDate(0, 0, 7)
	case "biweekly":
		return lastDate.AddDate(0, 0, 14)
	case "monthly":
		return lastDate.AddDate(0, 1, 0)
	default:
		return lastDate.AddDate(0, 0, 7)
	}
}
