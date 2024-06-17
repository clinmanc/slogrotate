package slogrotate

import (
	"encoding/json"
	"fmt"
	"time"
)

type Frequency int

const (
	FrequencyNone Frequency = iota
	FrequencyYearly
	FrequencyMonthly
	FrequencyWeekly
	FrequencyDaily
	FrequencyHourly
	FrequencyMinutely
)

var frequencyToString = map[Frequency]string{
	FrequencyNone:     "none",
	FrequencyYearly:   "yearly",
	FrequencyMonthly:  "monthly",
	FrequencyWeekly:   "weekly",
	FrequencyDaily:    "daily",
	FrequencyHourly:   "hourly",
	FrequencyMinutely: "minutely",
}

var stringToFrequency = map[string]Frequency{
	"none":     FrequencyNone,
	"yearly":   FrequencyYearly,
	"monthly":  FrequencyMonthly,
	"weekly":   FrequencyWeekly,
	"daily":    FrequencyDaily,
	"hourly":   FrequencyHourly,
	"minutely": FrequencyMinutely,
}

func (f *Frequency) MarshalJSON() ([]byte, error) {
	str, ok := frequencyToString[*f]
	if !ok {
		return nil, fmt.Errorf("invalid frequency: %d", f)
	}
	return json.Marshal(str)
}

func (f *Frequency) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	freq, ok := stringToFrequency[str]
	if !ok {
		return fmt.Errorf("invalid frequency: %s", str)
	}
	*f = freq
	return nil
}

func (f *Frequency) IsSame(lastTime time.Time, currentTime time.Time) bool {
	freq := *f

	if freq >= FrequencyYearly && lastTime.Year() != currentTime.Year() {
		return false
	}
	if freq >= FrequencyMonthly && lastTime.Month() != currentTime.Month() {
		return false
	}
	if freq >= FrequencyWeekly {
		_, lastWeek := lastTime.ISOWeek()
		_, currentWeek := currentTime.ISOWeek()
		if lastWeek != currentWeek {
			return false
		}
	}
	if freq >= FrequencyDaily && lastTime.Day() != currentTime.Day() {
		return false
	}
	if freq >= FrequencyHourly && lastTime.Hour() != currentTime.Hour() {
		return false
	}
	if freq >= FrequencyMinutely && lastTime.Minute() != currentTime.Minute() {
		return false
	}

	return true
}
