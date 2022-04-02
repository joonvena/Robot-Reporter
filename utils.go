package main

import (
	"regexp"
	"strings"
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

func getSuiteName(i int, output Output) string {
	for k := 0; k < len(output.Statistics.Suite.Stat); k++ {
		splittedID := strings.SplitN(output.Suite.Suite.Test[i].ID, "-", 3)
		suiteID := splittedID[0] + "-" + splittedID[1]
		testSuiteID := output.Statistics.Suite.Stat[k].ID
		if testSuiteID == suiteID {
			return output.Statistics.Suite.Stat[k].Name
		}
	}
	return ""
}
