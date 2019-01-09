package utils

import (
	"fmt"
	"time"
)

func PrintStdOut(format string, args ...interface{}) {
	logTime := time.Now().Local().Format("2006-01-02 15:04:05")

	fmt.Printf(fmt.Sprintf("%s %s\n", logTime, format), args...)
}
