package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// time="2019-07-05 08:52:58.996"
const timeRegex = `time="(?P<time>\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}.\d{3})"`
const timeLayout = "2006-01-02 15:04:05.999"

func main() {
	timePattern, err := regexp.Compile(timeRegex)
	if err != nil {
		log.Fatalf("could compile regex %s: %v", timeRegex, err)
		return
	}
	filePattern := flag.String("file", "testdata/*.log", "file glob to match")
	if filePattern == nil {
		flag.PrintDefaults()
		return
	}

	filenames, err := filepath.Glob(*filePattern)
	if err != nil {
		log.Fatalf("could not parse glob: %v", err)
		return
	}

	var lines int
	days := map[string]int{}
	hours := map[string]int{}
	for _, filename := range filenames {
		file, openErr := os.Open(filename)
		if openErr != nil {
			log.Fatalf("could open file %q: %v", filename, openErr)
			return
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			lines++
			matches := timePattern.FindStringSubmatch(line)
			if len(matches) == 0 {
				continue
			}
			match := matches[len(matches)-1]
			timestamp, parseErr := time.Parse(timeLayout, match)
			if parseErr != nil {
				continue
			}
			year, month, day := timestamp.Date()
			hour := timestamp.Hour()
			dayStr := fmt.Sprintf("%d %s %d", year, month, day)
			days[dayStr] = days[dayStr] + 1
			hourStr := fmt.Sprintf("%d %s %d %d", year, month, day, hour)
			hours[hourStr] = hours[hourStr] + 1
		}
	}
	log.Printf("number of lines: %d", lines)
	log.Printf("number of days: %d", len(days))
	var maxHour string
	var maxHourCount int
	for hour, count := range hours {
		if count > maxHourCount {
			maxHour = hour
			maxHourCount = count
		}
	}
	log.Printf("most popular hour: %s %d", maxHour, maxHourCount)
}
