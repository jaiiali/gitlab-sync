package main

import (
	"os"
	"os/exec"
)

func clone(url, dir string) ([]byte, error) {
	cmd := exec.Command("git", "clone", url, dir, "--recursive")

	return cmd.Output()
}

func pull(path string) ([]byte, error) {
	err := os.Chdir(path)
	if err != nil {
		return []byte{}, err
	}

	cmd := exec.Command("git", "pull", "--all")

	return cmd.Output()
}
