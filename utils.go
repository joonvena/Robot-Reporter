package main

import (
	"fmt"
	"regexp"
	"time"
)

const (
	timeFormatLayout = "20060102 15:04:05,000"
)

func getExecutionTime(start string, end string) (float64, error) {
	startTime, err := time.Parse(timeFormatLayout, start)
	if err != nil {
		return 0.0, err
	}
	endTime, err := time.Parse(timeFormatLayout, end)
	if err != nil {
		return 0.0, err
	}

	diff := endTime.Sub(startTime)
	return diff.Seconds(), nil
}

func resultCommentExists(comment string) bool {
	match, _ := regexp.MatchString("^### Robot Results", comment)
	return match
}

func passPercentage(passed int, failed int) string {
	if passed != 0 && failed == 0 {
		return "100"
	}
	if passed != 0 && failed != 0 {
		return fmt.Sprintf("%.2f", (float64(passed) / float64(failed)))
	}
	return "0"
}
