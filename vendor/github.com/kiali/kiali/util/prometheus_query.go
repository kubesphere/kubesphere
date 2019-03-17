package util

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

func AdjustRateInterval(namespaceCreationTime time.Time, queryTime time.Time, ratesInterval string) (string, error) {
	startTime, err := GetStartTimeForRateInterval(queryTime, ratesInterval)
	if err != nil {
		return "", err
	}

	if startTime.Before(namespaceCreationTime) {
		interval := queryTime.Sub(namespaceCreationTime)
		ratesInterval = fmt.Sprintf("%vs", int(interval.Seconds()))
	}

	return ratesInterval, nil
}

func GetStartTimeForRateInterval(baseTime time.Time, rateInterval string) (time.Time, error) {
	duration, err := model.ParseDuration(rateInterval)
	if err != nil {
		return time.Time{}, err
	}

	return baseTime.Add(-time.Duration(duration)), nil
}
