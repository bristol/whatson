package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Event struct {
	Description string `json:"description"`
	EndTime     int    `json:"end_time"`
	Link        string `json:"link"`
	Location    struct {
		Address   string  `json:"address"`
		City      string  `json:"city"`
		Country   string  `json:"country"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Name      string  `json:"name"`
	} `json:"location"`
	Org       string `json:"org"`
	StartTime int64  `json:"start_time"`
	Title     string `json:"title"`
	time      time.Time
}

func getRepoDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.bristol-events", home), nil
}

func readJson(file string, event *Event) error {
	jsonFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(byteValue, event)
	return err
}

func getEvents(after, before time.Time) ([]Event, error) {
	var files []string
	dir, err := getRepoDir()
	if err != nil {
		return []Event{}, err
	}

	root := fmt.Sprintf("%s/events", dir)
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			if strings.HasSuffix(path, ".json") {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	var events []Event
	var event Event
	for _, file := range files {
		readJson(file, &event)

		event.time = time.Unix(event.StartTime, 0)

		if event.time.Before(before) && event.time.After(after) {
			events = append(events, event)
		}
	}

	return events, nil
}

func repoExists() (bool, error) {
	dir, err := getRepoDir()
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func cloneRepo() error {
	dir, err := getRepoDir()
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "clone", "https://github.com/bristol/events.git", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func updateRepo() error {
	dir, err := getRepoDir()
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "pull", "origin", "master")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	command := "week"
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "update":
			command = "update"
		case "today":
			command = "today"
		default:
			panic("unknown command")
		}
	}

	exists, err := repoExists()
	if err != nil {
		panic(err)
	}

	if !exists {
		cloneRepo()
	}

	if command == "update" {
		updateRepo()
		return
	}

	after := time.Now()
	var before time.Time
	switch command {
	case "week":
		before = after.AddDate(0, 0, 7)
	case "today":
		before = after.AddDate(0, 0, 1)
	}

	events, err := getEvents(after, before)
	if err != nil {
		panic(err)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].time.Before(events[j].time)
	})

	lastDay := ""
	for _, event := range events {
		nextDay := event.time.Weekday().String()
		if nextDay != lastDay {
			lastDay = nextDay
			fmt.Println(nextDay)
		}
		fmt.Printf("  %s [%s] %s (at %s)\n", event.time.Format("3:04PM"), event.Org, event.Title, event.Location.Name)
	}
}
