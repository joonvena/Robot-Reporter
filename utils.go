package main

import (
	"fmt"
	"regexp"
	"strconv"
)

func getExecutionTime(elapsed string) (float64, error) {
	// Convert string to float64
	f, err := strconv.ParseFloat(elapsed, 64)
	if err != nil {
		return 0.0, err
	}
	return f, nil
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
		return fmt.Sprintf("%.2f", (float64(passed) / (float64(passed) + float64(failed)) * 100))
	}
	return "0"
}
