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

func getEvents() ([]Event, error) {
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

	now := time.Now()
	inOneWeek := now.AddDate(0, 0, 7)

	var events []Event
	var event Event
	for _, file := range files {
		jsonFile, err := os.Open(file)
		if err != nil {
			fmt.Println(err)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		json.Unmarshal(byteValue, &event)

		event.time = time.Unix(event.StartTime, 0)

		if event.time.Before(inOneWeek) && event.time.After(now) {
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

func main() {
	exists, err := repoExists()
	if err != nil {
		panic(err)
	}

	if !exists {
		cloneRepo()
	}

	events, err := getEvents()
	if err != nil {
		panic(err)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].time.Before(events[j].time)
	})

	for _, event := range events {
		fmt.Printf("%s ", event.time.Weekday().String())
		fmt.Println(event.Title)
	}
}
