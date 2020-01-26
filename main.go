package main

import (
	"fmt"
	"os"
	"os/exec"
)

func getRepoDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.bristol-events", home), nil
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
}
